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

package com.stellarstation.api.test.satellite;

import static org.assertj.core.api.Assertions.assertThat;

import com.google.common.collect.ImmutableList;
import com.google.common.util.concurrent.Futures;
import com.google.common.util.concurrent.SettableFuture;
import com.google.protobuf.ByteString;
import com.stellarstation.api.test.auth.ApiClientModule;
import com.stellarstation.api.v1.SatelliteStreamRequest;
import com.stellarstation.api.v1.SatelliteStreamResponse;
import com.stellarstation.api.v1.SendSatelliteCommandsRequest;
import com.stellarstation.api.v1.monitoring.AntennaState;
import dagger.Component;
import io.grpc.stub.StreamObserver;
import java.util.ArrayList;
import java.util.List;
import javax.inject.Singleton;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class SatelliteStreamerTest {
  private SatelliteStreamer streamer;

  private static final String SATELLITE_ID = "98";

  @BeforeEach
  void setUp() {
    streamer = DaggerSatelliteStreamerTest_SatelliteStreamerComponent.create().streamer();
  }

  @Component(modules = ApiClientModule.class)
  @Singleton
  interface SatelliteStreamerComponent {
    SatelliteStreamer streamer();
  }

  @Test
  void telemetryAndCommand() throws Exception {
    final SettableFuture<Boolean> testCompletionFuture = SettableFuture.create();
    final SettableFuture<StreamObserver<SatelliteStreamRequest>> requestObserverFuture =
        SettableFuture.create();

    class TelemetryAndCommandTestStreamObserver implements StreamObserver<SatelliteStreamResponse> {
      private final List<Integer> safeModeStates = new ArrayList();
      private int telemetryReceived;
      private int totalResponseReceived;

      @Override
      public void onNext(SatelliteStreamResponse response) {
        if (response.hasReceiveTelemetryResponse()) {
          totalResponseReceived++;
          if (totalResponseReceived > 30) {
            onCompleted();
          }

          ByteString data = response.getReceiveTelemetryResponse().getTelemetry().getData();
          if (data.size() > 2) {
            // The second last byte of the telemetry indicates the current state of the
            // fake satellite used in the test. The value is either of 0 or 1.
            int state = data.byteAt(data.size() - 2);
            safeModeStates.add(state);

            // Send a command to toggle the state.
            SendSatelliteCommandsRequest commandsRequest =
                SendSatelliteCommandsRequest.newBuilder()
                    .addCommand(ByteString.copyFrom(new byte[] {0x01, 0x01}))
                    .build();
            Futures.getUnchecked(requestObserverFuture)
                .onNext(
                    SatelliteStreamRequest.newBuilder()
                        .setSatelliteId(SATELLITE_ID)
                        .setSendSatelliteCommandsRequest(commandsRequest)
                        .build());

            telemetryReceived++;
            if (telemetryReceived == 2) {
              onCompleted();
            }
          }
        }
      }

      @Override
      public void onError(Throwable t) {
        testCompletionFuture.setException(t);
      }

      @Override
      public void onCompleted() {
        testCompletionFuture.set(true);
      }

      public List<Integer> getSafeModeStates() {
        return ImmutableList.copyOf(safeModeStates);
      }
    }

    TelemetryAndCommandTestStreamObserver responseObserver =
        new TelemetryAndCommandTestStreamObserver();
    StreamObserver<SatelliteStreamRequest> requestObserver = streamer.openStream(responseObserver);
    requestObserverFuture.set(requestObserver);

    requestObserver.onNext(
        SatelliteStreamRequest.newBuilder().setSatelliteId(SATELLITE_ID).build());

    assertThat(testCompletionFuture.get()).isTrue();
    assertThat(responseObserver.getSafeModeStates()).containsExactlyInAnyOrder(0, 1);
  }

  @Test
  void events() throws Exception {
    final SettableFuture<Boolean> testCompletionFuture = SettableFuture.create();

    class EventTestStreamObserver implements StreamObserver<SatelliteStreamResponse> {
      private AntennaState antennaState;
      private int totalResponseReceived;

      @Override
      public void onNext(SatelliteStreamResponse response) {
        totalResponseReceived++;
        if (totalResponseReceived > 10) {
          onCompleted();
        }

        if (response.hasStreamEvent()) {
          antennaState =
              response
                  .getStreamEvent()
                  .getPlanMonitoringEvent()
                  .getGroundStationState()
                  .getAntenna();

          onCompleted();
        }
      }

      @Override
      public void onError(Throwable t) {
        testCompletionFuture.setException(t);
      }

      @Override
      public void onCompleted() {
        testCompletionFuture.set(true);
      }

      public AntennaState getAntennaState() {
        return antennaState;
      }
    }

    EventTestStreamObserver responseObserver = new EventTestStreamObserver();
    StreamObserver<SatelliteStreamRequest> requestObserver = streamer.openStream(responseObserver);

    requestObserver.onNext(
        SatelliteStreamRequest.newBuilder()
            .setSatelliteId(SATELLITE_ID)
            .setEnableEvents(true)
            .build());

    assertThat(testCompletionFuture.get()).isTrue();

    // Check antenna states are valid.
    assertThat(responseObserver.getAntennaState()).isNotNull();
    assertThat(responseObserver.getAntennaState().getAzimuth().getCommand()).isEqualTo(1.0);
    assertThat(responseObserver.getAntennaState().getAzimuth().getMeasured()).isEqualTo(1.02);
    assertThat(responseObserver.getAntennaState().getElevation().getCommand()).isEqualTo(20.0);
    assertThat(responseObserver.getAntennaState().getElevation().getMeasured()).isEqualTo(19.5);
  }
}
