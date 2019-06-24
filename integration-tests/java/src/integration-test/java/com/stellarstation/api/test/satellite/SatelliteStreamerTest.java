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
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;
import java.util.concurrent.LinkedBlockingQueue;
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
  void openStream() throws InterruptedException {
    final CountDownLatch latch = new CountDownLatch(1);
    final BlockingQueue<SendSatelliteCommandsRequest> queue =
        new LinkedBlockingQueue<SendSatelliteCommandsRequest>();

    final List<Double> azimuthCommand = new ArrayList();
    final List<Double> azimuthMeasured = new ArrayList();
    final List<Double> elevationCommand = new ArrayList();
    final List<Double> elevationMeasured = new ArrayList();
    final List<Integer> stateList = new ArrayList();

    StreamObserver<SatelliteStreamRequest> requestObserver =
        streamer.openStream(
            new StreamObserver<SatelliteStreamResponse>() {

              @Override
              public void onNext(SatelliteStreamResponse response) {
                if (response.hasReceiveTelemetryResponse()) {
                  ByteString data = response.getReceiveTelemetryResponse().getTelemetry().getData();
                  if (data.size() > 2) {
                    // The second last byte of the telemetry indicates the current state of the
                    // fake satellite used in the test. The value is either of 0 or 1.
                    int state = data.byteAt(data.size() - 2);
                    stateList.add(state);

                    if (stateList.size() == 2) {
                      latch.countDown();
                    }

                    // Send a command to toggle the state.
                    SendSatelliteCommandsRequest command =
                        SendSatelliteCommandsRequest.newBuilder()
                            .addCommand(ByteString.copyFrom(new byte[] {0x01, 0x01}))
                            .build();
                    queue.add(command);
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
                latch.countDown();
              }

              @Override
              public void onCompleted() {
                latch.countDown();
              }
            });

    requestObserver.onNext(
        SatelliteStreamRequest.newBuilder().setSatelliteId(SATELLITE_ID).build());

    Executor executor = Executors.newSingleThreadExecutor();
    executor.execute(
        () -> {
          try {
            // Sends commands in the blocking queue to the API.
            while (latch.getCount() > 0) {
              SendSatelliteCommandsRequest commandsRequest = queue.take();
              requestObserver.onNext(
                  SatelliteStreamRequest.newBuilder()
                      .setSatelliteId(SATELLITE_ID)
                      .setSendSatelliteCommandsRequest(commandsRequest)
                      .build());
            }
          } catch (Exception e) {
            throw new RuntimeException(e);
          }
        });

    latch.await(30, TimeUnit.SECONDS);

    // Check safe mode state is changed..
    assertThat(stateList).hasSize(2);
    int expected = 1 - stateList.get(0);
    assertThat(expected).isEqualTo(stateList.get(1));

    // Check antenna states are valid.
    assertThat(azimuthCommand).containsOnly(1.0);
    assertThat(azimuthMeasured).containsOnly(1.02);
    assertThat(elevationCommand).containsOnly(20.0);
    assertThat(elevationMeasured).containsOnly(19.5);

    requestObserver.onCompleted();
  }
}
