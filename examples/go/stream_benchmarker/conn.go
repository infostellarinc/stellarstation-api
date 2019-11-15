// Copyright © 2019 Infostellar, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package benchmark

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// NewDefaultCredentials initializes gRPC credentials using Stellar Default Credentials.
func newDefaultCredentials(apiKey string) (credentials.PerRPCCredentials, error) {
	return oauth.NewJWTAccessFromFile(apiKey)
}

// Dial opens a gRPC connection to the StellarStation API with authentication setup.
func Dial(apiKey string, apiURL string) (*grpc.ClientConn, error) {
	creds, err := newDefaultCredentials(apiKey)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{}

	return grpc.Dial(
		apiURL,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(creds))
}
