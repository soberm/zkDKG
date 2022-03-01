package dkg

import (
	"bytes"
	"client/internal/pkg/group/curve25519"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/share"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
)

type DistKeyShare struct {
	Commits     []kyber.Point
	Share       *share.PriShare
	PrivatePoly []kyber.Scalar
}

type Participant struct {
	index int
	pub   kyber.Point
}

type DistKeyGenerator struct {
	suite              suites.Suite
	client             *ethclient.Client
	chainID            *big.Int
	contract           *ZKDKGContract
	ethereumPrivateKey *ecdsa.PrivateKey
	long               kyber.Scalar
	pub                kyber.Point
	participants       map[int]*Participant
	index              *big.Int
	priPoly            *share.PriPoly
	shares             map[int]kyber.Scalar
	commitments        map[int][]kyber.Point
}

func NewDistributedKeyGenerator(config *Config) (*DistKeyGenerator, error) {

	curve := &curve25519.ProjectiveCurve{}
	curve.Init(ParamBabyJubJub(), false)
	suite := &curve25519.SuiteCurve25519{ProjectiveCurve: *curve}

	client, err := ethclient.Dial(config.EthereumNode)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("chainID: %v", err)
	}

	contract, err := NewZKDKGContract(common.HexToAddress(config.ContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("zkDKG contract: %v", err)
	}

	ethereumPrivateKey, err := crypto.HexToECDSA(config.EthereumPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("hex to ecdsa: %v", err)
	}

	long, err := HexToScalar(suite, config.DkgPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("hex to scalar: %v", err)
	}

	return &DistKeyGenerator{
		suite:              suite,
		client:             client,
		chainID:            chainID,
		contract:           contract,
		ethereumPrivateKey: ethereumPrivateKey,
		long:               long,
		pub:                suite.Point().Mul(long, nil),
		participants:       make(map[int]*Participant),
		shares:             make(map[int]kyber.Scalar),
		commitments:        make(map[int][]kyber.Point),
	}, nil

}

