# Copyright 2022 Infostellar, Inc.

# A nice set of tools used by the examples.

from enum import Enum

from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc
from stellarstation.api.v1 import stellarstation_pb2_grpc
import grpc

# The ID of your satellite as it exists in StellarStation
SATELLITE_ID = '300'

# The path to your API key obtained from StellarStation console
API_KEY_PATH = '/media/sf_stellarstation_vm_shared/api_keys/proffitt-prod-sacatapult-key.json'

# The path to your machine's SSL certificates
SSL_CA_CERT_PATH = "/etc/ssl/certs/ca-certificates.crt"

# As defined in stellarstation.proto > message 'Plan' > enum Status
class Status(Enum):
    RESERVED = 0
    EXECUTING = 1
    SUCCEEDED = 2
    FAILED = 3
    CANCELED = 4
    PROCESSING = 5

def get_grpc_client(api_key_path, ssl_ca_certificate_path):
    jwt_credentials = google_auth_jwt.Credentials.from_service_account_file(
        api_key_path,
        audience='https://api.stellarstation.com',
        token_lifetime=60)
    
    ca = open(ssl_ca_certificate_path, 'rb')
    ssl_channel_credentials = ca.read()
    
    google_jwt_credentials = google_auth_jwt.OnDemandCredentials.from_signing_credentials(jwt_credentials)

    # Increase grpc msg size allowance:
    options = [('grpc.max_send_message_length', 512 * 1024 * 1024),
               ('grpc.max_receive_message_length', 512 * 1024 * 1024)]
    
    channel = google_auth_transport_grpc.secure_authorized_channel(
            google_jwt_credentials,
            None,
            'api.stellarstation.com:443',
            ssl_credentials=grpc.ssl_channel_credentials(ssl_channel_credentials),
            options = options)

    client = stellarstation_pb2_grpc.StellarStationServiceStub(channel)

    return client
