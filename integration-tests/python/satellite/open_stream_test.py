# Copyright 2019 Infostellar, Inc.
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
"""The integration tests for OpenSatelliteStream API in StellarStationService."""

from queue import Queue

from stellarstation.api.v1 import stellarstation_pb2

SATELLITE_ID = '98'


# This test checks the expected status after sending status change command.
def test_open_satellite_stream(stub_factory):
    client = stub_factory.get_satellite_service_stub()

    request_queue = Queue()
    request_iterator = generate_request(request_queue)

    expected_status = -1
    for response in client.OpenSatelliteStream(request_iterator):
        if response.HasField("receive_telemetry_response"):
            telemetry_data = response.receive_telemetry_response.telemetry.data
            assert len(telemetry_data) > 1

            # The second last byte of the telemetry indicates the current status of the fake satellite used
            # in the test. The value is either of 0 or 1.
            is_safe_mode = int(telemetry_data[-2])
            assert (is_safe_mode == 0 or is_safe_mode == 1)

            if expected_status < 0:
                # Set expected status based on the current value.
                expected_status = 1 - is_safe_mode

                # Send the command to toggle the state.
                command = [bytes(b"\x01\x01")]
                request_queue.put(command)
            else:
                assert is_safe_mode == expected_status
                return
        else:
            gs_state = response.stream_event.plan_monitoring_event.ground_station_state
            assert gs_state.HasField("antenna")
            assert gs_state.antenna.azimuth.command == 1.0
            assert gs_state.antenna.azimuth.measured == 1.02
            assert gs_state.antenna.elevation.command == 20.0
            assert gs_state.antenna.elevation.measured == 19.5


# Yields a request sent to the stream opened by OpenSatelliteStream.
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
