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
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/***************
 * Main Method *
 ***************/

func main() {
	var configFile, groundstation string

	flag.StringVar(&configFile, "c", "config.json", "config file")
	flag.StringVar(&groundstation, "g", "", "ground station (blank for default)")
	flag.Parse()

	config, err := LoadConfigFromJSON(configFile)
	if err != nil {
		log.Fatalf("Error loading config file: %v\n", err)
	}

	sender := NewSender(config, groundstation)

	sender.Start()
	defer sender.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		sender.Stop()
	}()

	sender.Wait()
}
