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

package com.stellarstation.api.test.satellite;

import static org.assertj.core.api.Assertions.assertThat;

import com.stellarstation.api.test.auth.ApiClientModule;
import com.stellarstation.api.test.util.TimestampUtilities;
import com.stellarstation.api.v1.Plan;
import dagger.Component;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import javax.inject.Singleton;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class PlanManagerTest {
  private PlanManager manager;

  private static final String SATELLITE_ID = "98";
  private static final ZoneOffset UTC = ZoneOffset.UTC;

  @BeforeEach
  void setUp() {
    manager = DaggerPlanManagerTest_PlanComponent.create().manager();
  }

  @Component(modules = ApiClientModule.class)
  @Singleton
  interface PlanComponent {
    PlanManager manager();
  }

  @Test
  void list() {
    Instant aosAfter = LocalDateTime.of(2018, 12, 1, 0, 0).toInstant(UTC);
    Instant aosBefore = LocalDateTime.of(2018, 12, 31, 0, 0).toInstant(UTC);

    List<Plan> plans = manager.list(SATELLITE_ID, aosAfter, aosBefore);
    assertThat(plans).isNotNull();
    assertThat(plans.size()).isNotZero();

    Plan plan = plans.get(0);

    // Check AOS and LOS.
    Instant aos = TimestampUtilities.toInstant(plan.getAosTime());
    Instant los = TimestampUtilities.toInstant(plan.getLosTime());
    Instant now = Instant.now();

    // We are assuming that the server will not return passes that are due to start very soon
    // which could cause this to fail. (e.g. within the next few seconds)
    // Also, this could fail when the server and client have significant clock skew.
    assertThat(aos).isBefore(now);

    assertThat(los).isBefore(now);
    assertThat(los).isAfter(aos);

    // Check other properties.
    assertThat(plan.getId()).isNotNull();
    assertThat(plan.getGroundStationCountryCode()).isEqualTo("CA");
    assertThat(plan.getStatus()).isEqualTo(Plan.Status.FAILED);
  }
}
