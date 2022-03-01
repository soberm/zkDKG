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
	flag.Parse()

	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config: %v", err)
	}

	var config dkg.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unmarshal config into struct, %v", err)
	}

	gen, err := dkg.NewDistributedKeyGenerator(&config)
	if err != nil {
		fmt.Printf("%v", err)
	}

	distKeyShare, err := gen.Generate(context.Background())
	if err != nil {
		fmt.Printf("%v", err)
	}

	log.Infof("Public Key: %+v", distKeyShare.Public())
}
