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
import com.stellarstation.api.v1.orbit.Tle;
import dagger.Component;
import javax.inject.Singleton;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class TleManagerTest {
  private TleManager manager;

  private static final String SATELLITE_ID = "98";

  @BeforeEach
  void setUp() {
    manager = DaggerTleManagerTest_TleComponent.create().manager();
  }

  @Component(modules = ApiClientModule.class)
  @Singleton
  interface TleComponent {
    TleManager manager();
  }

  @Test
  void getTle() {
    Tle tle = manager.getTle(SATELLITE_ID);
    assertThat(tle).isNotNull();
    assertThat(tle.getLine1()).isNotNull();
    assertThat(tle.getLine2()).isNotNull();
  }
}
