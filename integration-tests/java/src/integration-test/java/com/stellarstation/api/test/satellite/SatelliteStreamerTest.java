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

import com.google.common.util.concurrent.Futures;
import com.google.common.util.concurrent.ListenableFuture;
import com.google.common.util.concurrent.SettableFuture;
import com.google.protobuf.ByteString;
import com.stellarstation.api.test.auth.ApiClientModule;
import com.stellarstation.api.v1.SatelliteStreamRequest;
import com.stellarstation.api.v1.SatelliteStreamResponse;
import com.stellarstation.api.v1.SendSatelliteCommandsRequest;
import com.stellarstation.api.v1.monitoring.AntennaState;
import com.stellarstation.api.v1.monitoring.GroundStationState;
import dagger.Component;
import io.grpc.stub.StreamObserver;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
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
  void openStream() throws Exception {
    final List<Double> azimuthCommand = new ArrayList();
    final List<Double> azimuthMeasured = new ArrayList();
    final List<Double> elevationCommand = new ArrayList();
    final List<Double> elevationMeasured = new ArrayList();

    final SettableFuture<Boolean> commandTestFuture = SettableFuture.create();
    final SettableFuture<StreamObserver<SatelliteStreamRequest>> requestObserverFuture =
        SettableFuture.create();

    StreamObserver<SatelliteStreamRequest> requestObserver =
        streamer.openStream(
            new StreamObserver<SatelliteStreamResponse>() {
              private int initialState = -1;

              @Override
              public void onNext(SatelliteStreamResponse response) {
                if (response.hasReceiveTelemetryResponse()) {
                  ByteString data = response.getReceiveTelemetryResponse().getTelemetry().getData();
                  if (data.size() > 2) {
                    // The second last byte of the telemetry indicates the current state of the
                    // fake satellite used in the test. The value is either of 0 or 1.
                    int state = data.byteAt(data.size() - 2);
                    if (initialState < 0) {
                      initialState = state;
                    } else {
                      commandTestFuture.set(initialState == 1 - state);
                    }

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
                  }
                } else {
                  GroundStationState gsState =
                      response.getStreamEvent().getPlanMonitoringEvent().getGroundStationState();
                  AntennaState antennaState = gsState.getAntenna();

                  azimuthCommand.add(antennaState.getAzimuth().getCommand());
                  azimuthMeasured.add(antennaState.getAzimuth().getMeasured());
                  elevationCommand.add(antennaState.getElevation().getCommand());
                  elevationMeasured.add(antennaState.getElevation().getMeasured());
                }
              }

              @Override
              public void onError(Throwable t) {
                commandTestFuture.setException(t);
              }

              @Override
              public void onCompleted() {}
            });
    requestObserverFuture.set(requestObserver);

    requestObserver.onNext(
        SatelliteStreamRequest.newBuilder()
            .setSatelliteId(SATELLITE_ID)
            .setEnableEvents(true)
            .build());

    ListenableFuture<Boolean> timeoutFuture =
        Futures.withTimeout(
            commandTestFuture, 30, TimeUnit.SECONDS, Executors.newSingleThreadScheduledExecutor());

    // Check safe mode state is changed..
    assertThat(timeoutFuture.get()).isTrue();

    // Check antenna states are valid.
    assertThat(azimuthCommand).containsOnly(1.0);
    assertThat(azimuthMeasured).containsOnly(1.02);
    assertThat(elevationCommand).containsOnly(20.0);
    assertThat(elevationMeasured).containsOnly(19.5);

    requestObserver.onCompleted();
  }
}
