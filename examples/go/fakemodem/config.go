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
	"encoding/json"
	"io/ioutil"
	"log"
)

/*****************
 * Configuration *
 *****************/

var DefaultPlanUpdateIntervalInMinutes = 5

// ChannelConfig defines the configuration for a specific
// communications channel.
type ChannelConfig struct {
	// Name should be a descriptive name.
	Name string
	// ID should match the communication channel ID
	// provided by StellarStation.
	ID string
	// Telemetry should be the filename of the satellite
	// telemetry to send when a pass is running.
	Telemetry string
}

// SatelliteConfig defines the configuration for a single satellite.
type SatelliteConfig struct {
	// Name should be a descriptive name.
	Name string
	// ID should match the satellite ID
	// provided by StellarStation.
	ID string
	// Channels is a list of communication channel configurations.
	Channels []ChannelConfig
}

// GroundStationConfig defines the configuration for
// the ground station.
type GroundStationConfig struct {
	// Name should be a descriptive name.
	Name string
	// ID should match the ground station ID
	// provided by StellarStation.
	ID string
	// Address is the URL of the ground station.
	Address string
	// Key should be the filename of the API key to use.
	Key string
	// PlanUpdateIntervalInMinutes is the number of minutes between plan update checks.
	PlanUpdateIntervalInMinutes int
}

// Config contains all of the configuration for the application.
type Config struct {
	// Default contains the name of the default ground station.
	Default string
	// GroundStations contains the groundstation configurations.
	GroundStations []GroundStationConfig
	// Satellites contains the satellite configurations.
	Satellites []SatelliteConfig
}

// LoadConfigFromJSON loads a configuration file from JSON.
func LoadConfigFromJSON(configFile string) (config *Config, err error) {
	log.Printf("Loading config file: %s\n", configFile)

	var data []byte

	data, err = ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config = &Config{}

	err = json.Unmarshal(data, config)

	return config, err
}
