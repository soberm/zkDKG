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

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/mod"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
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
	ethereumPrivateKey  *ecdsa.PrivateKey
	long                kyber.Scalar
	pub                 kyber.Point
	participants        map[uint64]*Participant
	index               uint64
	priPoly             *share.PriPoly
	shares              map[uint64]kyber.Scalar
	commitments         map[uint64][]kyber.Point
	broadcastsCollected chan bool
	rogue			    bool
	ignoreInvalid	    bool
	disputed			bool
}

const bufferTimeInSecs uint16 = 2

func NewDistributedKeyGenerator(config *Config, idPipe string, rogue bool, ignoreInvalid bool) (*DistKeyGenerator, error) {

	param := ParamBabyJubJub()
	curve := &curve25519.ProjectiveCurve{}
	curve.Init(param, false)
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

	var pipe *os.File = nil
	if idPipe != "" {
		if pipe, err = os.OpenFile(idPipe, os.O_WRONLY, os.ModeNamedPipe); err != nil {
			return nil, fmt.Errorf("open pipe: %v", err)
		}
	}

	polyProver, err := NewProver(config.MountSource, pipe)
	if err != nil {
		return nil, fmt.Errorf("prover: %v", err)
	}

	return &DistKeyGenerator{
		suite:               suite,
		polyProver:          polyProver,
		curveParams:         param,
		client:              client,
		chainID:             chainID,
		contract:            contract,
		ethereumPrivateKey:  ethereumPrivateKey,
		long:                long,
		pub:                 suite.Point().Mul(long, nil),
		participants:        make(map[uint64]*Participant),
		shares:              make(map[uint64]kyber.Scalar),
		commitments:         make(map[uint64][]kyber.Point),
		broadcastsCollected: make(chan bool, 1),
		rogue:  			 rogue,
		ignoreInvalid:		 ignoreInvalid,
		disputed: 			 false,
	}, nil

}

func (d *DistKeyGenerator) Generate() (kyber.Point, error) {
	ctx, cancel := context.WithCancel(context.Background())

	log.Info("Generating distributed private key...")

	go func() {
		if err := d.WatchBroadcastSharesLog(ctx); err != nil {
			log.Errorf("Watching broadcast shares log failed: %v", err)
		}
		cancel()
	}()

	go func() {
		if err := d.WatchDisputeShareLog(ctx); err != nil {
			log.Errorf("Watching dispute share log failed: %v", err)
		}
		cancel()
	}()

	if err := d.RegisterAndWait(ctx); err != nil {
		return nil, fmt.Errorf("register and wait: %v", err)
	}

	if err := d.CollectParticipants(); err != nil {
		return nil, fmt.Errorf("collect participants: %w", err)
	}

	if err := d.BroadcastAndWait(ctx); err != nil {
		return nil, fmt.Errorf("broadcast and wait: %v", err)
	}

	disputeEnd := d.DisputeSharePeriodEnd().C

	select {
	case <-d.broadcastsCollected:
		// Do nothing
	case <-ctx.Done():
		// The context is cancelled when a broadcast is invalid or disputed or on any other error
		return nil, errors.New("can't collect valid undisputed share for every participant")
	}

	pub, err := d.ComputePublicKey()
	if err != nil {
		return nil, fmt.Errorf("compute public key: %v", err)
	}

	pubInt, err := PointToBig(pub)
	if err != nil {
		return nil, fmt.Errorf("point to big: %w", err)
	}

	// TODO Don't automatically let the first node submit the PK
	if d.index == 1 {
		select {
		case <-disputeEnd:
			if d.disputed {
				<-ctx.Done() // Wait for proof to be generated and submitted
				return nil, errors.New("own broadcast got disputed")
			}
		case <-ctx.Done():
			var whichBroadcast string
			if d.disputed {
				whichBroadcast = "own"
			} else {
				whichBroadcast = "a"
			}
			return nil, fmt.Errorf("%s broadcast got disputed", whichBroadcast)
		}

		if err := d.SubmitPublicKey(pubInt); err != nil {
			return nil, fmt.Errorf("submit public key: %v", err)
		}		
	} else {
		if err := d.WatchPublicKeySubmissionLog(ctx, pubInt); err != nil {
			return nil, fmt.Errorf("watch public key submission log: %v", err)
		}
	}

	return pub, nil
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

	log.Infof("Registered as participant with index %d", d.index)
	return nil
}

