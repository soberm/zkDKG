package main

import (
	"client/pkg/dkg"
	"context"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	configFile := flag.String("c", "./configs/config_1.json", "filename of the config file")
	idPipe := flag.String("id-pipe", "", "filename of the named pipe used for writing the docker IDs of the zokrates containers")
	rogue := flag.Bool("rogue", false, "whether the node should behave dishonest and publish invalid commitments")
	ignoreInvalid := flag.Bool("ignore-invalid", false, "do not dispute invalid shares, commitments or public keys")
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

	gen, err := dkg.NewDistributedKeyGenerator(&config, *idPipe, *rogue, *ignoreInvalid)
	if err != nil {
		fmt.Printf("%v", err)
	}

	distKeyShare, err := gen.Generate(context.Background())
	if err != nil {
		fmt.Printf("%v", err)
	}

	log.Infof("Public Key: %+v", distKeyShare.Public())
}
