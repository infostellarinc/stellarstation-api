package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/infostellarinc/stellarstation-api/examples/go/for-satellite-operators/sts"
)

func main() {

	var apiAddress string
	var apiKey string
	var satelliteId string

	flag.StringVar(&apiKey, "key", "", "Go example configuration file path")
	flag.StringVar(&apiAddress, "addr", "api.stellarstation.com:443", "Go example configuration file path")
	flag.StringVar(&satelliteId, "id", "", "ID of the satellite to target")

	flag.Parse()

	client, err := sts.NewClient(apiAddress, apiKey, &tls.Config{})
	if err != nil {
		panic(err)
	}

	res, err := client.GetAvailablePasses(context.Background(), satelliteId)
	if err != nil {
		panic(err)
	}

	out, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
