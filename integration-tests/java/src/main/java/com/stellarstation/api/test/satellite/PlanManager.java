/*
 * Copyright 2020 Infostellar, Inc.
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

package com.stellarstation.api.test.satellite;

import com.stellarstation.api.test.util.TimestampUtilities;
import com.stellarstation.api.v1.ListPlansRequest;
import com.stellarstation.api.v1.ListPlansResponse;
import com.stellarstation.api.v1.Plan;
import com.stellarstation.api.v1.StellarStationServiceGrpc.StellarStationServiceBlockingStub;
import java.time.Instant;
import java.util.List;
import javax.annotation.Nullable;
import javax.inject.Inject;

public class PlanManager {
  private final StellarStationServiceBlockingStub client;

  @Inject
  public PlanManager(StellarStationServiceBlockingStub client) {
    this.client = client;
  }

  @Nullable
  public List<Plan> list(String satelliteId, Instant aosAfter, Instant aosBefore) {
    ListPlansResponse res =
        client.listPlans(
            ListPlansRequest.newBuilder()
                .setSatelliteId(satelliteId)
                .setAosAfter(TimestampUtilities.fromInstant(aosAfter))
                .setAosBefore(TimestampUtilities.fromInstant(aosBefore))
                .build());

    if (res == null) {
      return null;
    }

    return res.getPlanList();
  }
}
