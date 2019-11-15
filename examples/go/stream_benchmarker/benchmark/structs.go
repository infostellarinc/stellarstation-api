package main

import (
	"fmt"
	"io"
	"time"
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
	var averageDataSize float64

	var totalFirstByteLatency time.Duration
	var totalLastByteLatency time.Duration
	var totalDataSize int64

	var metricsCount int64
	var outOfOrderData int64
	var previousFirstByteTime = time.Time{}

	var planID string

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
