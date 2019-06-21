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
package com.stellarstation.api.test.util;

import com.google.protobuf.Timestamp;
import com.google.protobuf.util.Timestamps;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;

public class TimestampUtilities {

  public static Timestamp fromLocalDateTime(LocalDateTime localDateTime, ZoneOffset zone) {
    Instant instant = localDateTime.toInstant(zone);
    return Timestamp.newBuilder()
        .setSeconds(instant.getEpochSecond())
        .setNanos(instant.getNano())
        .build();
  }

  public static LocalDateTime toLocalDateTime(Timestamp timestamp, ZoneOffset zone) {
    return LocalDateTime.ofEpochSecond(Timestamps.toSeconds(timestamp), timestamp.getNanos(), zone);
  }

  private TimestampUtilities() {}
}
