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
package com.stellarstation.api.fakeserver;

import com.typesafe.config.ConfigMemorySize;
import java.time.Duration;
import org.curioswitch.common.server.framework.immutables.JavaBeanStyle;
import org.immutables.value.Value.Immutable;
import org.immutables.value.Value.Modifiable;

@Immutable
@Modifiable
@JavaBeanStyle
public interface FakeServerConfig {

  /** Telemetry will be published on every expiry of this duration. */
  Duration getTelemetryPublishingInterval();

  /** Gets the size of each telemetry blob. */
  ConfigMemorySize getTelemetryPayloadSize();

  /** Gets the length of time until the server closes the stream. */
  Duration getSessionTimeout();
}
