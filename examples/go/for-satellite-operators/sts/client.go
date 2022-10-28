package sts

import (
	"context"
	"crypto/tls"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type Client struct {
	c stellarstation.StellarStationServiceClient
}

func NewClient(apiAddress, apiKeyPath string, tlsConf *tls.Config) (*Client, error) {
	jwt, err := oauth.NewJWTAccessFromFile(apiKeyPath)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		apiAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
		grpc.WithPerRPCCredentials(jwt))
	if err != nil {
		return nil, err
	}

	client := stellarstation.NewStellarStationServiceClient(conn)

	return &Client{
		c: client,
	}, err
}

func (c *Client) GetAvailablePasses(ctx context.Context, satelliteId string) ([]*stellarstation.Pass, error) {
	res, err := c.c.ListUpcomingAvailablePasses(ctx, &stellarstation.ListUpcomingAvailablePassesRequest{
		SatelliteId: satelliteId,
	})
	if err != nil {
		return nil, err
	}

	return res.GetPass(), nil
}
