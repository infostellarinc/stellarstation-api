# Copyright 2018 Infostellar, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""The factory of creating gRPC stubs for StellarStation API."""

import os

from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc
from stellarstation.api.v1 import stellarstation_pb2_grpc
from stellarstation.api.v1.groundstation import groundstation_pb2_grpc

# The file of StellarStation API key. The private key downloaded from the StellarStation Console.
ENV_API_KEY_NAME = "STELLARSTATION_API_KEY"

# The API URL of StellarStation API.
ENV_API_URL_NAME = "STELLARSTATION_API_URL"


class StubFactory:
    # Initialize the channel for gRPC connection with JWT credential.
    def __init__(self):
        api_key = os.getenv(ENV_API_KEY_NAME)
        api_url = os.getenv(ENV_API_URL_NAME, "api.stellarstation.com:443")

        if not api_key:
            raise Exception('{} need to be set.' % ENV_API_KEY_NAME)

        credentials = google_auth_jwt.Credentials.from_service_account_file(
            api_key,
            audience='https://api.stellarstation.com')

        jwt_creds = google_auth_jwt.OnDemandCredentials.from_signing_credentials(
            credentials)
        self.channel = google_auth_transport_grpc.secure_authorized_channel(
            jwt_creds, None, api_url)

    # Returns the stub for StellarStationService.
    def get_satellite_service_stub(self):
        return stellarstation_pb2_grpc.StellarStationServiceStub(self.channel)

    # Returns the stub for GroundStationService.
    def get_gs_service_stub(self):
        return groundstation_pb2_grpc.GroundStationServiceStub(self.channel)
