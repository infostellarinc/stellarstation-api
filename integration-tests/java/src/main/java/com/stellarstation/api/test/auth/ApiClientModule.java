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
package com.stellarstation.api.test.auth;

import com.google.auth.oauth2.ServiceAccountJwtAccessCredentials;
import com.stellarstation.api.v1.StellarStationServiceGrpc;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import dagger.Module;
import dagger.Provides;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.auth.MoreCallCredentials;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.net.URI;
import java.nio.file.Files;
import java.nio.file.Paths;
import javax.inject.Singleton;

@Module
public class ApiClientModule {

  public static final String CONFIG = "apiIntegrationTest";

  public static final String API_KEY = "apiKey";
  public static final String API_SERVER_URL_KEY = "server.url";
  public static final String API_SERVER_PORT_KEY = "server.port";

  @Provides
  @Singleton
  public static Config config() {
    return ConfigFactory.load();
  }

  @Provides
  @Singleton
  public static StellarStationServiceGrpc.StellarStationServiceBlockingStub client(Config config) {
    Config apiConfig = config.getConfig(CONFIG);

    try {
      ServiceAccountJwtAccessCredentials credentials =
          ServiceAccountJwtAccessCredentials.fromStream(
              Files.newInputStream(Paths.get(apiConfig.getString(API_KEY))),
              URI.create("https://api.stellarstation.com"));

      ManagedChannel channel =
          ManagedChannelBuilder.forAddress(
                  apiConfig.getString(API_SERVER_URL_KEY), apiConfig.getInt(API_SERVER_PORT_KEY))
              .build();
      StellarStationServiceGrpc.StellarStationServiceBlockingStub client =
          StellarStationServiceGrpc.newBlockingStub(channel)
              .withCallCredentials(MoreCallCredentials.from(credentials));

      return client;
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }
  }

  private ApiClientModule() {}
}
