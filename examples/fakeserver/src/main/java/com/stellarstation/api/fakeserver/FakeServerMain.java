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

import com.linecorp.armeria.server.Server;
import com.linecorp.armeria.server.ServerBuilder;
import com.linecorp.armeria.server.auth.AuthService;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigBeanFactory;
import dagger.Component;
import dagger.Module;
import dagger.Provides;
import dagger.multibindings.IntoSet;
import java.util.function.Consumer;
import javax.inject.Singleton;
import org.curioswitch.common.server.framework.ServerModule;
import org.curioswitch.common.server.framework.grpc.GrpcServiceDefinition;

public class FakeServerMain {

  @Module(includes = ServerModule.class)
  abstract static class FakeServerModule {

    @Provides
    @IntoSet
    static GrpcServiceDefinition service(
        FakeStellarStationService service, FakeApiKeyAuthorizer authorizer) {
      return new GrpcServiceDefinition.Builder()
          .addServices(service)
          .decorator(AuthService.builder().addOAuth2(authorizer).newDecorator())
          .path("/")
          .build();
    }

    @Provides
    @Singleton
    static FakeServerConfig schedulerConfig(Config config) {
      return ConfigBeanFactory.create(
              config.getConfig("fakeServer"), ModifiableFakeServerConfig.class)
          .toImmutable();
    }

    @Provides
    @IntoSet
    static Consumer<ServerBuilder> serverCustomizer() {
      return sb -> sb.http(8081);
    }

    private FakeServerModule() {}
  }

  @Component(modules = FakeServerModule.class)
  @Singleton
  interface FakeServerComponent {
    Server server();
  }

  public static void main(String[] args) {
    DaggerFakeServerMain_FakeServerComponent.create().server();
  }

  private FakeServerMain() {}
}
