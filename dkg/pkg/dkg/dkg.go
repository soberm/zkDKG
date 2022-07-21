package dkg

import (
	"client/internal/pkg/group/curve25519"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/iden3/go-iden3-crypto/poseidon"
	log "github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/mod"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/sync/errgroup"
)

type Participant struct {
	index uint64
	pub   kyber.Point
}

type DistKeyGenerator struct {
	suite               suites.Suite
	polyProver          *Prover
	curveParams         *curve25519.Param
	client              *ethclient.Client
	chainID             *big.Int
	contract            *ZKDKGContract
	contractAbi			abi.ABI
	contractAddress		common.Address
	ethereumAddress  	common.Address
	ethereumPrivateKey	*ecdsa.PrivateKey
	long                kyber.Scalar
	extendedPeriod		bool
	pub                 kyber.Point
	participants        map[uint64]*Participant
	index               uint64
	priPoly             *share.PriPoly
	shares              map[uint64]kyber.Scalar
	commitments         map[uint64][]kyber.Point
	rogue			    bool
	ignoreInvalid	    bool
	broadcastOnly		bool
}

var errAbortion error = errors.New("protocol aborted due to insufficient remaining participants")
const bufferTimeInSecs uint64 = 2

func NewDistributedKeyGenerator(config *Config, idPipe string, rogue, ignoreInvalid, broadcastOnly bool) (*DistKeyGenerator, error) {

	param := ParamBabyJubJub()
	curve := &curve25519.ProjectiveCurve{}
	curve.Init(param, false)
	suite := &curve25519.SuiteCurve25519{ProjectiveCurve: *curve}

	client, err := ethclient.Dial(config.EthereumNode)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("chainID: %w", err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(ZKDKGContractMetaData.ABI))
	if err != nil {
		return nil, fmt.Errorf("read abi: %w", err)
	}

	contractAddress := common.HexToAddress(config.ContractAddress)

	contract, err := NewZKDKGContract(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("zkDKG contract: %w", err)
	}

	ethereumPrivateKey, err := crypto.HexToECDSA(config.EthereumPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("hex to ecdsa: %w", err)
	}

	ethereumPublicKey := crypto.PubkeyToAddress(ethereumPrivateKey.PublicKey)

	long, err := HexToScalar(suite, config.DkgPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("hex to scalar: %w", err)
	}

	var pipe *os.File = nil
	if idPipe != "" {
		if pipe, err = os.OpenFile(idPipe, os.O_WRONLY, os.ModeNamedPipe); err != nil {
			return nil, fmt.Errorf("open pipe: %w", err)
		}
	}

	polyProver, err := NewProver(config.MountSource, pipe)
	if err != nil {
		return nil, fmt.Errorf("prover: %w", err)
	}

	return &DistKeyGenerator{
		suite:               suite,
		polyProver:          polyProver,
		curveParams:         param,
		client:              client,
		chainID:             chainID,
		contract:            contract,
		contractAbi: 		 contractAbi,
		contractAddress: 	 contractAddress,
		ethereumAddress:     ethereumPublicKey,
		ethereumPrivateKey:  ethereumPrivateKey,
		long:                long,
		pub:                 suite.Point().Mul(long, nil),
		participants:        make(map[uint64]*Participant),
		shares:              make(map[uint64]kyber.Scalar),
		commitments:         make(map[uint64][]kyber.Point),
		rogue:  			 rogue,
		ignoreInvalid:		 ignoreInvalid,
		broadcastOnly:		 broadcastOnly,
	}, nil

}

