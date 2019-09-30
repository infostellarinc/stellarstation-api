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
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

/***************
 * API Client *
 ***************/

// Client is a StellarStation Ground Station API client.
// It also periodically checks for new plans.
type Client struct {
	groundstation GroundStationConfig // Currently selected ground station config
	done          chan struct{}
	running       bool
	runningLock   *sync.Mutex
	client        api.GroundStationServiceClient
	plans         map[string]*api.Plan
	plansLock     *sync.Mutex
}

// NewClient creates a new Client instance
func NewClient() *Client {
	client := &Client{
		running:     false,
		runningLock: &sync.Mutex{},
		plans:       make(map[string]*api.Plan),
		plansLock:   &sync.Mutex{},
	}
	return client
}

// Connect connects to the ground station and begins checking for plans.
func (c *Client) Connect(groundstation GroundStationConfig, planCheckInterval time.Duration) {
	// First disconnect any active connection
	c.Stop()
	c.Wait()

	// Now start the new connection
	c.groundstation = groundstation

	c.runningLock.Lock()
	defer func() {
		c.running = true
		c.runningLock.Unlock()
	}()

	c.done = make(chan struct{})

	go func() {
		<-c.done
		log.Printf("Shutting down API client for %v...\n", groundstation.Name)
	}()

	log.Printf("Starting API client for %v...\n", groundstation.Name)
	log.Printf("Ground station configuration: %+v\n", groundstation)

	c.connect()

	go c.UpdatePlans()
	updatePlans := time.NewTicker(planCheckInterval)

	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-updatePlans.C:
				go c.UpdatePlans()
			}
		}
	}()
}

// Stop shuts down the client
func (c *Client) Stop() {
	c.runningLock.Lock()
	defer func() {
		c.running = false
		c.runningLock.Unlock()
	}()

	if c.running {
		close(c.done)
	}
}

// Wait will wait for the client to stop before returning
func (c *Client) Wait() {
	c.runningLock.Lock()
	running := c.running
	done := c.done
	c.runningLock.Unlock()

	if !running {
		return
	}
	<-done
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
		<-c.done
		conn.Close()
	}()

	c.client = api.NewGroundStationServiceClient(conn)
}

// UpdatePlans updates the plan list for the current ground station
func (c *Client) UpdatePlans() {
	c.plansLock.Lock()
	defer c.plansLock.Unlock()
	plans, err := c.ListPlans(c.groundstation.ID)
	if err != nil {
		log.Printf("Failed to list plans: %v\n", err)
		return
	}
	for id := range c.plans {
		delete(c.plans, id)
	}
	for _, plan := range plans {
		c.plans[plan.PlanId] = plan
	}
}

// ListPlans gets upcoming plans for the given ground station
func (c *Client) ListPlans(groundstationId string) ([]*api.Plan, error) {
	now := time.Now()
	end := now.Add(time.Hour)

	nowTs, _ := ptypes.TimestampProto(now)
	endTs, _ := ptypes.TimestampProto(end)

	listPlansRequest := &api.ListPlansRequest{
		GroundStationId: groundstationId,
		AosAfter:        nowTs,
		AosBefore:       endTs,
	}

	log.Printf("ListPlans Request: %+v\n", listPlansRequest)

	listPlansResponse, err := c.client.ListPlans(context.Background(), listPlansRequest)
	if err != nil {
		return nil, err
	}

	log.Printf("ListPlans Response: %+v\n", listPlansResponse)

	return listPlansResponse.Plan, nil
}
