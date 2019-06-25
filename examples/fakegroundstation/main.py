import grpc

import time
from concurrent import futures

from stellarstation.api.v1.groundstation import groundstation_pb2_grpc

from fakegroundstation.ground_station_service import GroundStationServiceServicer

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


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
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)
