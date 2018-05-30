/*
 * Copyright 2018 Infostellar, Inc.
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

package com.stellarstation.api.fakeserver;

import static java.util.concurrent.CompletableFuture.completedFuture;

import com.google.api.client.json.jackson2.JacksonFactory;
import com.google.api.client.json.webtoken.JsonWebSignature;
import com.google.common.io.Resources;
import com.linecorp.armeria.server.ServiceRequestContext;
import com.linecorp.armeria.server.auth.Authorizer;
import com.linecorp.armeria.server.auth.OAuth2Token;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.PublicKey;
import java.util.concurrent.CompletionStage;
import javax.inject.Inject;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.openssl.PEMParser;
import org.bouncycastle.openssl.jcajce.JcaPEMKeyConverter;

/**
 * An {@link Authorizer} that verifies fakeserver's private key has been correctly used to provide
 * credentials.
 */
class FakeApiKeyAuthorizer implements Authorizer<OAuth2Token> {

  private static final Logger logger = LogManager.getLogger();

  private final PublicKey publicKey;

  @Inject
  FakeApiKeyAuthorizer() {
    publicKey = readPublicKey();
  }

  @Override
  public CompletionStage<Boolean> authorize(ServiceRequestContext ctx, OAuth2Token data) {
    final JsonWebSignature signature;
    try {
      signature = JsonWebSignature.parse(JacksonFactory.getDefaultInstance(), data.accessToken());
    } catch (IOException e) {
      logger.warn("Could not parse access token.", e);
      return completedFuture(false);
    }

    try {
      if (!signature.verifySignature(publicKey)) {
        return completedFuture(false);
      }
    } catch (GeneralSecurityException e) {
      logger.warn("Could not initialize crypto.", e);
      return completedFuture(false);
    }

    if (!signature.getPayload().getIssuer().equals("fakeclient@example.com")) {
      return completedFuture(false);
    }
    return completedFuture(true);
  }

  private static PublicKey readPublicKey() {
    try (var parser =
        new PEMParser(
            new InputStreamReader(
                Resources.getResource("public-key.pem").openStream(), StandardCharsets.UTF_8))) {
      Object obj;
      while ((obj = parser.readObject()) != null) {
        if (obj instanceof SubjectPublicKeyInfo) {
          return new JcaPEMKeyConverter().getPublicKey((SubjectPublicKeyInfo) obj);
        }
      }
      throw new IllegalStateException("Could not find public key.");
    } catch (IOException e) {
      throw new IllegalStateException("Could not read public key.", e);
    }
  }
}
