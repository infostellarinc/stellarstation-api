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

import com.stellarstation.api.v1.ListUpcomingAvailablePassesRequest;
import com.stellarstation.api.v1.ListUpcomingAvailablePassesResponse;
import com.stellarstation.api.v1.Pass;
import com.stellarstation.api.v1.StellarStationServiceGrpc.StellarStationServiceBlockingStub;
import java.util.List;
import javax.annotation.Nullable;
import javax.inject.Inject;

public class PassManager {
  private final StellarStationServiceBlockingStub client;

  @Inject
  public PassManager(StellarStationServiceBlockingStub client) {
    this.client = client;
  }

  @Nullable
  public List<Pass> list(String satelliteId) {
    ListUpcomingAvailablePassesResponse res =
        client.listUpcomingAvailablePasses(
            ListUpcomingAvailablePassesRequest.newBuilder().setSatelliteId(satelliteId).build());

    if (res == null) {
      return null;
    }

    return res.getPassList();
  }
}
