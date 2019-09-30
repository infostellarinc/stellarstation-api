/*
 * Copyright 2019 Infostellar, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

/**************
 * API Client *
 **************/

// Client is a StellarStation Ground Station API client.
type Client struct {
	groundstation GroundStationConfig
	client        api.GroundStationServiceClient
	runner        *Runner
}

// NewClient creates a new Client instance
func NewClient() *Client {
	client := &Client{
		runner: NewRunner(),
	}
	return client
}

// Connect connects to the ground station.
func (c *Client) Connect(groundstation GroundStationConfig) {
	startFunction := func(r *Runner) {
		log.Printf("Connecting to ground station %v (%v).\n", groundstation.Name, groundstation.ID)
		log.Printf("Ground station configuration: %+v\n", groundstation)

		c.groundstation = groundstation

		c.connect()
	}

	stopFunction := func(r *Runner) {
		log.Printf("Disconnecting from ground station %v (%v).\n", groundstation.Name, groundstation.ID)
	}

	c.runner.Start(startFunction, stopFunction)
}

// Stop shuts down the client
func (c *Client) Stop() {
	c.runner.Stop()
}

// Wait waits for the client to shut down
func (c *Client) Wait() {
	c.runner.Wait()
}

/*************
 * API Calls *
 *************/

// connect connects to the selected ground station
func (c *Client) connect() {
	jwtCreds, err := oauth.NewJWTAccessFromFile(
		c.groundstation.Key)
	if err != nil {
		log.Fatalf("Failed to create JWT credentials: %v", err)
	}

	tc := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	conn, err := grpc.Dial(c.groundstation.Address,
		grpc.WithTransportCredentials(tc),
		grpc.WithPerRPCCredentials(jwtCreds))

	if err != nil {
		log.Fatalf("Couldn't connect to ground station: %v\n", err)
	}

	go func() {
		c.runner.Wait()
		conn.Close()
	}()

	c.client = api.NewGroundStationServiceClient(conn)
}

// ListPlans gets upcoming plans for the given ground station
func (c *Client) ListPlans(start time.Time, end time.Time) ([]*api.Plan, error) {
	startTs, _ := ptypes.TimestampProto(start)
	endTs, _ := ptypes.TimestampProto(end)

	listPlansRequest := &api.ListPlansRequest{
		GroundStationId: c.groundstation.ID,
		AosAfter:        startTs,
		AosBefore:       endTs,
	}

	//log.Printf("ListPlans Request: %+v\n", listPlansRequest)

	listPlansResponse, err := c.client.ListPlans(context.Background(), listPlansRequest)
	if err != nil {
		return nil, err
	}

	//log.Printf("ListPlans Response: %+v\n", listPlansResponse)

	return listPlansResponse.Plan, nil
}

// OpenGroundStationStream returns a bidirectional streaming client.
func (c *Client) OpenGroundStationStream(ctx context.Context) (api.GroundStationService_OpenGroundStationStreamClient, error) {
	return c.client.OpenGroundStationStream(ctx)
}

// TelemetryRequest wraps a telemetry message in a request object.
func (c *Client) TelemetryRequest(telemetry *api.SatelliteTelemetry) *api.GroundStationStreamRequest {
	return &api.GroundStationStreamRequest{
		GroundStationId: c.groundstation.ID,
		Request: &api.GroundStationStreamRequest_SatelliteTelemetry{
			SatelliteTelemetry: telemetry,
		},
	}
}
