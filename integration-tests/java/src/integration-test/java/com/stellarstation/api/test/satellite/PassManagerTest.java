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

import static org.assertj.core.api.Assertions.assertThat;

import com.stellarstation.api.test.auth.ApiClientModule;
import com.stellarstation.api.test.util.TimestampUtilities;
import com.stellarstation.api.v1.Pass;
import dagger.Component;
import java.time.Instant;
import java.util.List;
import javax.inject.Singleton;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class PassManagerTest {
  private PassManager manager;

  private static final String SATELLITE_ID = "98";

  @BeforeEach
  void setUp() {
    manager = DaggerPassManagerTest_PassComponent.create().manager();
  }

  @Component(modules = ApiClientModule.class)
  @Singleton
  interface PassComponent {
    PassManager manager();
  }

  @Test
  void list() {
    List<Pass> passes = manager.list(SATELLITE_ID);
    assertThat(passes).isNotNull();
    assertThat(passes.size()).isNotZero();

    Pass pass = passes.get(0);

    // Check AOS and LOS.
    Instant aos = TimestampUtilities.toInstant(pass.getAosTime());
    Instant los = TimestampUtilities.toInstant(pass.getLosTime());
    Instant now = Instant.now();
    assertThat(aos).isAfter(now);
    assertThat(los).isAfter(now);
    assertThat(los).isAfter(aos);

    // Check other properties.
    assertThat(pass.getChannelSetTokenCount()).isNotZero();
    assertThat(pass.getMaxElevationDegrees()).isGreaterThan(0);
  }
}
