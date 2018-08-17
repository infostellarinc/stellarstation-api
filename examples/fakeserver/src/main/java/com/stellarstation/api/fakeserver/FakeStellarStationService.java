/*
 * Copyright 2018 Infostellar, Inc.
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

package com.stellarstation.api.fakeserver;

import com.google.protobuf.ByteString;
import com.google.protobuf.util.Timestamps;
import com.linecorp.armeria.common.RequestContext;
import com.linecorp.armeria.server.ServiceRequestContext;
import com.stellarstation.api.v1.ReceiveTelemetryResponse;
import com.stellarstation.api.v1.SatelliteStreamRequest;
import com.stellarstation.api.v1.SatelliteStreamResponse;
import com.stellarstation.api.v1.StellarStationServiceGrpc.StellarStationServiceImplBase;
import com.stellarstation.api.v1.Telemetry;
import com.typesafe.config.ConfigMemorySize;
import io.grpc.Status;
import io.grpc.StatusException;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import java.time.Clock;
import java.time.Duration;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.ThreadLocalRandom;
import java.util.concurrent.TimeUnit;
import javax.inject.Inject;

/**
 * A fake implementation of {@link StellarStationServiceImplBase} which:
 *
 * <ul>
 *   <li>Returns 1MB of random telemetry payload every second.
 *   <li>Echos received comand payloads back as telemetry payloads.
 *   <li>Cancels the stream after 5 minutes.
 * </ul>
 *
 * This functionality should allow API clients to check send/receive functionality along with
 * recovering from errors (e.g., reconnecting the stream).
 */
class FakeStellarStationService extends StellarStationServiceImplBase {
  private final FakeServerConfig config;

  @Inject
  FakeStellarStationService(FakeServerConfig config) {
    this.config = config;
  }

  @Override
  public StreamObserver<SatelliteStreamRequest> openSatelliteStream(
      StreamObserver<SatelliteStreamResponse> responseObserver) {
    ServiceRequestContext ctx = RequestContext.current();
    ctx.setRequestTimeout(Duration.ZERO);
    ctx.setMaxRequestLength(0);
    ScheduledFuture<?> future =
        ctx.eventLoop()
            .scheduleAtFixedRate(
                () -> sendRandomTelemetry(config.getTelemetryPayloadSize(), responseObserver),
                0,
                config.getTelemetryPublishingFrequency().toMillis(),
                TimeUnit.MILLISECONDS);

    ctx.eventLoop()
        .schedule(
            () -> {
              future.cancel(true);
              responseObserver.onError(new StatusException(Status.CANCELLED));
            },
            config.getSessionTimeout().getSeconds(),
            TimeUnit.SECONDS);

    return new StreamObserver<SatelliteStreamRequest>() {
      @Override
      public void onNext(SatelliteStreamRequest value) {
        if (!value.getSatelliteId().equals("5")) {
          throw new StatusRuntimeException(Status.INVALID_ARGUMENT);
        }
        for (ByteString payload : value.getSendSatelliteCommandsRequest().getCommandList()) {
          sendTelemetry(payload, responseObserver);
        }
      }

      @Override
      public void onError(Throwable t) {
        future.cancel(true);
      }

      @Override
      public void onCompleted() {
        future.cancel(true);
        responseObserver.onCompleted();
      }
    };
  }

  private static void sendRandomTelemetry(
      ConfigMemorySize size, StreamObserver<SatelliteStreamResponse> responseObserver) {
    byte[] payload = new byte[(int) size.toBytes()];
    ThreadLocalRandom.current().nextBytes(payload);
    sendTelemetry(ByteString.copyFrom(payload), responseObserver);
  }

  private static void sendTelemetry(
      ByteString payload, StreamObserver<SatelliteStreamResponse> responseObserver) {
    SatelliteStreamResponse response =
        SatelliteStreamResponse.newBuilder()
            .setReceiveTelemetryResponse(
                ReceiveTelemetryResponse.newBuilder()
                    .setTelemetry(
                        Telemetry.newBuilder()
                            .setTimeFirstByteReceived(
                                Timestamps.fromMillis(Clock.systemUTC().millis()))
                            .setTimeLastByteReceived(
                                Timestamps.fromMillis(
                                    Clock.systemUTC().millis() + TimeUnit.SECONDS.toMillis(1)))
                            .setData(payload)
                            .build())
                    .build())
            .build();
    responseObserver.onNext(response);
  }
}