func (d *DistKeyGenerator) Generate() (kyber.Point, error) {
	ctx := context.Background()

	log.Info("Generating distributed private key...")

	distributionEnd := make(chan struct{})
	broadcastsCollected := make(chan struct{})
	g, ctx := errgroup.WithContext(ctx)

	if !d.broadcastOnly {
		g.Go(func() error {
			if err := d.WatchBroadcastSharesLog(ctx, distributionEnd, broadcastsCollected); err != nil {
				return fmt.Errorf("watching broadcast shares log failed: %w", err)
			}
			return nil
		})

		g.Go(func() error {
			if err := d.WatchDistributionEndLog(ctx); err != nil {
				return fmt.Errorf("watching distribution end log failed: %w", err)
			}
			close(distributionEnd)
			return nil
		})

		g.Go(func() error {
			if err := d.WatchDisputeShareLog(ctx); err != nil {
				return fmt.Errorf("watching dispute share log failed: %w", err)
			}
			return nil
		})

		g.Go(func() error {
			if err := d.WatchExclusion(ctx); err != nil {
				return fmt.Errorf("watching exclusion failed: %w", err)
			}
			return nil
		})

		g.Go(func() error {
			if err := d.WatchAbortion(ctx); err != nil {
				if errors.Is(err, errAbortion) {
					return errAbortion
				}
				return fmt.Errorf("watching abortion failed: %w", err)
			}
			return nil
		})
	}

	if err := d.RegisterAndWait(ctx); err != nil {
		return nil, fmt.Errorf("register and wait: %w", err)
	}

	if err := d.CollectParticipants(); err != nil {
		return nil, fmt.Errorf("collect participants: %w", err)
	}

	if err := d.DistributeShares(); err != nil {
		return nil, fmt.Errorf("distribute shares: %w", err)
	}

	if d.broadcastOnly {
		return nil, nil
	}

	select {
	case <-distributionEnd:
		// Do nothing
	case <-ctx.Done():
		// The context is cancelled when an unexpected error has occurred in one of the goroutines
		return nil, g.Wait()
	}

	disputeEnd := d.DisputeSharePeriodEnd()

	<-broadcastsCollected

	select {
	case <-disputeEnd:
		// Do nothing
	case <-ctx.Done():
		// The context is cancelled when an unexpected error has occurred in one of the goroutines
		return nil, g.Wait()
	}

	if err := d.checkExpiredDisputes(); err != nil {
		return nil, fmt.Errorf("check expired disputes: %w", err)
	}

	pub, err := d.ComputePublicKey()
	if err != nil {
		return nil, fmt.Errorf("compute public key: %w", err)
	}

	pkLog := make(chan struct{})
	g.Go(func() error {
		if err := d.WatchPublicKeySubmissionLog(ctx, pub); err != nil {
			return fmt.Errorf("watching public key submission log failed: %w", err)
		}
		close(pkLog)
		return nil
	})

	if err := d.SubmitPublicKey(pub); err != nil {
		if errors.Is(err, errAbortion) {
			return nil, err
		}

		if ctx.Err() != nil {
			return nil, g.Wait()
		}

		log.Warnf("Public key submission failed, waiting for other participant's submission: %v", err)

		select {
		case <-pkLog:
			// Do nothing
		case <-ctx.Done():
			// The context is cancelled when an unexpected error has occurred in one of the goroutines
			return nil, g.Wait()
		}
	}

	return pub, nil
}

