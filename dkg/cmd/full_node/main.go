package main

import (
	"client/pkg/dkg"
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	configFile := flag.String("c", "./configs/config.json", "filename of the config file")
	idPipe := flag.String("id-pipe", "", "filename of the named pipe used for writing the docker IDs of the zokrates containers")
	rogue := flag.Bool("rogue", false, "whether the node should behave dishonest and publish invalid commitments")
	ignoreInvalid := flag.Bool("ignore-invalid", false, "do not dispute invalid shares, commitments or public keys")
	broadcastOnly := flag.Bool("broadcast-only", false, "only generate and broadcast shares and commitments, then exit")
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

	gen, err := dkg.NewDistributedKeyGenerator(&config, *idPipe, *rogue, *ignoreInvalid, *broadcastOnly)
	if err != nil {
		log.Errorf("Initializing DKG protocol: %v", err)
		os.Exit(1)
	}

	pub, err := gen.Generate()
	if err != nil {
		log.Errorf("Executing DKG protocol: %v", err)
		os.Exit(1)
	}

	if !*broadcastOnly {
		log.Infof("Public Key: %+v", pub)
	}

	os.Exit(0)
}
