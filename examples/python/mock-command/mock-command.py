# Copyright 2022 Infostellar, Inc.
# This example opens stream for the provided satellite ID. It also sends a fixed mock command (CMD)
# to the radio transmitter every time a TLM message is received.

import argparse
from datetime import datetime
import time
from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc
from stellarstation.api.v1 import stellarstation_pb2
from stellarstation.api.v1 import stellarstation_pb2_grpc
import grpc
from queue import Queue

def main():

    #########    cli flags    #########

    parser = argparse.ArgumentParser()
    parser.add_argument("-i",
                        "--id",
                        help="Satellite ID from the StellarStation system",
                        default="5")
    parser.add_argument("-ch",
                        "--channelset",
                        help="Channel Set name or ID from the StellarStation system",
                        default="test")
    parser.add_argument("-k",
                        "--key",
                        help="API key file path",
                        default="./api-key.json")
    parser.add_argument("-e",
                        "--endpoint",
                        help="API endpoint",
                        default='api.stellarstation.com:443')
    parser.add_argument("-c",
                        "--sslcert",
                        help="SSL CA Certificate path",
                        default='./tls.crt')
    args = parser.parse_args()

    SATELLITE_ID = args.id
    CH_SET_ID = args.channelset
    API_KEY_PATH = args.key
    END_POINT = args.endpoint
    SSL_CA_CERT_PATH = args.sslcert

    MEGA_BYTE = 1024 ** 2

    #########    gRPC client connection    #########
    
    # Load the private key downloaded from the StellarStation Console:
    credentials = google_auth_jwt.Credentials.from_service_account_file(
        API_KEY_PATH,
        audience='https://api.stellarstation.com')

    ca = open(SSL_CA_CERT_PATH, 'rb')
    creds = ca.read()
    ssl_credentials=grpc.ssl_channel_credentials(creds)

    # Setup the gRPC client:
    jwt_creds = google_auth_jwt.OnDemandCredentials.from_signing_credentials(credentials)
    # Increase grpc msg size allowance:
    options = [('grpc.max_send_message_length', 512 * MEGA_BYTE),
               ('grpc.max_receive_message_length', 512 * MEGA_BYTE)]
    
    channel = google_auth_transport_grpc.secure_authorized_channel(
            jwt_creds,
            None,
            END_POINT,
            ssl_credentials,
            options = options)
    
    client = stellarstation_pb2_grpc.StellarStationServiceStub(channel)


    ######### Satellite telemetry stream reception & CMD transmission #########
        
    # Queues (for sending info to the iterator) and iterator creation:
    ack_queue = Queue()
    cmd_queue = Queue()
    request_iterator = generate_request(ack_queue, cmd_queue, SATELLITE_ID, CH_SET_ID)
    
    for response in client.OpenSatelliteStream(request_iterator): 
     
        if response.HasField("receive_telemetry_response"):
            
            # Reads the ack_id of every response and injects the value to request_queue that the iterator will lookup:
            ACK_ID = response.receive_telemetry_response.message_ack_id  
            received_ack = stellarstation_pb2.ReceiveTelemetryAck(message_ack_id = ACK_ID)
            ack_queue.put(received_ack)
            """
            print("Got telemetry_response_message with number of telemetry items = ",
                  len(response.receive_telemetry_response.telemetry),
                  "ACK_ID: ", ACK_ID)
            """
        
        # Here, for every TLM response received, a mock CMD is created to be sent back:
        # The mock CMD is eb900123456789abcd0123456789abcdef0000000055555534c5c5c5c5c5c5c579
        _command = [bytes(b'\xeb\x90\x01\x23\x45\x67\x89\xab\xcd\xef\x01\x23\x45\x67\x89\xab\xcd\xef\x00\x00\x00\x00\x55\x55\x55\x34\xc5\xc5\xc5\xc5\xc5\xc5\xc5\x79')]
        cmd_queue.put(_command)


# This generator yields the requests to send on the stream opened by OpenSatelliteStream.
def generate_request(ack_queue, cmd_queue, sat_id, chan_id):

    # Send the first request to activate the stream. Telemetry will start to be received at this point.
    # It is recommended to start this at least 2 minutes before the first telemetry packet is expected.
    print(datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3] + " " +
          "Opening stream, Satellite ID = ", sat_id)
    
    yield stellarstation_pb2.SatelliteStreamRequest(
            satellite_id = sat_id,
            enable_events = True,
            enable_flow_control = True)
    
    while True:
        time.sleep(.05) # This timer should be adjusted to the datarate expected
        
        # Looks for any TLM ack coming through the ack_queue:
        if not ack_queue.empty():
            
            ack = ack_queue.get()
            
            yield stellarstation_pb2.SatelliteStreamRequest(
                    satellite_id = sat_id,
                    telemetry_received_ack = ack)
            
        # Looks for any CMD to be sent from the cmd_queue:         
        if not cmd_queue.empty():
        
            cmd = cmd_queue.get()
            command_request = stellarstation_pb2.SendSatelliteCommandsRequest(
                    command=cmd,
                    channel_set_id=chan_id)
            
            satellite_stream_request = stellarstation_pb2.SatelliteStreamRequest(
                satellite_id=sat_id,
                send_satellite_commands_request=command_request)
            
            print(datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3], " Sending CMD")
            yield satellite_stream_request
        
        yield stellarstation_pb2.SatelliteStreamRequest(
                satellite_id = sat_id)

if __name__ == '__main__':
    main()