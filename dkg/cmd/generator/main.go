package main

import (
	"client/internal/pkg/group/curve25519"
	"client/pkg/dkg"
	"context"
	"flag"
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
		log.Fatalf("read config: %v", err)
	}

	var config dkg.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unmarshal config into struct, %v", err)
	}

	var pipe *os.File = nil
	if *idPipe != "" {
		if pipe, err = os.OpenFile(*idPipe, os.O_WRONLY, os.ModeNamedPipe); err != nil {
			log.Errorf("Open pipe: %v", err)
			os.Exit(1)
		}
	}

	prover, err := dkg.NewProver(config.MountSource, pipe)
	if err != nil {
		log.Errorf("Create prover: %v", err)
		os.Exit(1)
	}

	curve := &curve25519.ProjectiveCurve{}
	curve.Init(dkg.ParamBabyJubJub(), false)
	suite := &curve25519.SuiteCurve25519{ProjectiveCurve: *curve}

	threshold := (*participants + 1) / 2
	args := make([]*big.Int, 0)
	commits := make([]kyber.Point, threshold)
	commitsHashInput := make([]byte, 0)

	for i := 0; i < len(commits); i++ {
		commits[i] = suite.Point().Pick(suite.RandomStream())
	}

	for i := 0; i < len(commits); i++ {
		commit := commits[i].(*curve25519.ProjPoint)
		commitX, commitY := commit.GetXY()
		args = append(args, &commitX.V, &commitY.V)

		compressed, err := commit.MarshalBinary()
		if err != nil {
			log.Errorf("Marshal commit: %v", err)
			os.Exit(1)
		}
		commitsHashInput = append(commitsHashInput, compressed...)
	}

	long, err := dkg.HexToScalar(suite, config.DkgPrivateKey)
	if err != nil {
		log.Errorf("Hex to scalar: %v", err)
		os.Exit(1)
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

	commitsHash := crypto.Keccak256(commitsHashInput)

	hashInput := make([]byte, 0)

	hashInput = append(hashInput, commitsHash...)

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
		log.Errorf("Compute witness: %v", err)
		os.Exit(1)
	}

	if _, err := prover.GenerateProof(context.Background(), dkg.EvalPolyProof); err != nil {
		log.Errorf("Compute proof: %v", err)
		os.Exit(1)
	}
}
