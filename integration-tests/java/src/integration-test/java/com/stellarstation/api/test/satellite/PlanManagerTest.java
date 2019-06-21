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

import com.stellarstation.api.test.auth.ApiClientModule;
import com.stellarstation.api.test.util.TimestampUtilities;
import com.stellarstation.api.v1.Plan;
import dagger.Component;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import javax.inject.Singleton;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class PlanManagerTest {
  private PlanManager manager;

  private static final String SATELLITE_ID = "98";
  private static final ZoneOffset ZONE_OFFSET = ZoneOffset.UTC;

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
    LocalDateTime aosAfter = LocalDateTime.of(2018, 12, 1, 0, 0);
    LocalDateTime aosBefore = LocalDateTime.of(2018, 12, 31, 0, 0);

    List<Plan> plans = manager.list(SATELLITE_ID, aosAfter, aosBefore, ZONE_OFFSET);
    assertThat(plans).isNotNull();
    assertThat(plans.size()).isNotZero();

    Plan plan = plans.get(0);

    // Check AOS and LOS.
    LocalDateTime aos = TimestampUtilities.toLocalDateTime(plan.getAosTime(), ZONE_OFFSET);
    LocalDateTime los = TimestampUtilities.toLocalDateTime(plan.getLosTime(), ZONE_OFFSET);
    LocalDateTime now = LocalDateTime.now(ZONE_OFFSET);
    assertThat(aos).isBefore(now);
    assertThat(los).isBefore(now);
    assertThat(los).isAfter(aos);

    // Check other properties.
    assertThat(plan.getId()).isNotNull();
    assertThat(plan.getGroundStationCountryCode()).isEqualTo("CA");
    assertThat(plan.getStatus()).isEqualTo(Plan.Status.FAILED);
  }
}
