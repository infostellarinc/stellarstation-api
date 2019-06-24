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

import com.stellarstation.api.v1.SatelliteStreamRequest;
import com.stellarstation.api.v1.SatelliteStreamResponse;
import com.stellarstation.api.v1.StellarStationServiceGrpc.StellarStationServiceStub;
import io.grpc.stub.StreamObserver;
import javax.annotation.Nullable;
import javax.inject.Inject;

public class SatelliteStreamer {
  private final StellarStationServiceStub client;

  @Inject
  public SatelliteStreamer(StellarStationServiceStub client) {
    this.client = client;
  }

  @Nullable
  public StreamObserver<SatelliteStreamRequest> openStream(
      StreamObserver<SatelliteStreamResponse> responseObserver) {
    StreamObserver<SatelliteStreamRequest> requestObserver =
        client.openSatelliteStream(responseObserver);
    return requestObserver;
  }
}
