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
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/ptypes"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarstation-api/examples/go/benchmark"
)

const dateTimeFormat = "2006-01-02 15:04:05"

type metric struct {
	timeFirstByteReceived time.Time
	timeLastByteReceived  time.Time
	dataSize              int
	metricTime            time.Time
}

type outputDetails struct {
	now                            time.Time
	mostRecentTimeLastByteReceived time.Time
	averageFirstByteLatency        time.Duration
	averageLastByteLatency         time.Duration
	totalDataSize                  int64
	metricsCount                   int64
	averageDataSize                float64
	dataRateInMbps                 float64
	outOfOrderData                 int64
	planID                         string
}

type passSummary struct {
	firstByteTime                time.Time
	lastByteTime                 time.Time
	totalAverageFirstByteLatency time.Duration
	totalAverageLastByteLatency  time.Duration
	totalBytes                   int64
	totalMbps                    float64
	numIntervals                 float64
	totalOutOfOrderData          int64
	initialPlanID                string
}

type sessionSummary struct {
	startOfSession         time.Time
	endOfSession           time.Time
	totalSessionBytes      int64
	totalSessionMbps       float64
	totalSessionOutOfOrder int64
	numberOfPlansProcessed int
}

type metricsData struct {
	initialTime                    time.Time
	metrics                        []metric
	mostRecentTimeLastByteReceived time.Time
	timeOfReporting                time.Time
	planID                         string
}

func (metricsData *metricsData) compileDetails() outputDetails {
	now := time.Now().UTC()

	var averageFirstByteLatency time.Duration
	var averageLastByteLatency time.Duration
	var averageDataSize float64 = 0

	var totalFirstByteLatency time.Duration
	var totalLastByteLatency time.Duration
	var totalDataSize int64 = 0

	var metricsCount int64 = 0
	var outOfOrderData int64 = 0
	var previousFirstByteTime = time.Time{}

	var planID string = ""

	for _, metric := range metricsData.metrics {
		firstByteTime := metric.timeFirstByteReceived
		lastByteTime := metric.timeLastByteReceived
		dataSize := metric.dataSize
		metricTime := metric.metricTime

		firstByteLatency := metricTime.Sub(firstByteTime)
		lastByteLatency := metricTime.Sub(lastByteTime)

		totalFirstByteLatency += firstByteLatency
		totalLastByteLatency += lastByteLatency
		totalDataSize += int64(dataSize)
		metricsCount++

		if metric.timeLastByteReceived.After(metricsData.mostRecentTimeLastByteReceived) {
			metricsData.mostRecentTimeLastByteReceived = metric.timeLastByteReceived
		}

		if previousFirstByteTime.After(firstByteTime) {
			outOfOrderData++
		}

		previousFirstByteTime = firstByteTime
	}

	if metricsCount > 0 {
		averageFirstByteLatency = totalFirstByteLatency / time.Duration(metricsCount)
		averageLastByteLatency = totalLastByteLatency / time.Duration(metricsCount)
		averageDataSize = float64(totalDataSize) / float64(metricsCount)
	}

	dataRateInMbps := (float64(totalDataSize) / (now.Sub(metricsData.initialTime).Seconds())) * 8 / 1024 / 1024

	planID = metricsData.planID

	outputDetails := outputDetails{
		now:                            metricsData.timeOfReporting,
		mostRecentTimeLastByteReceived: metricsData.mostRecentTimeLastByteReceived,
		averageFirstByteLatency:        averageFirstByteLatency,
		averageLastByteLatency:         averageLastByteLatency,
		totalDataSize:                  totalDataSize,
		metricsCount:                   metricsCount,
		averageDataSize:                averageDataSize,
		dataRateInMbps:                 dataRateInMbps,
		outOfOrderData:                 outOfOrderData,
		planID:                         planID,
	}

	return outputDetails
}

func (outputDetails *outputDetails) printHeader(output io.Writer) {
	fmt.Fprintf(output, "Pass %v:\n%10v%25v%25v%20v%18v%18v%15v%16v%11v%20v",
		outputDetails.planID,
		"PlanID",
		"DATE",
		"Most recent",
		"First byte latency",
		"Last byte latency",
		"Total bytes",
		"Num messages",
		"Avg bytes",
		"Mbps",
		"Out of order\n")
}

func (outputDetails *outputDetails) print(output io.Writer) {
	fmt.Fprintf(output, "%10v%25v%25v%20.2f%18.2f%18v%15v%16.2f%11.2f%20v\n",
		outputDetails.planID,
		outputDetails.now.Format(dateTimeFormat),
		outputDetails.mostRecentTimeLastByteReceived.Format(dateTimeFormat),
		outputDetails.averageFirstByteLatency.Seconds(),
		outputDetails.averageLastByteLatency.Seconds(),
		outputDetails.totalDataSize,
		outputDetails.metricsCount,
		outputDetails.averageDataSize,
		outputDetails.dataRateInMbps,
		outputDetails.outOfOrderData)
}

