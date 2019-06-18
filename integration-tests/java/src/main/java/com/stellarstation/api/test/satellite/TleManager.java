package com.stellarstation.api.test.satellite;

import com.stellarstation.api.v1.GetTleRequest;
import com.stellarstation.api.v1.GetTleResponse;
import com.stellarstation.api.v1.StellarStationServiceGrpc.StellarStationServiceBlockingStub;
import com.stellarstation.api.v1.orbit.Tle;
import javax.annotation.Nullable;
import javax.inject.Inject;

public class TleManager {
  private StellarStationServiceBlockingStub client;

  @Inject
  public TleManager(StellarStationServiceBlockingStub client) {
    this.client = client;
  }

  @Nullable
  public Tle getTle(String satelliteId) {
    GetTleResponse res =
        client.getTle(GetTleRequest.newBuilder().setSatelliteId(satelliteId).build());

    if (res == null) {
      return null;
    }

    return res.getTle();
  }
}