func (d *DistKeyGenerator) CollectParticipants() error {

	count, err := d.contract.CountParticipants(nil)
	if err != nil {
		return fmt.Errorf("count participants: %w", err)
	}

	for i := uint64(1); i <= count.Uint64(); i++ {
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

func (d *DistKeyGenerator) RegisterAndWait(ctx context.Context) error {

	registrationEndLogs := make(chan *ZKDKGContractRegistrationEndLog)
	defer close(registrationEndLogs)

	sub, err := d.contract.WatchRegistrationEndLog(
		&bind.WatchOpts{
			Context: ctx,
		},
		registrationEndLogs,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	if err := d.Register(); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	log.Info("Waiting until registration is finished...")

	for {
		select {
		case <-registrationEndLogs:
			return nil
		case err = <-sub.Err():
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *DistKeyGenerator) BroadcastAndWait(ctx context.Context) error {
	distributionEndLogs := make(chan *ZKDKGContractDistributionEndLog)
	defer close(distributionEndLogs)

	sub, err := d.contract.WatchDistributionEndLog(
		&bind.WatchOpts{
			Context: ctx,
		},
		distributionEndLogs,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	log.Info("Broadcasting commitments and shares...")
	if err := d.DistributeShares(); err != nil {
		return fmt.Errorf("distribute shares: %w", err)
	}

	log.Info("Waiting until distribution is finished...")

	for {
		select {
		case <-distributionEndLogs:
			return nil
		case err = <-sub.Err():
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *DistKeyGenerator) DisputeSharePeriodEnd() *time.Timer {
	var duration time.Duration
	if period, err := d.contract.SHARESDISPUTEPERIOD(nil); err == nil {
		duration, _ = time.ParseDuration(fmt.Sprintf("%ds", period + bufferTimeInSecs))
	} else {
		log.Warnf("Failed to retrieve share dispute period, using fallback value: %w", err)
		duration, _ = time.ParseDuration("5m")
	}
	return time.NewTimer(duration)
}

func (d *DistKeyGenerator) ComputePublicKey() (kyber.Point, error) {
	log.Info("Computing distributed key share...")
	distKeyShare, err := d.DistKeyShare()
	if err != nil {
		return nil, fmt.Errorf("dist key share: %w", err)
	}

	poly := share.NewPubPoly(d.suite, nil, distKeyShare.Commits)
	fig := d.suite.Point().Base().Mul(distKeyShare.Share.V, nil)
	i := int(d.index)

	test := poly.Eval(i)

	if test.V.Equal(fig) {
		log.Infof("Overall share is valid")
	}

	return distKeyShare.Public(), nil
}

func (d *DistKeyGenerator) SubmitPublicKey(pub [2]*big.Int) error {
	opts, err := bind.NewKeyedTransactorWithChainID(d.ethereumPrivateKey, d.chainID)
	if err != nil {
		return fmt.Errorf("keyed transactor with chainID: %w", err)
	}
	opts.GasPrice = big.NewInt(1000000000)

	tx, err := d.contract.SubmitPublicKey(opts, pub)
	if err != nil {
		return fmt.Errorf("submit public key: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		return fmt.Errorf("wait mined submit public key: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}
	log.Info("Submitted public key")

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
			if err := d.HandleBroadcastSharesLog(event); err != nil {
				return fmt.Errorf("handle event: %v", err)
			}
		case err = <-sub.Err():
			return err
		case <-ctx.Done():
			return nil
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

	dealerIndex := broadcastSharesLog.BroadcasterIndex
	pubKeyDealer := d.participants[dealerIndex].pub

	commitments := inputs[0].([][2]*big.Int)
	shares := inputs[1].([]*big.Int)

	commits, err := BigToPoints(d.suite, commitments)
	if err != nil {
		returnedErr := fmt.Errorf("received invalid commits from dealer %d", dealerIndex)

		if d.ignoreInvalid {
			return returnedErr
		}

		log.Infoln("Disputing invalid commits")

		if err := d.DisputeShare(dealerIndex, shares); err != nil {
			return fmt.Errorf("%v: dispute share: %v", returnedErr, err)
		}

		return returnedErr
	}

	i := d.index
	j := i
	if i > dealerIndex {
		j -= 1
	}

	fie := mod.NewInt(new(big.Int).SetBytes(shares[j - 1].Bytes()), &d.curveParams.P)

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
		log.Infof("Received valid share from dealer %v", dealerIndex)
	} else {
		returnedErr := fmt.Errorf("received invalid share from dealer %v", dealerIndex)

		if d.ignoreInvalid {
			return returnedErr
		}

		if err := d.DisputeShare(dealerIndex, shares); err != nil {
			return fmt.Errorf("%v: dispute share: %w", returnedErr, err)
		}

		return returnedErr
	}

	d.shares[dealerIndex] = fi.V
	d.commitments[dealerIndex] = commits

	if len(d.shares) == len(d.participants) {
		d.broadcastsCollected <- true
	}

	return nil
}

func (d *DistKeyGenerator) WatchDisputeShareLog(ctx context.Context) error {
	sink := make(chan *ZKDKGContractDisputeShare)
	defer close(sink)

	sub, err := d.contract.WatchDisputeShare(
		&bind.WatchOpts{
			Context: ctx,
		},
		sink,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	select {
	case event := <-sink:
		if err := d.HandleDisputeShareLog(event); err != nil {
			return fmt.Errorf("handle event: %v", err)
		}

		return nil
	case err = <-sub.Err():
		return err
	case <-ctx.Done():
		return nil
	}
}

func (d *DistKeyGenerator) HandleDisputeShareLog(disputeShareEvent *ZKDKGContractDisputeShare) error {
	if d.index != disputeShareEvent.DisputeeIndex {
		log.Info("Received dispute for other participant, aborting")
		return nil
	}

	d.disputed = true

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

	hashInput = append(hashInput, pubProoferX.V.FillBytes(buf)...)
	hashInput = append(hashInput, pubProoferY.V.FillBytes(buf)...)

	hashInput = append(hashInput, pubDisputerX.V.FillBytes(buf)...)
	hashInput = append(hashInput, pubDisputerY.V.FillBytes(buf)...)

	hashInput = append(hashInput, index.FillBytes(buf)...)

	hashInput = append(hashInput, fiBig.FillBytes(buf)...)

	rawHash := crypto.Keccak256(hashInput)
	hash := []*big.Int{
		new(big.Int).SetBytes(rawHash[:16]),
		new(big.Int).SetBytes(rawHash[16:]),
	}

	args = append(args, hash...)

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

func (d *DistKeyGenerator) WatchPublicKeySubmissionLog(ctx context.Context, pk [2]*big.Int) error {
	sink := make(chan *ZKDKGContractPublicKeySubmission)
	defer close(sink)

	sub, err := d.contract.WatchPublicKeySubmission(
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
			log.Infof("Handling public key submission log...")
			if err := d.HandlePublicKeySubmissionLog(event, pk); err != nil {
				log.Errorf("Handling public key submission log failed: %v", err)
			}
		case err = <-sub.Err():
			return err
		case <-ctx.Done():
			return nil
		}
	}
}

func (d *DistKeyGenerator) HandlePublicKeySubmissionLog(pkSubmissionLog *ZKDKGContractPublicKeySubmission, computedPk [2]*big.Int) error {
	submissionTx, _, err := d.client.TransactionByHash(context.Background(), pkSubmissionLog.Raw.TxHash)
	if err != nil {
		return fmt.Errorf("transaction by hash: %w", err)
	}

	txData := submissionTx.Data()
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

	submittedPk := inputs[0].([2]*big.Int)

	if computedPk[0].Cmp(submittedPk[0]) == 0 && computedPk[1].Cmp(submittedPk[1]) == 0 {
		log.Infoln("Public key valid")
		return nil
	}

	log.Infoln("Submitted public key invalid")

	if d.ignoreInvalid {
		return nil
	}

	args := make([]*big.Int, 0)

	firstCoefficients := make([]byte, 0)
	for i := uint64(0); i < uint64(len(d.participants)); i++ {
		firstCoefficient := d.commitments[i][0]

		bin, err := firstCoefficient.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal binary coefficient: %w", err)
		}

		point, err := PointToBig(firstCoefficient)
		if err != nil {
			return fmt.Errorf("coefficient to big: %w", err)
		}
		firstCoefficients = append(firstCoefficients, bin...)
		args = append(args, point[:]...)
	}

	args = append(args, submittedPk[:]...)

	rawHash := crypto.Keccak256(firstCoefficients)
	hash := []*big.Int{
		new(big.Int).SetBytes(rawHash[:16]),
		new(big.Int).SetBytes(rawHash[16:]),
	}

	args = append(args, hash...)

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

	disputeTx, err := d.contract.DisputePublicKey(opts, KeyVerifierProof(*proof.Proof))
	if err != nil {
		return fmt.Errorf("dispute public key: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), d.client, disputeTx)
	if err != nil {
		return fmt.Errorf("wait mined register: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("receipt status failed")
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
		_, c := pubPoly.Info()
		log.Infof("Adding commitments: %v", c)
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
	threshold, err := d.contract.Threshold(nil)
	if err != nil {
		return fmt.Errorf("threshold: %w", err)
	}

	secret := d.suite.Scalar().Pick(d.suite.RandomStream())
	d.priPoly = share.NewPriPoly(d.suite, int(threshold.Int64()), secret, d.suite.RandomStream())
	pubPoly := d.priPoly.Commit(nil)

	_, commits := pubPoly.Info()

	commitsString := "Commitments"
	if d.rogue {
		commitsString = "Fake commitments"
		commits[0].Null()
	}

	d.commitments[d.index] = commits
	d.shares[d.index] = d.priPoly.Eval(int(d.index)).V

	commitments, err := PointsToBig(commits)
	if err != nil {
		return fmt.Errorf("points to big: %w", err)
	}

	log.Infof("%s: %v", commitsString, commits)

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
	opts.GasLimit = 200000

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
	pre := dhExchange(d.suite, privateKey, publicKey)

	sharedKey, _ := pre.(*curve25519.ProjPoint)
	x, _ := sharedKey.GetXY()
	b, err := x.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal binary: %w", err)
	}

	commitsBin, err := commits[0].MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal commit: %w", err)
	}

	hash := crypto.Keccak256Hash(
		b,
		commitsBin,
	)
	return mod.NewInt(new(big.Int).SetBytes(hash.Bytes()), &d.curveParams.P), nil
}
