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
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"

	v1 "github.com/infostellarinc/go-stellarstation/api/v1"
	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
	radio "github.com/infostellarinc/go-stellarstation/api/v1/radio"
)

/*********
 * Modem *
 *********/

// Modem does all of the actual work.
type Modem struct {
	config         *Config
	groundstations map[string]GroundStationConfig      // Indexed by name
	satellites     map[string]SatelliteConfig          // Indexed by ID
	channels       map[string]map[string]ChannelConfig // Indexed by satellite ID and channel ID
	runners        map[string]*Runner                  // Indexed by plan ID
	runnersLock    *sync.Mutex
	client         *Client
	planWatcher    *PlanWatcher
}

// NewModem creates a new Modem instance
func NewModem(config *Config) *Modem {
	modem := &Modem{
		config:         config,
		groundstations: make(map[string]GroundStationConfig),
		satellites:     make(map[string]SatelliteConfig),
		channels:       make(map[string]map[string]ChannelConfig),
		runners:        make(map[string]*Runner),
		runnersLock:    &sync.Mutex{},
		client:         NewClient(),
	}

	modem.planWatcher = NewPlanWatcher(modem.client)

	for _, gs := range config.GroundStations {
		modem.groundstations[gs.Name] = gs
	}

	for _, s := range config.Satellites {
		modem.satellites[s.ID] = s
		m := make(map[string]ChannelConfig)
		for _, c := range s.Channels {
			m[c.ID] = c
		}
		modem.channels[s.ID] = m
	}

	return modem
}

// ConnectToGroundStation connects to the requested ground station
// or the default ground station if the given name does not match
// any of the available ground station configurations.
func (m *Modem) ConnectToGroundStation(name string) {
	if m.config.GroundStations == nil || len(m.config.GroundStations) == 0 {
		log.Fatalf("No ground stations defined\n")
	}

	var gs GroundStationConfig
	found := false

	if name != "" {
		gs, found = m.groundstations[name]
		if !found {
			log.Printf("Couldn't find requested ground station: %s\n", name)
		}
	}

	if !found && m.config.Default != "" {
		name = m.config.Default
		log.Printf("Using default ground station: %s\n", name)

		gs, found = m.groundstations[name]
		if !found {
			log.Printf("Couldn't find default ground station: %s\n", name)
		}
	}

	if !found {
		name = m.config.GroundStations[0].Name
		log.Printf("Using first ground station: %s\n", name)

		gs, found = m.groundstations[name]
		if !found {
			// This should never happen.
			log.Printf("Couldn't find first ground station: %s\n", name)
		}
	}

	if !found {
		// This should never happen
		log.Fatalf("Ground station config not found.\n")
	}

	m.client.Connect(gs)

	if gs.PlanUpdateIntervalInMinutes == 0 {
		gs.PlanUpdateIntervalInMinutes = DefaultPlanUpdateIntervalInMinutes
	}

	planUpdateInterval := time.Minute * time.Duration(gs.PlanUpdateIntervalInMinutes)

	m.planWatcher.Start(planUpdateInterval, m.PlanStart, m.PlanEnd)
}

// Stop will stop the modem
func (m *Modem) Stop() {
	m.planWatcher.Stop()
	m.client.Stop()
}

// Wait will wait for the modem to stop before returning
func (m *Modem) Wait() {
	m.planWatcher.Wait()
	m.client.Wait()
}

// Client returns the API client used by this modem
func (m *Modem) Client() *Client {
	return m.client
}

// PlanWatcher returns the Plan Watcher used by this modem
func (m *Modem) PlanWatcher() *PlanWatcher {
	return m.planWatcher
}

// PlanStart is executed when a plan starts
func (m *Modem) PlanStart(plan *api.Plan) {
	m.runnersLock.Lock()
	defer m.runnersLock.Unlock()

	_, found := m.runners[plan.PlanId]
	if found {
		log.Printf("!!!!! Plan already running. %v\n", shortPlanData(plan))
		return
	}

	// Load data
	// TODO: Fix this when the plan has a satellite ID
	data, err := ioutil.ReadFile("data.bin")
	if err != nil {
		log.Printf("!!!!! Couldn't load data file. Plan ID: %v, Error: %v\n", plan.PlanId, err)
		return
	}

	log.Printf("~~~~~ Loaded %v bytes of data. Plan ID: %v\n", len(data), plan.PlanId)

	_, aos, los, _, err := parseTimestamps(plan)
	if err != nil {
		log.Printf("!!!!! Couldn't parse plan timestamps. Plan ID: %v, Error: %v\n", plan.PlanId, err)
	}

	startFunction := func(ctx context.Context) {
		log.Printf(">>>>> Plan started. %v\n", shortPlanData(plan))
		go func() {
			// Wait for AOS to start streaming
			time.Sleep(time.Until(aos))

			log.Printf("***** Plan AOS reached. %v\n", shortPlanData(plan))

			losTimer := time.NewTimer(time.Until(los))
			defer losTimer.Stop()

			// TODO: Adjust timing to match data rate
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			stream, err := m.client.OpenGroundStationStream(ctx)
			if err != nil {
				log.Printf("!!!!! Couldn't open stream. Plan ID: %v, Error: %v\n", plan.PlanId, err)
				return
			}

			log.Printf("||||| Stream opened.  Plan ID: %v\n", plan.PlanId)

			for {
				select {
				case <-ctx.Done():
					return
				case <-losTimer.C:
					log.Printf("..... Plan LOS reached. %v\n", shortPlanData(plan))
					return
				case <-ticker.C:
					request := m.TelemetryRequest(plan, data)
					err = stream.Send(request)
					if err != nil {
						log.Printf("!!!!! Couldn't send telemetry request. Plan ID: %v, Error: %v\n", plan.PlanId, err)
					}
					log.Printf("^^^^^ Sent %v bytes. Plan ID: %v\n", len(data), plan.PlanId)
				}
			}
		}()
	}

	stopFunction := func() {
		log.Printf("<<<<< Plan ended. %v\n", shortPlanData(plan))
	}

	runner := NewRunner()
	go runner.Start(startFunction, stopFunction)
	m.runners[plan.PlanId] = runner
}

// PlanEnd is executed when a plan ends.
func (m *Modem) PlanEnd(plan *api.Plan) {
	m.runnersLock.Lock()
	defer m.runnersLock.Unlock()

	runner, found := m.runners[plan.PlanId]
	if !found {
		log.Printf("!!!!! Plan not running. %v\n", shortPlanData(plan))
		return
	}
	go runner.Stop()
	delete(m.runners, plan.PlanId)
}

// TelemetryRequest creates a telemetry request objet for the given plan and data.
func (m *Modem) TelemetryRequest(plan *api.Plan, data []byte) *api.GroundStationStreamRequest {
	now := ptypes.TimestampNow()
	freq := plan.DownlinkRadioDevice.CenterFrequencyHz

	framing := v1.Framing_BITSTREAM

	switch plan.DownlinkRadioDevice.Protocol.Framing.(type) {
	case *radio.CommunicationProtocol_Ax25:
		// Set AX.25 framing if the satellite uses AX.25
		framing = v1.Framing_AX25
	}

	telemetry := &api.SatelliteTelemetry{
		PlanId: plan.PlanId,
		Telemetry: &v1.Telemetry{
			Data:                  data,
			Framing:               framing,
			TimeFirstByteReceived: now,
			TimeLastByteReceived:  now,
			DownlinkFrequencyHz:   freq,
		},
	}
	return m.client.TelemetryRequest(telemetry)
}
