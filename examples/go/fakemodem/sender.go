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
 * Data Sender *
 ***************/

// Sender does all of the actual work.
type Sender struct {
	config         *Config
	groundstations map[string]GroundStationConfig      // Indexed by name
	satellites     map[string]SatelliteConfig          // Indexed by ID
	channels       map[string]map[string]ChannelConfig // Indexed by satellite ID and channel ID
	groundstation  GroundStationConfig                 // Currently selected ground station config
	done           chan struct{}
	running        bool
	runningLock    *sync.Mutex
	client         api.GroundStationServiceClient
	plans          map[string]*api.Plan
	plansLock      *sync.Mutex
}

// NewSender creates a new Sender instance
func NewSender(config *Config, groundstation string) *Sender {
	sender := &Sender{
		config:         config,
		groundstations: make(map[string]GroundStationConfig),
		satellites:     make(map[string]SatelliteConfig),
		channels:       make(map[string]map[string]ChannelConfig),
		running:        false,
		runningLock:    &sync.Mutex{},
		plans:          make(map[string]*api.Plan),
		plansLock:      &sync.Mutex{},
	}

	for _, gs := range config.GroundStations {
		sender.groundstations[gs.Name] = gs
	}

	for _, s := range config.Satellites {
		sender.satellites[s.ID] = s
		m := make(map[string]ChannelConfig)
		for _, c := range s.Channels {
			m[c.ID] = c
		}
		sender.channels[s.ID] = m
	}

	sender.SelectGroundStation(groundstation)

	return sender
}

// SelectGroundStation selects the requested ground station config
// or the default configuration if the given name does not match
// any of the available ground station configurations.
// This method does not do anything if the sender is already running.
func (s *Sender) SelectGroundStation(name string) {
	s.runningLock.Lock()
	defer s.runningLock.Unlock()

	if s.running {
		log.Printf("Can't select new ground station while running.\n")
		return
	}

	if s.config.GroundStations == nil || len(s.config.GroundStations) == 0 {
		log.Fatalf("No ground stations defined\n")
	}

	var gs GroundStationConfig
	found := false

	if name != "" {
		gs, found = s.groundstations[name]
		if !found {
			log.Printf("Couldn't find requested ground station: %s\n", name)
		}
	}

	if !found && s.config.Default != "" {
		name = s.config.Default
		log.Printf("Using default ground station: %s\n", name)

		gs, found = s.groundstations[name]
		if !found {
			log.Printf("Couldn't find default ground station: %s\n", name)
		}
	}

	if !found {
		name = s.config.GroundStations[0].Name
		log.Printf("Using first ground station: %s\n", name)

		gs, found = s.groundstations[name]
		if !found {
			// This should never happen.
			log.Printf("Couldn't find first ground station: %s\n", name)
		}
	}

	if !found {
		// This should never happen
		log.Fatalf("Ground station config not found.\n")
	}

	s.groundstation = gs

}

// Start begins execution of the data sender.
func (s *Sender) Start() {
	s.runningLock.Lock()
	defer func() {
		s.running = true
		s.runningLock.Unlock()
	}()

	s.done = make(chan struct{})

	go func() {
		<-s.done
		log.Printf("Shutting down...")
	}()

	log.Printf("Starting...")
	log.Printf("Ground station: %+v\n", s.groundstation)
	log.Printf("Satellites: %+v\n", s.config.Satellites)

	s.connect()

	go s.UpdatePlans()
	updatePlans := time.NewTicker(time.Minute * 5)

	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-updatePlans.C:
				go s.UpdatePlans()
			}
		}
	}()
}

// Stop shuts down the data sender.
func (s *Sender) Stop() {
	s.runningLock.Lock()
	defer func() {
		s.running = false
		s.runningLock.Unlock()
	}()

	if s.running {
		close(s.done)
	}
}

// Wait will wait for the sender to stop before returning
func (s *Sender) Wait() {
	s.runningLock.Lock()
	running := s.running
	done := s.done
	s.runningLock.Unlock()

	if !running {
		return
	}
	<-done
}

/*************
 * API Calls *
 *************/

// connect connects to the selected ground station
func (s *Sender) connect() {

	jwtCreds, err := oauth.NewJWTAccessFromFile(
		s.groundstation.Key)
	if err != nil {
		log.Fatalf("Failed to create JWT credentials: %v", err)
	}

	tc := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	conn, err := grpc.Dial(s.groundstation.Address,
		grpc.WithTransportCredentials(tc),
		grpc.WithPerRPCCredentials(jwtCreds))

	if err != nil {
		log.Fatalf("Couldn't connect to ground station: %v\n", err)
	}

	go func() {
		<-s.done
		conn.Close()
	}()

	s.client = api.NewGroundStationServiceClient(conn)
}

// UpdatePlans updates the plan list for the current ground station
func (s *Sender) UpdatePlans() {
	s.plansLock.Lock()
	defer s.plansLock.Unlock()
	plans, err := s.ListPlans(s.groundstation.ID)
	if err != nil {
		log.Printf("Failed to list plans: %v\n", err)
		return
	}
	for id := range s.plans {
		delete(s.plans, id)
	}
	for _, plan := range plans {
		s.plans[plan.PlanId] = plan
	}
}

// ListPlans gets upcoming plans for the given ground station
func (s *Sender) ListPlans(groundstationId string) ([]*api.Plan, error) {
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

	listPlansResponse, err := s.client.ListPlans(context.Background(), listPlansRequest)
	if err != nil {
		return nil, err
	}

	log.Printf("ListPlans Response: %+v\n", listPlansResponse)

	return listPlansResponse.Plan, nil
}
