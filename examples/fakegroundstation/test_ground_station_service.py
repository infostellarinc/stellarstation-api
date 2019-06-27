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
# See the License for the specific language governing permissions and
# limitations under the License.

import time

import grpc
import grpc_testing

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1.groundstation import groundstation_pb2
from stellarstation.api.v1 import transport_pb2

from fakegroundstation.ground_station_service import GroundStationServiceServicer


SECONDS_IN_MINUTE = 60
SECONDS_IN_HOUR = 60 * SECONDS_IN_MINUTE


def setup_test_server():
    servicers = {
        groundstation_pb2.DESCRIPTOR.services_by_name['GroundStationService']: GroundStationServiceServicer()
    }

    return grpc_testing.server_from_dictionary(
        servicers, grpc_testing.strict_real_time()
    )


def test_request() -> None:
    test_server = setup_test_server()

    request = groundstation_pb2.ListPlansRequest(
        ground_station_id="2",
        aos_after=Timestamp(seconds=int(time.time()) - 2 * SECONDS_IN_MINUTE, nanos=0),
        aos_before=Timestamp(seconds=int(time.time() + SECONDS_IN_HOUR), nanos=0),
    )

    list_plans_method = test_server.invoke_unary_unary(
        method_descriptor=(groundstation_pb2.DESCRIPTOR.services_by_name['GroundStationService'].methods_by_name['ListPlans']),
        invocation_metadata={},
        request=request, timeout=1
    )

    response, metadata, code, details = list_plans_method.termination()
    print(response.plan[0].satellite_coordinates[:3])
    assert response.plan[0].plan_id == "3"
    assert code == grpc.StatusCode.OK


def test_stream() -> None:
    test_server = setup_test_server()

    client = test_server.invoke_stream_stream(
        method_descriptor=(
        groundstation_pb2.DESCRIPTOR.services_by_name['GroundStationService'].methods_by_name['OpenGroundStationStream']),
        invocation_metadata={},
        timeout=1
    )
    initial_request = groundstation_pb2.GroundStationStreamRequest(
        ground_station_id="2",
        stream_tag="4",
    )
    client.send_request(initial_request)

    telemetry_request = groundstation_pb2.GroundStationStreamRequest(
        ground_station_id="2",
        stream_tag="4",
        satellite_telemetry=groundstation_pb2.SatelliteTelemetry(
            plan_id="3",
            telemetry=transport_pb2.Telemetry(
                data=bytes('telemetry', encoding='ascii')
            )
        )
    )
    client.send_request(telemetry_request)
    response = client.take_response()
    assert response.plan_id == "3"
    assert response.satellite_commands == groundstation_pb2.SatelliteCommands(
        command=[
            bytes("command1", 'ascii'),
            bytes("command2", 'ascii'),
            bytes("command3", 'ascii'),
        ]
    )

    client.requests_closed()
    client.termination()
