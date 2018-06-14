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
"""The Python implementation of the gRPC stellarstation-api client."""

import base64
import os
import time

import grpc
import stellarstation_pb2
import stellarstation_pb2_grpc

from google import auth as google_auth
from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc

os.environ['GRPC_SSL_CIPHER_SUITES'] = 'ECDHE-RSA-AES128-GCM-SHA256'

SATELLITE_ID = '5'


def run():
  # Load the private key downloaded from the StellarStation Console.
  credentials = google_auth_jwt.Credentials.from_service_account_file(
      '../../fakeserver/src/misc/api-key.json',
      audience='https://api.stellarstation.com')

  # Setup the gRPC client.
  jwt_creds = google_auth_jwt.OnDemandCredentials.from_signing_credentials(
      credentials)
  channel_credential = grpc.ssl_channel_credentials(
      open('../../fakeserver/src/main/resources/tls.crt', 'br').read())
  channel = google_auth_transport_grpc.secure_authorized_channel(
      jwt_creds, None, 'localhost:8080', channel_credential)
  client = stellarstation_pb2_grpc.StellarStationServiceStub(channel)

  # Open satellite stream
  request_iterator = generate_request()
  for value in client.OpenSatelliteStream(request_iterator):
    print(
        "Got response: ",
        base64.b64encode(value.receive_telemetry_response.telemetry.data)[:100])


# This generator yields the requests to send on the stream opened by OpenSatelliteStream.
# The client side of the stream will be closed when this generator returns (in this example, it never returns).
def generate_request():

  # Send the first request to activate the stream. Telemetry will start
  # to be received at this point.
  yield stellarstation_pb2.SatelliteStreamRequest(satellite_id=SATELLITE_ID)

  while True:
    command_request = stellarstation_pb2.SendSatelliteCommandsRequest(
        output_framing=0,
        command=[
            bytes(b'a' * 5000),
            bytes(b'b' * 5000),
            bytes(b'c' * 5000),
            bytes(b'd' * 5000),
            bytes(b'e' * 5000),
        ])

    satellite_stream_request = stellarstation_pb2.SatelliteStreamRequest(
        satellite_id=SATELLITE_ID,
        send_satellite_commands_request=command_request)

    yield satellite_stream_request
    time.sleep(3)


if __name__ == '__main__':
  run()