func (d *DistKeyGenerator) Register(ctx context.Context) error {
	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	pub, err := PointToBig(d.pub)
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}

	estimate, err := d.estimateGas(ctx, "register", pub)
	if err != nil {
		return fmt.Errorf("estimate gas: %w", err)
	}

	opts.GasLimit = estimate + 30000

	tx, err := d.contract.Register(opts, pub)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, d.client, tx)
	if err != nil {
		return fmt.Errorf("wait mined register: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}

	participant, err := d.contract.Participants(nil, d.ethereumAddress)
	if err != nil {
		return fmt.Errorf("participants: %w", err)
	}

	d.index = participant.Index

	log.Infof("Registered as participant with index %d", d.index)
	return nil
}

func (d *DistKeyGenerator) CollectParticipants() error {

	log.Info("Collecting participants...")

	// TODO Pass context to this and similar calls
	pks, err := d.contract.PublicKeys(nil)
	if err != nil {
		return fmt.Errorf("collect public keys: %w", err)
	}

	for i := uint64(1); i <= uint64(len(pks)); i++ {
		pub, err := BigToPoint(d.suite, pks[i-1])
		if err != nil {
			return fmt.Errorf("big to point: %w", err)
		}

		d.participants[i] = &Participant{index: i, pub: pub}
	}

	return nil

}

func WatchEvent[K any](
	ctx context.Context,
	subscribeLog func(*bind.WatchOpts, chan<- K) (event.Subscription, error),
	afterSubscribe func() error,
	handleEvent func(K) error,
	once bool,
) error {
	events := make(chan K)
	defer close(events)

	sub, err := subscribeLog(
		&bind.WatchOpts{
			Context: ctx,
		},
		events,
	)
	if err != nil {
		return fmt.Errorf("subscribe log: %w", err)
	}
	defer sub.Unsubscribe()

	if afterSubscribe != nil {
		if err := afterSubscribe(); err != nil {
			return fmt.Errorf("after subscribe: %w", err)
		}
	}

	for {
		select {
		case event := <-events:
			if handleEvent != nil {
				if err := handleEvent(event); err != nil {
					return fmt.Errorf("handle event: %w", err)
				}
			}
		case err := <-sub.Err():
			return fmt.Errorf("subscription: %w", err)
		case <-ctx.Done():
			return fmt.Errorf("context: %w", ctx.Err())
		}

		if once {
			return nil
		}
	}
}

func (d *DistKeyGenerator) RegisterAndWait(ctx context.Context) error {
	return WatchEvent(
		ctx,
		d.contract.WatchRegistrationEndLog,
		func() error {
			if err := d.Register(ctx); err != nil {
				return fmt.Errorf("register: %w", err)
			}

			log.Info("Waiting until registration is finished...")
			return nil
		},
		nil,
		true,
	)
}

func (d *DistKeyGenerator) DisputeSharePeriodEnd() <-chan struct{} {
	end := make(chan struct{})

	go func() {
		timer := time.NewTimer(d.durationUntilPhaseEnd())

		for {
			<-timer.C
			if !d.extendedPeriod {
				break
			}

			d.extendedPeriod = false
			timer.Reset(d.durationUntilPhaseEnd())
		}
		close(end)
	}()

	return end
}

func (d *DistKeyGenerator) durationUntilPhaseEnd() time.Duration {
	if period, err := d.contract.PhaseEnd(nil); err != nil {
		log.Warnf("Failed to retrieve current phase end, using fallback value: %v", err)
		duration, _ := time.ParseDuration("5m")
		return duration
	} else {
		return time.Until(time.Unix(int64(period + bufferTimeInSecs), 0))
	}
}

func (d *DistKeyGenerator) ComputePublicKey() (kyber.Point, error) {
	log.Info("Computing distributed key share...")
	distKeyShare, err := d.DistKeyShare()
	if err != nil {
		return nil, fmt.Errorf("dist key share: %w", err)
	}

	poly := share.NewPubPoly(d.suite, nil, distKeyShare.Commits)
	fig := d.suite.Point().Base().Mul(distKeyShare.Share.V, nil)
	i := int(d.index) - 1

	test := poly.Eval(i)

	if !test.V.Equal(fig) {
		return nil, errors.New("overall share is invalid")
	}

	return distKeyShare.Public(), nil
}

func (d *DistKeyGenerator) checkExpiredDisputes() error {
	indices, err := d.contract.ExpiredDisputes(nil, big.NewInt(time.Now().Unix()))
	if err != nil {
		return fmt.Errorf("contract call: %w", err)
	}

	for i, expired := range indices {
		if expired {
			d.HandleExclusion(uint64(i + 1))
		}
	}

	return nil
}

func (d *DistKeyGenerator) SubmitPublicKey(pub kyber.Point) error {
	defer d.polyProver.Close()

	args := make([]*big.Int, 0)

	firstCoefficients := make([]byte, 0)
	for i := uint64(1); i <= uint64(len(d.participants)); i++ {
		firstCoefficient := d.commitments[i][0]

		coeffProj, _ := firstCoefficient.(*curve25519.ProjPoint)
		coeffX, coeffY := coeffProj.GetXY()

		coeffBin, err := firstCoefficient.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal coefficient: %w", err)
		}
		firstCoefficients = append(firstCoefficients, coeffBin...)

		args = append(args, &coeffX.V, &coeffY.V)
	}

	hash := truncateHash(crypto.Keccak256(firstCoefficients))

	args = append(args, new(big.Int).SetBytes(hash))

	// This is actually necessary instead of simply accessing via the X and Y properties due to the normalization that takes place in GetXY
	pubX, pubY := pub.(*curve25519.ProjPoint).GetXY()
	pubXY := [2]*big.Int{&pubX.V, &pubY.V}

	log.Infof("Args: %d", args)

	if err := d.polyProver.ComputeWitness(context.Background(), KeyDerivProof, args); err != nil {
		return fmt.Errorf("compute witness for public key proof: %w", err)
	}

	proof, err := d.polyProver.GenerateProof(context.Background(), KeyDerivProof)
	if err != nil {
		return fmt.Errorf("generate proof for public key: %w", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	submitTx, err := d.contract.SubmitPublicKey(opts, pubXY, KeyVerifierProof(*proof.Proof))
	if err != nil {
		return fmt.Errorf("submit public key: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), d.client, submitTx)
	if err != nil {
		return fmt.Errorf("wait mined submit: %w", err)
	}

	for _, eventLog := range receipt.Logs {
		if eventLog.Topics[0] == crypto.Keccak256Hash([]byte("Abortion()")) {
			return errAbortion
		}
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}
	log.Info("Submitted public key")

	return nil
}

func (d *DistKeyGenerator) WatchBroadcastSharesLog(ctx context.Context, distributionEnd, broadcastsCollected chan struct{}) error {
	return WatchEvent(
		ctx,
		d.contract.WatchBroadcastSharesLog,
		nil,
		func(event *ZKDKGContractBroadcastSharesLog) error {
			return d.HandleBroadcastSharesLog(event, distributionEnd, broadcastsCollected)
		},
		false,
	)
}

func (d *DistKeyGenerator) HandleBroadcastSharesLog(broadcastSharesLog *ZKDKGContractBroadcastSharesLog, distributionEnd, broadcastsCollected chan struct{}) error {
	if d.ethereumAddress == broadcastSharesLog.Sender {
		// Ignore own broadcast
		return nil
	}

	inputs, err := d.getTxInputs(broadcastSharesLog.Raw.TxHash)
	if err != nil {
		return fmt.Errorf("get tx inputs: %w", err)
	}

	commitments := inputs[0].([]*big.Int)
	shares := inputs[1].([]*big.Int)

	dealerIndex := broadcastSharesLog.BroadcasterIndex
	pubKeyDealer := d.participants[dealerIndex].pub

	i := d.index
	j := i
	if i > dealerIndex {
		j -= 1
	}

	fie := mod.NewInt(new(big.Int).SetBytes(shares[j - 1].Bytes()), &d.curveParams.P)

	valid := true

	var decryptedShare kyber.Scalar
	commits, err := BigToPoints(d.suite, commitments)
	if err != nil {
		valid = false

		log.Infof("Received invalid curve points from dealer %d", dealerIndex)
		if !d.ignoreInvalid {
			d.scheduleDispute(dealerIndex, shares, distributionEnd)
		}
	} else {
		sharedKey, err := d.PreSharedKey(d.long, pubKeyDealer, commits)
		if err != nil {
			return fmt.Errorf("pre shared key: %w", err)
		}

		fi := &share.PriShare{
			I: int(i) - 1,
			V: d.suite.Scalar().Sub(fie, sharedKey),
		}

		pubPoly := share.NewPubPoly(d.suite, nil, commits)

		if pubPoly.Check(fi) {
			decryptedShare = fi.V
		} else {
			log.Infof("Received invalid share from dealer %d", dealerIndex)
			valid = false

			if !d.ignoreInvalid {
				d.scheduleDispute(dealerIndex, shares, distributionEnd)
			}
		}
	}

	if valid {
		log.Infof("Received valid broadcast from dealer %d", dealerIndex)
	} else {
		decryptedShare = d.suite.Scalar()

		commits = make([]kyber.Point, len(commitments))
		for i := range commits {
			commits[i] = d.suite.Point()
		}
	}

	d.shares[dealerIndex] = decryptedShare
	d.commitments[dealerIndex] = commits

	if len(d.shares) == len(d.participants) {
		close(broadcastsCollected)
	}

	return nil
}

func (d *DistKeyGenerator) WatchDistributionEndLog(ctx context.Context) error {
	return WatchEvent(
		ctx,
		d.contract.WatchDistributionEndLog,
		nil,
		nil,
		true,
	)
}

func (d *DistKeyGenerator) WatchDisputeShareLog(ctx context.Context) error {
	return WatchEvent(
		ctx,
		d.contract.WatchDisputeShare,
		nil,
		d.HandleDisputeShareLog,
		false,
	)
}

func (d *DistKeyGenerator) HandleDisputeShareLog(disputeShareEvent *ZKDKGContractDisputeShare) error {
	log.Infof("Received dispute for dealer %d", disputeShareEvent.DisputeeIndex)

	d.extendedPeriod = true

	if d.index != disputeShareEvent.DisputeeIndex {
		return nil
	}

	log.Info("Received dispute against own broadcast, defending")

	args := make([]*big.Int, 0)

	for _, commitment := range d.commitments[d.index] {
		c, _ := commitment.(*curve25519.ProjPoint)
		x, y := c.GetXY()
		args = append(args, &x.V, &y.V)
	}

	sk, _ := d.long.MarshalBinary()
	args = append(args, new(big.Int).SetBytes(sk))

	pubProofer := d.pub.(*curve25519.ProjPoint)
	pubProoferX, pubProoferY := pubProofer.GetXY()
	args = append(args, &pubProoferX.V, &pubProoferY.V)

	pubDisputer := d.participants[disputeShareEvent.DisputerIndex].pub.(*curve25519.ProjPoint)
	pubDisputerX, pubDisputerY := pubDisputer.GetXY()
	args = append(args, &pubDisputerX.V, &pubDisputerY.V)

	index := big.NewInt(int64(disputeShareEvent.DisputerIndex))
	args = append(args, index)

	priShare, _ := d.EncryptedPrivateShare(disputeShareEvent.DisputerIndex, d.commitments[d.index])
	fiBinary, _ := priShare.V.MarshalBinary()
	fiBig := new(big.Int).SetBytes(fiBinary)
	args = append(args, fiBig)

	buf := make([]byte, 32)

	hashInput := make([]byte, 0)

	a, err := d.contract.Addresses(nil, big.NewInt(int64(d.index) - 1))
	if err != nil {
		return fmt.Errorf("get address: %w", err)
	}

	commitmentsHash, err := d.contract.CommitmentHashes(nil, a)
	if err != nil {
		return fmt.Errorf("commitment hashes: %w", err)
	}

	hashInput = append(hashInput, commitmentsHash[:]...)

	pubProoferBin, _ := pubProofer.MarshalBinary()
	pubDisputerBin, _ := pubDisputer.MarshalBinary()
	hashInput = append(hashInput, pubProoferBin...)
	hashInput = append(hashInput, pubDisputerBin...)

	hashInput = append(hashInput, index.FillBytes(buf)...)

	hashInput = append(hashInput, fiBig.FillBytes(buf)...)

	hash := truncateHash(crypto.Keccak256(hashInput))

	args = append(args, new(big.Int).SetBytes(hash))

	log.Infof("Args: %d", args)

	err = d.polyProver.ComputeWitness(context.Background(), EvalPolyProof, args)
	if err != nil {
		return fmt.Errorf("compute witness: %w", err)
	}

	proof, err := d.polyProver.GenerateProof(context.Background(), EvalPolyProof)
	if err != nil {
		return fmt.Errorf("compute witness: %w", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	tx, err := d.contract.DefendShare(opts, ShareVerifierProof(*proof.Proof))
	if err != nil {
		return fmt.Errorf("dispute share: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		return fmt.Errorf("wait mined: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}

	log.Infoln("Share successfully defended")

	return nil
}

func (d *DistKeyGenerator) WatchExclusion(ctx context.Context) error {
	return WatchEvent(
		ctx,
		d.contract.WatchExclusion,
		nil,
		func(event *ZKDKGContractExclusion) error {
			d.HandleExclusion(event.Index)
			return nil
		},
		false,
	)
}

func (d *DistKeyGenerator) WatchAbortion(ctx context.Context) error {
	return WatchEvent(
		ctx,
		d.contract.WatchAbortion,
		nil,
		func(*ZKDKGContractAbortion) error {
			return errAbortion
		},
		true,
	)
}

func (d *DistKeyGenerator) HandleExclusion(index uint64) {
	d.commitments[index][0] = d.suite.Point().Null()
	d.shares[index] = d.suite.Scalar()
}

func (d *DistKeyGenerator) WatchPublicKeySubmissionLog(ctx context.Context, computedPk kyber.Point) error {
	return WatchEvent(
		ctx,
		d.contract.WatchPublicKeySubmission,
		nil,
		func(event *ZKDKGContractPublicKeySubmission) error {
			return d.HandlePublicKeySubmissionLog(ctx, computedPk, event)
		},
		true,
	)
}

func (d *DistKeyGenerator) HandlePublicKeySubmissionLog(ctx context.Context, computedPk kyber.Point, event *ZKDKGContractPublicKeySubmission) error {
	inputs, err := d.getTxInputs(event.Raw.TxHash)
	if err != nil {
		return fmt.Errorf("get tx inputs: %w", err)
	}

	submittedPkBig := inputs[0].([2]*big.Int)
	submittedPk := d.suite.Point().(*curve25519.ProjPoint)
	submittedPkX, submittedPkY := submittedPk.GetXY()
	submittedPkX.V.Set(submittedPkBig[0])
	submittedPkY.V.Set(submittedPkBig[1])

	if !computedPk.Equal(submittedPk) {
		return errors.New("computed public key differs from submitted public key")
	}
	
	return nil
}

func (d *DistKeyGenerator) DisputeShare(disputeeIndex uint64, shares []*big.Int) error {
	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	tx, err := d.contract.DisputeShare(opts, disputeeIndex, shares)
	if err != nil {
		return fmt.Errorf("dispute share: %w", err)
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
			I: int(d.index) - 1,
			V: sh,
		},
		PrivatePoly: d.priPoly.Coefficients(),
	}, nil
}

func (d *DistKeyGenerator) DistributeShares() error {
	threshold, err := d.contract.MinimumThreshold(nil)
	if err != nil {
		return fmt.Errorf("threshold: %w", err)
	}

	log.Info("Generating commitments and shares...")

	secret := d.suite.Scalar().Pick(d.suite.RandomStream())
	d.priPoly = share.NewPriPoly(d.suite, int(threshold.Int64()), secret, d.suite.RandomStream())
	pubPoly := d.priPoly.Commit(nil)

	_, commits := pubPoly.Info()

	if d.rogue {
		commits[0].Neg(commits[0])
	}

	d.commitments[d.index] = commits
	d.shares[d.index] = d.priPoly.Eval(int(d.index) - 1).V

	commitments, err := PointsToBig(commits)
	if err != nil {
		return fmt.Errorf("points to big: %w", err)
	}

	shares := make([]*big.Int, 0)
	for i := uint64(1); i <= uint64(len(d.participants)); i++ {
		if i == d.index {
			continue
		}

		participant := d.participants[i]

		priShare, err := d.EncryptedPrivateShare(participant.index, commits)
		if err != nil {
			return fmt.Errorf("encrypted private share: %w", err)
		}

		b, err := priShare.V.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal binary share: %w", err)
		}

		shares = append(shares, new(big.Int).SetBytes(b))
	}

	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	estimate, err := d.estimateGas(context.Background(), "broadcastShares", commitments, shares)
	if err != nil {
		return fmt.Errorf("estimate gas: %w", err)
	}
	opts.GasLimit = estimate + 30000

	tx, err := d.contract.BroadcastShares(opts, commitments, shares)
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

func (d *DistKeyGenerator) EncryptedPrivateShare(i uint64, commits []kyber.Point) (*share.PriShare, error) {
	priShare := d.priPoly.Eval(int(i) - 1)

	sharedKey, err := d.PreSharedKey(d.long, d.participants[i].pub, commits)
	if err != nil {
		return nil, fmt.Errorf("pre shared key: %w", err)
	}

	sharedKey.Add(sharedKey, priShare.V)
	priShare.V = sharedKey

	return priShare, nil
}

func (d *DistKeyGenerator) PreSharedKey(privateKey kyber.Scalar, publicKey kyber.Point, commits []kyber.Point) (kyber.Scalar, error) {
	pre := DhExchange(d.suite, privateKey, publicKey)

	sharedKey, _ := pre.(*curve25519.ProjPoint)
	sharedKeyX, _ := sharedKey.GetXY()

	commit, _ := commits[0].(*curve25519.ProjPoint)
	commitX, _ := commit.GetXY()

	hash, err := poseidon.Hash([]*big.Int{&sharedKeyX.V, &commitX.V})
	if err != nil {
		return nil, fmt.Errorf("poseidon: %w", err)
	}

	return mod.NewInt(hash, &d.curveParams.P), nil
}

func (d *DistKeyGenerator) estimateGas(ctx context.Context, fn string, args ...interface{}) (uint64, error) {
	data, err := d.contractAbi.Pack(fn, args...)
	if err != nil {
		return 0, fmt.Errorf("pack args: %w", err)
	}

	return d.client.EstimateGas(ctx, ethereum.CallMsg{
		From: d.ethereumAddress,
		To: &d.contractAddress,
		Data: data,
	})
}

func (d *DistKeyGenerator) scheduleDispute(dealerIndex uint64, shares []*big.Int, distributionEnd <-chan struct{}) {
	log.Infof("Starting dispute against dealer %d after distribution end", dealerIndex)

	go func() {
		<-distributionEnd

		log.Infof("Disputing invalid broadcast from dealer %d", dealerIndex)

		if err := d.DisputeShare(dealerIndex, shares); err != nil {
			log.Errorf("Dispute commits: %v", err)
		}
	}()
}

func (d *DistKeyGenerator) getTxInputs(txHash common.Hash) ([]interface{}, error) {
	tx, _, err := d.client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, fmt.Errorf("transaction by hash: %w", err)
	}

	txData := tx.Data()
	a, err := abi.JSON(strings.NewReader(ZKDKGContractABI))
	if err != nil {
		return nil, fmt.Errorf("abi from json: %w", err)
	}

	method, err := a.MethodById(txData[:4])
	if err != nil {
		return nil, fmt.Errorf("method by id: %w", err)
	}

	inputs, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return nil, fmt.Errorf("unpack inputs: %w", err)
	}

	return inputs, nil
}

func truncateHash(hash []byte) ([]byte) {
	// Truncate the first 3 bits s.t. value range is limited to 254 bits (field size of BabyJubJub)
	hash[0] &= 0b00011111

	return hash
}