func (d *DistKeyGenerator) Generate(ctx context.Context) error {
	log.Info("Generating distributed private key...")
	sink := make(chan *ZKDKGContractRegistrationEndLog)
	defer close(sink)

	sub, err := d.contract.WatchRegistrationEndLog(
		&bind.WatchOpts{
			Context: ctx,
		},
		sink,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	go func() {
		if err := d.WatchBroadcastSharesLog(ctx); err != nil {
			log.Errorf("Watching broadcast shares log failed: %v", err)
		}
	}()

	if err := d.Register(); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	log.Info("Waiting until registration is finished...")
	<-sink

	log.Info("Retrieving all participants for this run...")
	if err := d.Participants(); err != nil {
		return fmt.Errorf("participants: %w", err)
	}

	log.Info("Broadcasting commitments and deals...")
	if err := d.Deals(); err != nil {
		return fmt.Errorf("deals: %w", err)
	}

	time.Sleep(10 * time.Second)

	return nil
}

func (d *DistKeyGenerator) Register() error {

	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	pub, err := PointToBig(d.pub)
	if err != nil {
		return fmt.Errorf("pub to big: %w", err)
	}

	tx, err := d.contract.Register(opts, pub)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		return fmt.Errorf("wait mined register: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}

	d.index, err = d.contract.Participants(nil, crypto.PubkeyToAddress(d.ethereumPrivateKey.PublicKey))
	if err != nil {
		return fmt.Errorf("participants: %w", err)
	}

	log.Infof("Registered as participant with index %v", d.index)
	return nil
}

func (d *DistKeyGenerator) WatchBroadcastSharesLog(ctx context.Context) error {
	sink := make(chan *ZKDKGContractBroadcastSharesLog)
	defer close(sink)

	sub, err := d.contract.WatchBroadcastSharesLog(
		&bind.WatchOpts{
			Context: ctx,
		},
		sink,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case event := <-sink:
			log.Infof("Handling broadcast shares log...")
			if err := d.HandleBroadcastSharesLog(event); err != nil {
				log.Errorf("Handling broadcast shares log failed: %v", err)
			}
		case err = <-sub.Err():
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *DistKeyGenerator) HandleBroadcastSharesLog(broadcastSharesLog *ZKDKGContractBroadcastSharesLog) error {

	accountAddress := crypto.PubkeyToAddress(d.ethereumPrivateKey.PublicKey)
	if accountAddress == broadcastSharesLog.Sender {
		log.Infof("Ignored own broadcast")
		return nil
	}

	tx, _, err := d.client.TransactionByHash(context.Background(), broadcastSharesLog.Raw.TxHash)
	if err != nil {
		return fmt.Errorf("transaction by hash: %w", err)
	}

	txData := tx.Data()
	a, err := abi.JSON(strings.NewReader(ZKDKGContractABI))
	if err != nil {
		return fmt.Errorf("abi from json: %w", err)
	}

	method, err := a.MethodById(txData[:4])
	if err != nil {
		return fmt.Errorf("method by id: %w", err)
	}

	inputs, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return fmt.Errorf("unpack inputs: %w", err)
	}

	commitments := inputs[0].([][2]*big.Int)
	shares := inputs[1].([]*big.Int)

	commits, err := BigToPoints(d.suite, commitments)
	if err != nil {
		return fmt.Errorf("big to points: %w", err)
	}

	i := int(d.index.Int64())
	j := i
	if i > int(broadcastSharesLog.Index.Int64()) {
		j -= 1
	}

	fi := d.suite.Scalar().SetBytes(shares[j].Bytes())

	sharedKey, err := d.PreSharedKey(i, d.long, d.participants[int(broadcastSharesLog.Index.Int64())].pub)
	if err != nil {
		return fmt.Errorf("pre shared key: %w", err)
	}
	fi.Sub(fi, sharedKey)

	fig := d.suite.Point().Base().Mul(fi, nil)
	pubPoly := share.NewPubPoly(d.suite, nil, commits)

	sh := pubPoly.Eval(i)

	if !fig.Equal(sh.V) {
		log.Infof("Received invalid share from dealer %v", broadcastSharesLog.Index.Int64())
		return nil
	}
	log.Infof("Received valid share from dealer %v", broadcastSharesLog.Index.Int64())
	d.shares[i] = fi
	d.commitments[i] = commits

	return nil
}

func (d *DistKeyGenerator) DistKeyShare() (*DistKeyShare, error) {
	sh := d.suite.Scalar().Zero()
	var pub *share.PubPoly
	var err error
	for i, commitments := range d.commitments {
		sh = sh.Add(sh, d.shares[i])
		pubPoly := share.NewPubPoly(d.suite, nil, commitments)
		if pub == nil {
			pub = pubPoly
			continue
		}
		pub, err = pub.Add(pubPoly)
		if err != nil {
			return nil, fmt.Errorf("add: %w", err)
		}
	}
	_, commits := pub.Info()
	return &DistKeyShare{
		Commits: commits,
		Share: &share.PriShare{
			I: int(d.index.Int64()),
			V: sh,
		},
		PrivatePoly: d.priPoly.Coefficients(),
	}, nil
}

func (d *DistKeyGenerator) Participants() error {

	count, err := d.contract.CountParticipants(nil)
	if err != nil {
		return fmt.Errorf("count participants: %w", err)
	}

	for i := 0; i < int(count.Uint64()); i++ {
		participant, err := d.contract.FindParticipantByIndex(nil, big.NewInt(int64(i)))
		if err != nil {
			return fmt.Errorf("find participants by index: %w", err)
		}

		log.Printf("Adding participant %+v", participant)

		pub, err := BigToPoint(d.suite, participant.PublicKey)
		if err != nil {
			return fmt.Errorf("big to point: %w", err)
		}

		d.participants[i] = &Participant{index: i, pub: pub}
	}

	return nil

}

func (d *DistKeyGenerator) Deals() error {
	threshold, err := d.contract.Threshold(nil)
	if err != nil {
		return fmt.Errorf("threshold: %w", err)
	}

	secret := d.suite.Scalar().Pick(d.suite.RandomStream())
	d.priPoly = share.NewPriPoly(d.suite, int(threshold.Int64()), secret, d.suite.RandomStream())
	pubPoly := d.priPoly.Commit(nil)

	_, commits := pubPoly.Info()
	commitments, err := PointsToBig(commits)
	if err != nil {
		return fmt.Errorf("points to big: %w", err)
	}
	log.Infof("Commitments: %v", commitments)

	deals := make([]*big.Int, 0)
	for i := 0; i < len(d.participants); i++ {
		if i == int(d.index.Int64()) {
			continue
		}

		participant := d.participants[i]

		priShare, err := d.EncryptedPrivateShare(participant.index)
		if err != nil {
			return fmt.Errorf("encrypted private share: %w", err)
		}

		b, err := priShare.V.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal binary deal: %w", err)
		}
		deals = append(deals, new(big.Int).SetBytes(b))
	}

	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	tx, err := d.contract.BroadcastShares(opts, commitments, deals)
	if err != nil {
		return fmt.Errorf("broadcast shares: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		return fmt.Errorf("wait mined register: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}

	return nil
}

func (d *DistKeyGenerator) EncryptedPrivateShare(i int) (*share.PriShare, error) {
	priShare := d.priPoly.Eval(i)

	sharedKey, err := d.PreSharedKey(i, d.long, d.participants[i].pub)
	if err != nil {
		return nil, fmt.Errorf("pre shared key: %w", err)
	}
	priShare.V.Add(priShare.V, sharedKey)

	return priShare, nil
}

func (d *DistKeyGenerator) PreSharedKey(i int, privateKey kyber.Scalar, publicKey kyber.Point) (kyber.Scalar, error) {
	pre := dhExchange(d.suite, privateKey, publicKey)

	sharedKey, _ := pre.(*curve25519.ProjPoint)
	x, _ := sharedKey.GetXY()
	b, err := x.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal binary: %w", err)
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, int64(i))
	if err != nil {
		return nil, fmt.Errorf("binary write: %w", err)
	}

	hash := crypto.Keccak256Hash(b, buf.Bytes())
	return d.suite.Scalar().SetBytes(hash.Bytes()), nil
}
