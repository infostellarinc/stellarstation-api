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

const dateTimeFormat = "2006-01-02 15:04:05.0000"

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

func printHeader(output io.Writer) {
	fmt.Fprintf(output, "PlanID\tDATE\tMost recent received\tAvg seconds to last byte\tAvg seconds to first byte\tTotal bytes\tNum messages\tAvg bytes\tMbps\tOut of order count\n")
}

func printPassDetails(output io.Writer, outputDetails outputDetails) {
	fmt.Fprintf(output, "%v\t%v\t%v\t%0.6f\t%0.6f\t%v\t%v\t%0.6f\t%0.6f\t%v\n",
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

func printPassSummary(output io.Writer, ps passSummary) {
	fmt.Fprintln(output, "Pass Summary:\nPlanID\tFirst Byte Received At\tLast Byte Received At\tAverage First Byte Latency\tAverage Last Byte Latency\tTotal Bytes Received\tAverage Mbps\tTotal Out of Order Bytes")

	averageFirstByteLatency := ps.totalAverageFirstByteLatency.Seconds() / ps.numIntervals
	averageLastByteLatency := ps.totalAverageLastByteLatency.Seconds() / ps.numIntervals

	fmt.Fprintf(output, "%v\t%v\t%v\t%0.6f\t%0.6f\t%v\t%0.6f\t%v\n",
		ps.initialPlanID,
		ps.firstByteTime.Format(dateTimeFormat),
		ps.lastByteTime.Format(dateTimeFormat),
		averageFirstByteLatency,
		averageLastByteLatency,
		ps.totalBytes,
		ps.totalMbps/ps.numIntervals,
		ps.totalOutOfOrderData)
}

func (sessionSummary *sessionSummary) printSessionSummary(output io.Writer) {
	fmt.Fprintln(output, "Overall Summary:\nStart of session\tEnd of session\tNumber of Passes\tTotal Bytes Received\tAverage Mbps\tTotal Out Of Order")

	fmt.Fprintf(output, "%v\t%v\t%v\t%v\t%0.6f\t%v\n",
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

	apiKey := flag.String("privateKeyFile", "stellarstation-private-key.json", "api key")
	endpoint := flag.String("apiEndpoint", "api.stellarstation.com:443", "api endpoint")
	satelliteID := flag.String("satelliteID", "5", "satellite id")
	interval := flag.Duration("reportingInterval", 10*time.Second, "reporting interval")
	assumeEndAfterSeconds := flag.Duration("assumeEndAfter", 10*time.Second, "assume the pass has ended after X duration of no data")
	willPrintPassSummary := flag.Bool("passSummary", true, "print a pass summary after every pass")
	willPrintOverallSummary := flag.Bool("sessionSummary", true, "print an overall summary for all passes")
	exitAfterPassEnds := flag.Bool("exitAfterPassEnds", false, "exit the program after a pass ends")
	filename := flag.String("outputFile", "", "specify output file and benchmark will output to it instead of standard out")

	log.SetOutput(os.Stderr)
	output := os.Stdout

	flag.Parse()

	if *filename != "" {
		file, err := os.Create(*filename)
		if err != nil {
			log.Fatalf("Couldn't open output file. File: %v, Error: %v\n", filename, err)
		}
		defer file.Close()
		output = file
	}

	satelliteStreamOptions := &benchmark.SatelliteStreamOptions{
		AcceptedFraming: []stellarstation.Framing{stellarstation.Framing_AX25, stellarstation.Framing_BITSTREAM},
		SatelliteID:     *satelliteID,
		APIKey:          *apiKey,
		Endpoint:        *endpoint,
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
	reportingTicker := time.NewTicker(*interval)
	defer reportingTicker.Stop()

	log.Println("Listening for messages")
	var ps passSummary
	var md metricsData
	var passSummaries []passSummary
	var passStarted = false
	var done = false
	for !done {
		select {
		case streamResponse := <-streamChannel:
			message := streamResponse.Telemetry
			if len(message.Data) != 0 {
				planID := streamResponse.PlanId
				if passStarted == false {
					if planID != "" {
						log.Println("Receiving messages")
						printHeader(output)
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
				if (*assumeEndAfterSeconds > 0) && (outputDetails.totalDataSize == 0) {
					secondsSincePassFirstByte := time.Now().UTC().Sub(ps.firstByteTime)
					if secondsSincePassFirstByte > *assumeEndAfterSeconds {
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
					printPassDetails(output, outputDetails)
					md = metricsData{
						mostRecentTimeLastByteReceived: md.mostRecentTimeLastByteReceived,
						initialTime:                    timeOfReporting,
					}
				}
			}
		case <-signalChan:
			log.Println("Received interrupt, stopping benchmark")
			if ps.initialPlanID != "" {
				passSummaries = append(passSummaries, ps)
				printPassSummary(output, ps)
			}

			done = true
		case <-passEndedChan:
			log.Printf("Plan with ID %v ended after not receiving data for %v", ps.initialPlanID, *assumeEndAfterSeconds)
			if *willPrintPassSummary {
				printPassSummary(output, ps)
			}
			if *exitAfterPassEnds {
				done = true
			}
			passSummaries = append(passSummaries, ps)
			ps = passSummary{}
			passStarted = false
		}
	}

	if *willPrintOverallSummary {
		if len(passSummaries) > 0 {
			sessionSummary := calculateSessionSummary(passSummaries)
			sessionSummary.printSessionSummary(output)
		}
	}
	log.Println("Session ended")
}
