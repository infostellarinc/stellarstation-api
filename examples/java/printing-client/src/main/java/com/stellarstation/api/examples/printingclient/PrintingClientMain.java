/*
 * Copyright 2019 Infostellar, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package com.stellarstation.api.examples.printingclient;

import com.google.auth.oauth2.ServiceAccountJwtAccessCredentials;
import com.google.common.base.Strings;
import com.google.common.io.Resources;
import com.google.protobuf.ByteString;
import com.google.protobuf.Timestamp;
import com.stellarstation.api.v1.SatelliteStreamRequest;
import com.stellarstation.api.v1.SatelliteStreamResponse;
import com.stellarstation.api.v1.SendSatelliteCommandsRequest;
import com.stellarstation.api.v1.StellarStationServiceGrpc;
import com.stellarstation.api.v1.StellarStationServiceGrpc.StellarStationServiceStub;
import io.grpc.ManagedChannel;
import io.grpc.auth.MoreCallCredentials;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.StreamObserver;
import java.net.URI;
import java.util.Base64;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class PrintingClientMain {

  private static final Logger logger = LoggerFactory.getLogger(PrintingClientMain.class);
  private static StreamObserver<SatelliteStreamRequest> requestStream;

  public static void main(String[] args) throws Exception {
    // Load the private key downloaded from the StellarStation Console.
    ServiceAccountJwtAccessCredentials credentials =
        ServiceAccountJwtAccessCredentials.fromStream(
            Resources.getResource("api-key.json").openStream(),
            URI.create("https://api.stellarstation.com"));

    // Setup the gRPC client.
    ManagedChannel channel =
        NettyChannelBuilder.forAddress("localhost", 8080)
            // Only for testing, this should not be set when accessing the actual API
            .sslContext(
                GrpcSslContexts.forClient()
                    .trustManager(Resources.getResource("tls.crt").openStream())
                    .build())
            .build();
    StellarStationServiceStub client =
        StellarStationServiceGrpc.newStub(channel)
            .withCallCredentials(MoreCallCredentials.from(credentials));

    // Open the stream with an observer to handle telemetry responses.
    requestStream =
        client.openSatelliteStream(
            new StreamObserver<>() {
              @Override
              public void onNext(SatelliteStreamResponse value) {
                logger.info(
                    "Got response: {}",
                    Base64.getEncoder()
                        .encodeToString(
                            value
                                .getReceiveTelemetryResponse()
                                .getTelemetry()
                                .getData()
                                .toByteArray())
                        .substring(0, 100));

                // acknowledge message received (only needed if
                // `SatelliteStreamRequest.enable_flow_control` is true)
                String msgAckId =
                    value.getReceiveTelemetryResponse().getTelemetry().getMessageAckId();
                Timestamp timestamp =
                    Timestamp.newBuilder().setSeconds(System.currentTimeMillis() / 1000).build();
                var telemetryReceivedAck =
                    SatelliteStreamRequest.newBuilder()
                        .getTelemetryReceivedAckBuilder()
                        .setMessageAckId(msgAckId)
                        .setReceivedTimestamp(timestamp)
                        .build();
                requestStream.onNext(
                    SatelliteStreamRequest.newBuilder()
                        .setTelemetryReceivedAck(telemetryReceivedAck)
                        .build());
              }

              @Override
              public void onError(Throwable t) {
                logger.warn("Got error from server.", t);
              }

              @Override
              public void onCompleted() {}
            });

    // Send the first request to activate the stream. Telemetry will start to be received at
    // this point. EnableFlowControl=true is optional but recommended for high bitrate connections
    // greater than 50 Mbps
    requestStream.onNext(
        SatelliteStreamRequest.newBuilder().setSatelliteId("5").setEnableFlowControl(true).build());

    ScheduledExecutorService commandExecutor = Executors.newScheduledThreadPool(1);

    // Send some arbitrary commands every 3 seconds.
    commandExecutor.scheduleAtFixedRate(
        () -> {
          requestStream.onNext(
              SatelliteStreamRequest.newBuilder()
                  .setSatelliteId("5")
                  .setSendSatelliteCommandsRequest(
                      SendSatelliteCommandsRequest.newBuilder()
                          .addCommand(ByteString.copyFromUtf8(Strings.repeat("a", 5000)))
                          .addCommand(ByteString.copyFromUtf8(Strings.repeat("b", 5000)))
                          .addCommand(ByteString.copyFromUtf8(Strings.repeat("c", 5000)))
                          .addCommand(ByteString.copyFromUtf8(Strings.repeat("d", 5000)))
                          .addCommand(ByteString.copyFromUtf8(Strings.repeat("e", 5000)))
                          .build())
                  .build());
        },
        0,
        3,
        TimeUnit.SECONDS);

    Thread.sleep(Long.MAX_VALUE);
  }

  private PrintingClientMain() {}
}
