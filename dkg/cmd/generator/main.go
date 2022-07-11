package main

import (
	"client/internal/pkg/group/curve25519"
	"client/pkg/dkg"
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.dedis.ch/kyber/v3"
)

func main() {
	configFile := flag.String("c", "./configs/config.json", "filename of the config file")
	idPipe := flag.String("id-pipe", "", "filename of the named pipe used for writing the docker IDs of the zokrates containers")
	participants := flag.Int64("participants", 10, "the number of participants for the distributed key generation")
	flag.Parse()

	viper.SetConfigFile(*configFile)
	viper.SetConfigType("json")
	if err := viper.ReadInConfig(); err != nil {
		exit("Read config: %w", err)
	}

	var config dkg.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		exit("Unmarshal config into struct, %w", err)
	}

	var pipe *os.File = nil
	if *idPipe != "" {
		if pipe, err = os.OpenFile(*idPipe, os.O_WRONLY, os.ModeNamedPipe); err != nil {
			exit("Open pipe: %w", err)
		}
	}

	prover, err := dkg.NewProver(config.MountSource, pipe)
	if err != nil {
		exit("Create prover: %w", err)
	}

	defer prover.Close()

	curve := &curve25519.ProjectiveCurve{}
	curve.Init(dkg.ParamBabyJubJub(), false)
	suite := &curve25519.SuiteCurve25519{ProjectiveCurve: *curve}

	success := true
	if err := measurePolyEval(prover, int(*participants), suite, config.DkgPrivateKey); err != nil {
		success = false
		log.Errorf("Poly eval: %w", err)
	}

	if err := measureKeyDeriv(prover, int(*participants), suite); err != nil {
		success = false
		log.Errorf("Key deriv: %w", err)
	}

	if !success {
		os.Exit(1)
	}
}

func measurePolyEval(prover *dkg.Prover, participants int, suite *curve25519.SuiteCurve25519, privateKey string) error {
	threshold := (participants + 1) / 2
	args := make([]*big.Int, 0)
	pointsHashInput := make([]byte, 0)

	for i := 0; i < threshold; i++ {
		point := suite.Point().Pick(suite.RandomStream()).(*curve25519.ProjPoint)
		commitX, commitY := point.GetXY()
		args = append(args, &commitX.V, &commitY.V)

		compressed, err := point.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal commit: %w", err)
		}
		pointsHashInput = append(pointsHashInput, compressed...)
	}

	long, err := dkg.HexToScalar(suite, privateKey)
	if err != nil {
		return fmt.Errorf("hex to scalar: %w", err)
	}

	sk, _ := long.MarshalBinary()
	args = append(args, new(big.Int).SetBytes(sk))

	pubProofer := suite.Point().Mul(long, nil).(*curve25519.ProjPoint)
	pubProoferX, pubProoferY := pubProofer.GetXY()
	args = append(args, &pubProoferX.V, &pubProoferY.V)

	pubDisputer, _ := suite.Point().Pick(suite.RandomStream()).(*curve25519.ProjPoint)
	pubDisputerX, pubDisputerY := pubDisputer.GetXY()
	args = append(args, &pubDisputerX.V, &pubDisputerY.V)

	index := big.NewInt(1)
	args = append(args, index)

	share, _ := suite.Scalar().Pick(suite.RandomStream()).MarshalBinary()
	shareBig := new(big.Int).SetBytes(share)
	args = append(args, shareBig)

	pointsHash := crypto.Keccak256(pointsHashInput)

	hashInput := make([]byte, 0)

	hashInput = append(hashInput, pointsHash...)

	pubProoferBin, _ := pubProofer.MarshalBinary()
	hashInput = append(hashInput, pubProoferBin...)

	pubDisputerBin, _ := pubDisputer.MarshalBinary()
	hashInput = append(hashInput, pubDisputerBin...)

	buf := make([]byte, 32)

	hashInput = append(hashInput, index.FillBytes(buf)...)
	hashInput = append(hashInput, shareBig.FillBytes(buf)...)

	rawHash := crypto.Keccak256(hashInput)
	hash := []*big.Int{
		new(big.Int).SetBytes(rawHash[:16]),
		new(big.Int).SetBytes(rawHash[16:]),
	}

	args = append(args, hash...)

	log.Infof("Args: %v", args)

	if err := prover.ComputeWitness(context.Background(), dkg.EvalPolyProof, args); err != nil {
		return fmt.Errorf("compute witness: %w", err)
	}

	if _, err := prover.GenerateProof(context.Background(), dkg.EvalPolyProof); err != nil {
		return fmt.Errorf("generate proof: %w", err)
	}

	return nil
}

func measureKeyDeriv(prover *dkg.Prover, participants int, suite *curve25519.SuiteCurve25519) error {
	args := make([]*big.Int, 0)
	commits := make([]kyber.Point, participants)

	for i := 0; i < len(commits); i++ {
		commits[i] = suite.Point().Pick(suite.RandomStream())
	}

	firstCoefficients := make([]byte, 0)
	for i := 0; i < participants; i++ {
		commit := suite.Point().Pick(suite.RandomStream()).(*curve25519.ProjPoint)
		coeffX, coeffY := commit.GetXY()

		coeffBin, err := commit.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal coefficient: %w", err)
		}
		firstCoefficients = append(firstCoefficients, coeffBin...)

		args = append(args, &coeffX.V, &coeffY.V)
	}

	rawHash := crypto.Keccak256(firstCoefficients)
	hash := []*big.Int{
		new(big.Int).SetBytes(rawHash[:16]),
		new(big.Int).SetBytes(rawHash[16:]),
	}

	args = append(args, hash...)

	log.Infof("Args: %v", args)

	if err := prover.ComputeWitness(context.Background(), dkg.KeyDerivProof, args); err != nil {
		return fmt.Errorf("compute witness: %w", err)
	}

	if _, err := prover.GenerateProof(context.Background(), dkg.KeyDerivProof); err != nil {
		return fmt.Errorf("generate proof: %w", err)
	}

	return nil
}

func exit(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}
