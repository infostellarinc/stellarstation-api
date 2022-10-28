package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/infostellarinc/stellarstation-api/examples/go/for-satellite-operators/sts"
)

func main() {

	var configPath string
	var satelliteId string

	flag.StringVar(&configPath, "c", configPath, "Go example configuration file path")
	flag.StringVar(&satelliteId, "id", satelliteId, "ID of the satellite to target")

	flag.Parse()

	conf, err := sts.GetConfig(configPath)
	if err != nil {
		panic(err)
	}

	client, err := sts.NewClient(conf.ApiAddress, conf.ApiKeyPath, &tls.Config{})
	if err != nil {
		panic(err)
	}

	res, err := client.GetAvailablePasses(context.Background(), satelliteId)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