func (ps *passSummary) print(output io.Writer) {
	fmt.Fprintf(output, "\nPass Summary:\n%10v%25v%25v%20v%20v%20v%15v%15v\n",
		"PlanID",
		"First byte time",
		"Last byte time",
		"First byte batency",
		"Last byte latency",
		"Total bytes",
		"Average Mbps",
		"Out of order")

	averageFirstByteLatency := ps.totalAverageFirstByteLatency.Seconds() / ps.numIntervals
	averageLastByteLatency := ps.totalAverageLastByteLatency.Seconds() / ps.numIntervals

	fmt.Fprintf(output, "%10v%25v%25v%20.2f%20.2f%20v%15.2f%15v\n\n",
		ps.initialPlanID,
		ps.firstByteTime.Format(dateTimeFormat),
		ps.lastByteTime.Format(dateTimeFormat),
		averageFirstByteLatency,
		averageLastByteLatency,
		ps.totalBytes,
		ps.totalMbps/ps.numIntervals,
		ps.totalOutOfOrderData)
}

func (sessionSummary *sessionSummary) print(output io.Writer) {
	fmt.Fprintf(output, "Overall Summary:\n%22v%25v%15v%20v%15v%15v\n",
		"Start session",
		"End session",
		"Num passes",
		"Total bytes",
		"Average Mbps",
		"Out of order")

	fmt.Fprintf(output, "%22v%25v%15v%20v%15.2f%15v\n",
		sessionSummary.startOfSession.Format(dateTimeFormat),
		sessionSummary.endOfSession.Format(dateTimeFormat),
		sessionSummary.numberOfPlansProcessed,
		sessionSummary.totalSessionBytes,
		sessionSummary.totalSessionMbps/float64(sessionSummary.numberOfPlansProcessed),
		sessionSummary.totalSessionOutOfOrder)

}

func calculateSessionSummary(passSummaries []passSummary) sessionSummary {
	var startOfSession time.Time
	var endOfSession time.Time
	var totalSessionBytes int64
	var totalSessionMbps float64
	var totalSessionOutOfOrder int64
	var numberOfPlansProcessed int

	for i := 0; i < len(passSummaries); i++ {
		if i == 0 {
			startOfSession = passSummaries[i].firstByteTime
		}
		if i == len(passSummaries)-1 {
			endOfSession = passSummaries[i].lastByteTime
		}
		totalSessionBytes += passSummaries[i].totalBytes
		totalSessionMbps += passSummaries[i].totalMbps / passSummaries[i].numIntervals
		totalSessionOutOfOrder += passSummaries[i].totalOutOfOrderData
	}

	numberOfPlansProcessed = len(passSummaries)

	sessionSummary := sessionSummary{
		startOfSession:         startOfSession,
		endOfSession:           endOfSession,
		totalSessionBytes:      totalSessionBytes,
		totalSessionMbps:       totalSessionMbps,
		totalSessionOutOfOrder: totalSessionOutOfOrder,
		numberOfPlansProcessed: numberOfPlansProcessed,
	}

	return sessionSummary
}

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

	var streamChannel = make(chan *stellarstation.ReceiveTelemetryResponse)

	ss, err := benchmark.OpenSatelliteStream(satelliteStreamOptions, streamChannel)
	if err != nil {
		log.Fatal(err)

	}
	defer ss.Close()

	passEndedChan := make(chan bool)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	reportingTicker := time.NewTicker(interval)
	defer reportingTicker.Stop()

	log.Println("Listening for messages")
	var ps passSummary
	var md metricsData
	var passSummaries []passSummary
	var passStarted = false
	var done = false
	var printedHeader = false
	for !done {
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
						go func() { passEndedChan <- true }()
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

				if !done && outputDetails.planID != "" {
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

			done = true
		case <-passEndedChan:
			log.Printf("Plan with ID %v ended after not receiving messages for %v", ps.initialPlanID, assumeEndAfterDuration)
			if !willNotPrintPassSummary {
				ps.print(output)
			}
			if exitAfterPassEnds {
				done = true
			}
			passSummaries = append(passSummaries, ps)
			ps = passSummary{}
			passStarted = false
		}
	}

	if !willNotPrintOverallSummary {
		if len(passSummaries) > 0 {
			sessionSummary := calculateSessionSummary(passSummaries)
			sessionSummary.print(output)
		}
	}
	log.Println("Session ended")
}
