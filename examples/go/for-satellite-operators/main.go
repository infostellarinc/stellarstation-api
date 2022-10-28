package main

import (
	"context"
	"crypto/tls"
	"fmt"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func main() {
	apiUrl := "api.stellarstation.com:443"
	akiKeyPath := "./api-key.json"

	jwt, err := oauth.NewJWTAccessFromFile(akiKeyPath)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(
		apiUrl,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(jwt))
	if err != nil {
		panic(err)
	}

	client := stellarstation.NewStellarStationServiceClient(conn)
	ctx := context.Background()
	res, err := client.ListUpcomingAvailablePasses(ctx, &stellarstation.ListUpcomingAvailablePassesRequest{
		SatelliteId: "297",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.GetPass())
}
