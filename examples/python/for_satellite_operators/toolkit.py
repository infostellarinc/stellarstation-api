# Copyright 2022 Infostellar, Inc.

# A nice set of tools used by the examples.

from enum import Enum

from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc

from stellarstation.api.v1 import stellarstation_pb2_grpc

# As defined in stellarstation.proto > message 'Plan' > enum Status
class PlanStatus(Enum):
    RESERVED = 0
    EXECUTING = 1
    SUCCEEDED = 2
    FAILED = 3
    CANCELED = 4
    PROCESSING = 5

# As defined in monitoring.proto > message 'PlanLifecycleEvent' > enum Status
class PlanLifecycleEventStatus(Enum):
    UNKNOWN = 0
    PREPARING = 1
    EXECUTING = 2
    COMPLETED = 3
    FAILED = 4

def get_grpc_client(api_key_path, api_url_path):
    print('API Target: ', api_url_path)
    jwt_credentials = google_auth_jwt.Credentials.from_service_account_file(
        api_key_path,
        audience=api_url_path,
        token_lifetime=60)
    
    google_jwt_credentials = google_auth_jwt.OnDemandCredentials.from_signing_credentials(jwt_credentials)

    # Increase grpc msg size allowance:
    options = [('grpc.max_send_message_length', 512 * 1024 * 1024),
               ('grpc.max_receive_message_length', 512 * 1024 * 1024)]
    
    channel = google_auth_transport_grpc.secure_authorized_channel(
            google_jwt_credentials,
            None,
            api_url_path,
            options = options)

    client = stellarstation_pb2_grpc.StellarStationServiceStub(channel)

    return client
