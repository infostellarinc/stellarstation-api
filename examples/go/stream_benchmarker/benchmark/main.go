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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
	"github.com/infostellarinc/stellarstation-api/examples/go/benchmark"
)

type metric struct {
	timeFirstByteReceived time.Time
	timeLastByteReceived  time.Time
	dataSize              int
	metricTime            time.Time
}

type outputDetails struct {
	timenow                        time.Time
	mostRecentTimeLastByteReceived time.Time
	averageFirstByteLatency        time.Duration
	averageLastByteLatency         time.Duration
	totalDataSize                  int64
	metricsCount                   int64
	averageDataSize                float64
	mbps                           float64
	outOfOrderData                 int64
}

type sessionSummary struct {
	firstByteTime                time.Time
	lastByteTime                 time.Time
	totalAverageFirstByteLatency time.Duration
	totalAverageLastByteLatency  time.Duration
	totalBytes                   int64
	totalMbps                    float64
	numIntervals                 float64
	totalOutOfOrderData          int64
}

type metricsData struct {
	initialTime                    time.Time
	metrics                        []metric
	mostRecentTimeLastByteReceived time.Time
	timeOfReporting                time.Time
}

func (metricsData *metricsData) compileDetails() *outputDetails {
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

	for _, metric := range metricsData.metrics {
		var firstByteTime = metric.timeFirstByteReceived
		var lastByteTime = metric.timeLastByteReceived
		var dataSize = metric.dataSize
		var metricTime = metric.metricTime

		var firstByteLatency = metricTime.Sub(firstByteTime)
		var lastByteLatency = metricTime.Sub(lastByteTime)

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

	var mbps = (float64(totalDataSize) / float64((now.Sub(metricsData.initialTime).Seconds()))) * 8 / 1024 / 1024

	outputDetails := outputDetails{
		timenow:                        metricsData.timeOfReporting,
		mostRecentTimeLastByteReceived: metricsData.mostRecentTimeLastByteReceived,
		averageFirstByteLatency:        averageFirstByteLatency,
		averageLastByteLatency:         averageLastByteLatency,
		totalDataSize:                  totalDataSize,
		metricsCount:                   metricsCount,
		averageDataSize:                averageDataSize,
		mbps:                           mbps,
		outOfOrderData:                 outOfOrderData,
	}

	return &outputDetails
}

func printHeader() {
	header := "DATE\tMost recent received\tAvg seconds to last byte\tAvg seconds to first byte\tTotal bytes\tNum messages\tAvg bytes\tMbps\tOut of order count"
	fmt.Println(header)
}

func printDetails(outputDetails *outputDetails) {
	//outputDetails := metricsData.compileDetails()
	var b bytes.Buffer
	b.WriteString(outputDetails.timenow.Format("2006-01-02 15:04:05.0000") + "\t")
	b.WriteString(outputDetails.mostRecentTimeLastByteReceived.Format("2006-01-02 15:04:05.0000") + "\t")
	b.WriteString(fmt.Sprintf("%f", outputDetails.averageFirstByteLatency.Seconds()) + "\t")
	b.WriteString(fmt.Sprintf("%f", outputDetails.averageLastByteLatency.Seconds()) + "\t")
	b.WriteString(strconv.FormatInt(outputDetails.totalDataSize, 10) + "\t")
	b.WriteString(strconv.FormatInt(outputDetails.metricsCount, 10) + "\t")
	b.WriteString(fmt.Sprintf("%f", outputDetails.averageDataSize) + "\t")
	b.WriteString(fmt.Sprintf("%f", outputDetails.mbps) + "\t")
	b.WriteString(strconv.FormatInt(outputDetails.outOfOrderData, 10) + "\t")

	fmt.Println(b.String())
}

func printSummary(sessionSummary *sessionSummary) {
	var b bytes.Buffer
	b.WriteString("First Byte Time\t" + sessionSummary.firstByteTime.Format("2006-01-02 15:04:05.0000") + "\n")
	b.WriteString("Last Byte Time\t" + sessionSummary.lastByteTime.Format("2006-01-02 15:04:05.0000") + "\t")
	b.WriteString("Average First Byte Latency\t" + fmt.Sprintf("%f", sessionSummary.totalAverageFirstByteLatency.Seconds()/sessionSummary.numIntervals) + "\n")
	b.WriteString("Average Last Byte Latency\t" + fmt.Sprintf("%f", sessionSummary.totalAverageLastByteLatency.Seconds()/sessionSummary.numIntervals) + "\n")
	b.WriteString("Total Bytes\t" + strconv.FormatInt(sessionSummary.totalBytes, 10) + "\n")
	b.WriteString("Average Speed\t" + fmt.Sprintf("%f", sessionSummary.totalMbps/sessionSummary.numIntervals) + "\n")
	b.WriteString("Total Out Of Order Bytes\t" + strconv.FormatInt(sessionSummary.totalOutOfOrderData, 10) + "\n")

	fmt.Println(b.String())
}

func main() {

	apiKey := flag.String("key", "colin-alpha-key.json", "api key")
	endpoint := flag.String("endpoint", "api-alpha.stellarstation.com:443", "api endpoint")
	satelliteID := flag.String("id", "112", "satellite id")
	interval := flag.Int("interval", 10, "reporting interval")
	exitOnNoDataInSeconds := flag.Int("exitOnNoDataIn", -1, "number of seconds of no data received to auto exit program")
	printSessionSummary := flag.Bool("summary", true, "print a session summary")

	flag.Parse()

	satelliteStreamOptions := &benchmark.SatelliteStreamOptions{
		AcceptedFraming: []stellarstation.Framing{stellarstation.Framing_AX25, stellarstation.Framing_BITSTREAM},
		SatelliteID:     *satelliteID,
		APIKey:          *apiKey,
		Endpoint:        *endpoint,
	}

	var streamChannel = make(chan *stellarstation.Telemetry)

	ss, err := benchmark.OpenSatelliteStream(satelliteStreamOptions, streamChannel)
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	ticker := time.NewTicker(time.Second * time.Duration(*interval))
	defer ticker.Stop()

	fmt.Println("Listening for messages")
	var sessionSummary sessionSummary
	var md metricsData
	var storeMostRecentTimeLastByteReceived time.Time = time.Time{}
	var gotFirstMessage = false
	var done = false
	for !done {
		select {
		case message := <-streamChannel:
			if gotFirstMessage == false {
				fmt.Println("Receiving messages")
				printHeader()
				gotFirstMessage = true
				firstTime, _ := ptypes.Timestamp(message.TimeFirstByteReceived)
				md.initialTime = firstTime
				sessionSummary.firstByteTime = firstTime
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
			storeMostRecentTimeLastByteReceived = timeLastByteReceived
			md.mostRecentTimeLastByteReceived = storeMostRecentTimeLastByteReceived
		case timeOfReporting := <-ticker.C:
			if gotFirstMessage {
				md.timeOfReporting = timeOfReporting
				outputDetails := md.compileDetails()
				sessionSummary.totalBytes += outputDetails.totalDataSize
				sessionSummary.totalMbps += outputDetails.mbps
				sessionSummary.totalAverageFirstByteLatency += outputDetails.averageFirstByteLatency
				sessionSummary.totalAverageLastByteLatency += outputDetails.averageLastByteLatency
				sessionSummary.numIntervals++
				sessionSummary.lastByteTime = md.mostRecentTimeLastByteReceived
				sessionSummary.totalOutOfOrderData += outputDetails.outOfOrderData
				if (*exitOnNoDataInSeconds > 0) && (outputDetails.totalDataSize == 0) {
					secondsSinceFirstByte := time.Now().UTC().Sub(sessionSummary.firstByteTime)
					if secondsSinceFirstByte > time.Duration(*exitOnNoDataInSeconds) {
						done = true
					}
				}
				if !done {
					printDetails(outputDetails)
					md = metricsData{}
					md.initialTime = timeOfReporting
					md.mostRecentTimeLastByteReceived = storeMostRecentTimeLastByteReceived
				}
			}
		case <-signalChan:
			fmt.Println("\nReceived interrupt, stopping benchmark")
			done = true
		}
	}
	if *printSessionSummary {
		if sessionSummary.numIntervals > 0 {
			printSummary(&sessionSummary)
		}
	}
	ss.Close()
}
