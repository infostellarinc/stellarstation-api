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

import time
from concurrent import futures

import grpc
from google.protobuf.timestamp_pb2 import Timestamp

from stellarstation.api.v1.groundstation import groundstation_pb2_grpc
from stellarstation.api.v1.groundstation import groundstation_pb2


ONE_DAY_IN_SECONDS = 60 * 60 * 24
SECONDS_BEFORE_PLAN_START = 10
PLAN_DURATION_SECONDS = 600
CURRENT_PLAN_ID = "3"


class GroundStationServiceServicer(groundstation_pb2_grpc.GroundStationServiceServicer):
    def ListPlans(self, request, context) -> groundstation_pb2.ListPlansResponse:
        """Lists the plans for a particular ground station.

        The request will be closed with an `INVALID_ARGUMENT` status if `ground_station_id`,
        `aos_after`, or `aos_before` are missing, or the duration between the two times is longer than
        31 days.
        """
        print('Got request for ListPlans')
        if not request.ground_station_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details('Ground station ID not set')
            raise RuntimeError('Ground station ID not set')
        if request.aos_after is None:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details('AOS after not set')
            raise RuntimeError('AOS after not set')
        if request.aos_before is None:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details('AOS before not set')
            raise RuntimeError('AOS before not set')

        delta = request.aos_before.ToDatetime() - request.aos_after.ToDatetime()
        if delta.days > 31:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details('Duration between aos_after and aos_before > 31 days')
            raise RuntimeError('Duration between aos_after and aos_before > 31 days')

        now = time.time()

        satellite_coordinates = [
            groundstation_pb2.SatelliteCoordinates(
                time=Timestamp(seconds=int(now + SECONDS_BEFORE_PLAN_START + i)),
                range_rate=(2.1e7 + i * 1e4)
            )
            for i in range(PLAN_DURATION_SECONDS)
        ]
        # TODO: Fill in the other fields of plan
        response = groundstation_pb2.ListPlansResponse(
            plan=[groundstation_pb2.Plan(
                plan_id=CURRENT_PLAN_ID,
                satellite_coordinates=satellite_coordinates)]
        )
        return response

    def OpenGroundStationStream(self, request_iterator, context):
        """Open a stream from a ground station. The returned stream is bi-directional - it is used by
        the ground station to send telemetry received from a satellite and receive commands to send to
        the satellite. The ground station must keep this stream open while it is connected to the
        StellarStation network for use in executing plans - if the stream is cut, it must be
        reconnected with exponential backoff.

        The first `GroundStationStreamRequest` sent on the stream is used for configuring the stream.
        Unless otherwise specified, all configuration is taken from the first request and configuration
        values in subsequent requests will be ignored.

        There is no restriction on the number of active streams from a ground station (i.e., streams
        opened with the same `ground_station_id`). Most ground stations will issue a single stream to
        receive commands and send telemetry, but in certain cases, such as if uplink and downlink are
        handled by different computers, it can be appropriate to have multiple processes with their
        own stream. If opening multiple streams for a single ground station, it is the client's
        responsibility to handle the streams appropriately, for example by ensuring only one stream
        sends commands so they are not duplicated.

        If the ground station is not found or the API client is not authorized for it, the stream will
        be closed with a `NOT_FOUND` error.

        Status: ALPHA This API is under development and may not work correctly or be changed in backwards
        incompatible ways in the future.
        """
        request = next(request_iterator)
        ground_station_id = request.ground_station_id
        if not request.ground_station_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details('Ground station ID not set')
            raise RuntimeError('Ground station ID not set')
        for request in request_iterator:
            if request.ground_station_id != ground_station_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details('Unexpected ground station ID')
                raise RuntimeError('Unexpected ground staiton ID')
            print("Received request")
            print("Ground station ID", request.ground_station_id)
            print("Stream tag", request.stream_tag)
            if request.HasField('satellite_telemetry'):
                print("Satellite telemetry", request.satellite_telemetry)
                if request.satellite_telemetry.plan_id != CURRENT_PLAN_ID:
                    print("WARNING: plan ID from client telemetry is not equal to current plan ID")
            if request.HasField('stream_event'):
                print("Stream event", repr(request.stream_event))
            response = groundstation_pb2.GroundStationStreamResponse(
                plan_id=CURRENT_PLAN_ID,
                satellite_commands=groundstation_pb2.SatelliteCommands(
                    command=[bytes('command1', encoding='ascii'),
                             bytes('command2', encoding='ascii'),
                             bytes('command3', encoding='ascii')],
                )
            )
            yield response


if __name__ == '__main__':
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    groundstation_pb2_grpc.add_GroundStationServiceServicer_to_server(
        GroundStationServiceServicer(), server
    )
    server.add_insecure_port('[::]:50051')
    server.start()

    print('started server')

    try:
        while True:
            time.sleep(ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)
