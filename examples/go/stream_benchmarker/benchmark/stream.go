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

package benchmark

import (
	"context"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff"
	"google.golang.org/grpc"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
)

const (
	open   uint32 = 0
	closed uint32 = 1
)

const maxElapsedTime = 60 * time.Second

// SatelliteStreamOptions contains options for opening a satellite stream
type SatelliteStreamOptions struct {
	AcceptedFraming []stellarstation.Framing

	SatelliteID string
	APIKey      string
	Endpoint    string
}

// SatelliteStream is the interface for the stream
type SatelliteStream interface {
	Send(payload []byte) error

	io.Closer
}

type satelliteStream struct {
	satelliteID string
	apiKey      string
	endpoint    string

	acceptedFraming []stellarstation.Framing

	stream   stellarstation.StellarStationService_OpenSatelliteStreamClient
	conn     *grpc.ClientConn
	streamID string

	recvChan           chan<- *stellarstation.ReceiveTelemetryResponse
	recvLoopClosedChan chan struct{}

	state uint32
}

type streamResponses struct {
	streamEventResponses      []*stellarstation.SatelliteStreamResponse_StreamEvent
	receiveTelemetryResponses []*stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse
}

// OpenSatelliteStream opens a stream to a satellite over the StellarStation API.
func OpenSatelliteStream(o *SatelliteStreamOptions, recvChan chan<- *stellarstation.ReceiveTelemetryResponse) (SatelliteStream, error) {
	s := &satelliteStream{
		acceptedFraming: o.AcceptedFraming,
		satelliteID:     o.SatelliteID,
		apiKey:          o.APIKey,
		endpoint:        o.Endpoint,

		streamID:           "",
		recvChan:           recvChan,
		state:              open,
		recvLoopClosedChan: make(chan struct{}),
	}

	err := s.start()

	return s, err
}

// Send sends a packet to the satellite.
func (ss *satelliteStream) Send(payload []byte) error {
	req := stellarstation.SatelliteStreamRequest{
		SatelliteId: ss.satelliteID,
		Request: &stellarstation.SatelliteStreamRequest_SendSatelliteCommandsRequest{
			SendSatelliteCommandsRequest: &stellarstation.SendSatelliteCommandsRequest{
				Command: [][]byte{payload},
			},
		},
	}

	return ss.stream.Send(&req)
}

// Close closes the stream.
func (ss *satelliteStream) Close() error {
	atomic.StoreUint32(&ss.state, closed)

	ss.stream.CloseSend()
	ss.conn.Close()

	<-ss.recvLoopClosedChan

	return nil
}

func (ss *satelliteStream) recvLoop() {
	// Initialize exponential back off settings.
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxElapsedTime

	for {
		res, err := ss.stream.Recv()
		if atomic.LoadUint32(&ss.state) == closed {
			// Closed, so just shutdown the loop.
			close(ss.recvLoopClosedChan)
			return
		}
		if err != nil {
			log.Println(err)
			log.Println("reconnecting to the API stream.")

			rcErr := backoff.RetryNotify(func() error {
				err := ss.openStream()
				if err != nil {
					return err
				}

				response, err := ss.stream.Recv()
				if err != nil {
					return err
				}
				res = response

				return nil
			}, b,
				func(e error, duration time.Duration) {
					log.Printf("%s. Automatically retrying in %v", e, duration)
				})
			if rcErr != nil {
				// Couldn't reconnect to the server, bailout.
				log.Fatalf("error connecting to API stream: %v\n", err)
			}
			log.Println("connected to the API stream.")
		}
		ss.streamID = res.StreamId

		switch res.Response.(type) {
		case *stellarstation.SatelliteStreamResponse_ReceiveTelemetryResponse:

			telemetry := res.GetReceiveTelemetryResponse()
			payload := telemetry
			ss.recvChan <- payload
		}
	}
}

func (ss *satelliteStream) openStream() error {
	conn, err := Dial(ss.apiKey, ss.endpoint)
	if err != nil {
		return err
	}

	client := stellarstation.NewStellarStationServiceClient(conn)

	stream, err := client.OpenSatelliteStream(context.Background())
	if err != nil {
		conn.Close()
		return err
	}

	req := stellarstation.SatelliteStreamRequest{
		AcceptedFraming: ss.acceptedFraming,
		SatelliteId:     ss.satelliteID,
		StreamId:        ss.streamID,
	}

	err = stream.Send(&req)
	if err != nil {
		conn.Close()
		return err
	}

	ss.conn = conn
	ss.stream = stream
	return nil
}

func (ss *satelliteStream) start() error {
	err := ss.openStream()
	if err != nil {
		return err
	}
	go ss.recvLoop()

	return nil
}
