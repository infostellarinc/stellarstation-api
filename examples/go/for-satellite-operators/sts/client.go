package sts

import (
	"context"
	"crypto/tls"
	"fmt"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type Client struct {
	c stellarstation.StellarStationServiceClient
}

func NewClient(apiAddress, apiKeyPath string, tlsConf *tls.Config) (*Client, error) {
	if apiKeyPath == "" {
		return nil, fmt.Errorf("api key is empty")
	}

	jwt, err := oauth.NewJWTAccessFromFile(apiKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not create jwt access from file(%s): %w", apiKeyPath, err)
	}

	conn, err := grpc.NewClient(
		apiAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
		grpc.WithPerRPCCredentials(jwt))
	if err != nil {
		return nil, fmt.Errorf("new grpc client: %w", err)
	}

	client := stellarstation.NewStellarStationServiceClient(conn)

	return &Client{
		c: client,
	}, nil
}

func (c *Client) GetAvailablePasses(ctx context.Context, satelliteId string) ([]*stellarstation.Pass, error) {
	res, err := c.c.ListUpcomingAvailablePasses(ctx, &stellarstation.ListUpcomingAvailablePassesRequest{
		SatelliteId: satelliteId,
	})
	if err != nil {
		return nil, fmt.Errorf("problem listing upcoming available passes for satellite (%s): %w", satelliteId, err)
	}

	return res.GetPass(), nil
}
