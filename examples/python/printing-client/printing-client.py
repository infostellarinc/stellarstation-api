# Copyright 2019 Infostellar, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""The Python implementation of the gRPC stellarstation-api client."""

from queue import Queue

import base64
import os
import time

import grpc

from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc
from stellarstation.api.v1 import stellarstation_pb2
from stellarstation.api.v1 import stellarstation_pb2_grpc

SATELLITE_ID = '5'


def run():
    # Load the private key downloaded from the StellarStation Console.
    credentials = google_auth_jwt.Credentials.from_service_account_file(
        '../../fakeserver/src/main/jib/var/keys/api-key.json',
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
    request_queue = Queue()
    request_iterator = generate_request(request_queue)

    for response in client.OpenSatelliteStream(request_iterator):
        if response.HasField("receive_telemetry_response"):
            for telemetry in response.receive_telemetry_response.telemetry:
              print(
                  "Got response: ",
                  base64.b64encode(telemetry.data)[:100])

            command = [
                bytes(b'a' * 5000),
                bytes(b'b' * 5000),
                bytes(b'c' * 5000),
                bytes(b'd' * 5000),
                bytes(b'e' * 5000),
            ]
            request_queue.put(command)
            time.sleep(1)


# This generator yields the requests to send on the stream opened by OpenSatelliteStream.
# The client side of the stream will be closed when this generator returns (in this example, it never returns).
def generate_request(queue):
    # Send the first request to activate the stream. Telemetry will start
    # to be received at this point.
    yield stellarstation_pb2.SatelliteStreamRequest(satellite_id=SATELLITE_ID)

    while True:
        commands = queue.get()
        command_request = stellarstation_pb2.SendSatelliteCommandsRequest(command=commands)

        satellite_stream_request = stellarstation_pb2.SatelliteStreamRequest(
            satellite_id=SATELLITE_ID,
            send_satellite_commands_request=command_request)

        yield satellite_stream_request
        queue.task_done()


if __name__ == '__main__':
    run()
