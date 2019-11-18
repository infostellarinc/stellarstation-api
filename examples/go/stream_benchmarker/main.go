// Copyright © 2019 Infostellar, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/ptypes"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarstation-api/examples/go/stream_benchmarker/benchmark"
)

func main() {
	apiKey := "stellarstation-private-key.json"
	endpoint := "api.stellarstation.com:443"
	satelliteID := "5"
	interval := 10 * time.Second
	assumeEndAfterDuration := 10 * time.Second
	exitAfterPassEnds := false
	willNotPrintPassSummary := false
	willNotPrintOverallSummary := false
	filename := ""

	flag.StringVar(&apiKey, "k", apiKey, "StellarStation API Key file")
	flag.StringVar(&endpoint, "E", endpoint, "API endpoint")
	flag.StringVar(&satelliteID, "s", satelliteID, "Satellite ID as provided by StellarStation")
	flag.DurationVar(&interval, "i", interval, "Reporting interval.  (10s, 1m, etc.)  During a pass, an output line will be generated for each reporting interval.")
	flag.DurationVar(&assumeEndAfterDuration, "e", assumeEndAfterDuration, "Assume a pass has ended after this much time has passed without receiving any additional data")
	flag.BoolVar(&exitAfterPassEnds, "x", exitAfterPassEnds, "Exit the program after a pass ends")
	flag.BoolVar(&willNotPrintPassSummary, "P", willNotPrintPassSummary, "Do not print a pass summary after each pass")
	flag.BoolVar(&willNotPrintOverallSummary, "S", willNotPrintOverallSummary, "Do not print an overall summary when the program exits")
	flag.StringVar(&filename, "o", filename, "Write report output to a file instead of standard out")

	flag.Parse()

	log.SetOutput(os.Stderr)
	output := os.Stdout

	if filename != "" {
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Couldn't open output file. File: %v, Error: %v\n", filename, err)
		}
		defer file.Close()
		output = file
	}

	satelliteStreamOptions := &benchmark.SatelliteStreamOptions{
		AcceptedFraming: []stellarstation.Framing{stellarstation.Framing_AX25, stellarstation.Framing_BITSTREAM},
		SatelliteID:     satelliteID,
		APIKey:          apiKey,
		Endpoint:        endpoint,
	}

	streamChannel := make(chan *stellarstation.ReceiveTelemetryResponse)

	ss, err := benchmark.OpenSatelliteStream(satelliteStreamOptions, streamChannel)
	if err != nil {
		log.Fatal(err)

	}
	defer ss.Close()
	done := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	reportingTicker := time.NewTicker(interval)
	defer reportingTicker.Stop()

	log.Println("Listening for messages")
	var ps passSummary
	var md metricsData
	var passSummaries []passSummary
	passStarted := false
	printedHeader := false
	for {
		select {
		case streamResponse := <-streamChannel:
			message := streamResponse.Telemetry
			if len(message.Data) != 0 {
				planID := streamResponse.PlanId
				if passStarted == false {
					if planID != "" {
						log.Printf("Receiving messages for Plan ID %v", planID)
						passStarted = true
						firstTime, _ := ptypes.Timestamp(message.TimeFirstByteReceived)
						md.initialTime = firstTime
						ps.firstByteTime = firstTime
						ps.initialPlanID = planID
					}

				}
				timeFirstByteReceived, _ := ptypes.Timestamp(message.TimeFirstByteReceived)
				timeLastByteReceived, _ := ptypes.Timestamp(message.TimeLastByteReceived)
				numBytesInMessage := len(message.Data)

				metric := metric{
					timeFirstByteReceived: timeFirstByteReceived,
					timeLastByteReceived:  timeLastByteReceived,
					dataSize:              numBytesInMessage,
					metricTime:            time.Now().UTC(),
				}
				md.metrics = append(md.metrics, metric)
				if md.mostRecentTimeLastByteReceived.Before(timeLastByteReceived) {
					md.mostRecentTimeLastByteReceived = timeLastByteReceived
				}
				md.planID = planID
			}
		case timeOfReporting := <-reportingTicker.C:
			if passStarted {
				md.timeOfReporting = timeOfReporting
				outputDetails := md.compileDetails()
				if !printedHeader {
					outputDetails.printHeader(output)
					printedHeader = true
				}
				if (assumeEndAfterDuration > 0) && (outputDetails.totalDataSize == 0) {
					timeSincePassFirstByte := time.Now().UTC().Sub(ps.firstByteTime)
					if timeSincePassFirstByte > assumeEndAfterDuration {
						log.Printf("Plan with ID %v ended after not receiving messages for %v", ps.initialPlanID, assumeEndAfterDuration)
						if !willNotPrintPassSummary {
							ps.print(output)
						}
						if exitAfterPassEnds {
							close(done)
						}
						passSummaries = append(passSummaries, ps)
						ps = passSummary{}
						passStarted = false
					}
				} else {
					ps.totalBytes += outputDetails.totalDataSize
					ps.totalMbps += outputDetails.dataRateInMbps
					ps.totalAverageFirstByteLatency += outputDetails.averageFirstByteLatency
					ps.totalAverageLastByteLatency += outputDetails.averageLastByteLatency
					ps.numIntervals++
					ps.lastByteTime = md.mostRecentTimeLastByteReceived
					ps.totalOutOfOrderData += outputDetails.outOfOrderData
				}

				if outputDetails.planID != "" {
					outputDetails.print(output)
					md = metricsData{
						mostRecentTimeLastByteReceived: md.mostRecentTimeLastByteReceived,
						initialTime:                    timeOfReporting,
					}
				}
			}
		case <-signalChan:
			log.Println("Received interrupt, stopping benchmark")
			if !willNotPrintPassSummary && ps.initialPlanID != "" {
				passSummaries = append(passSummaries, ps)
				ps.print(output)
			}
			close(done)
		case <-done:
			if !willNotPrintOverallSummary {
				if len(passSummaries) > 0 {
					sessionSummary := calculateSessionSummary(passSummaries)
					sessionSummary.print(output)
				}
			}
			log.Println("Session ended")
			return
		}
	}
}
