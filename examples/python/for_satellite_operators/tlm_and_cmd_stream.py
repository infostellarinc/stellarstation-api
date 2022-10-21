# Copyright 2022 Infostellar, Inc.
# Opens a stream to both receive telemetry and send commands.

from datetime import datetime
from time import sleep
from queue import Queue

from stellarstation.api.v1 import stellarstation_pb2

import toolkit
import MY_CONFIG

def generate_request(request_queue):
    request_ctr = 0

    while True:
        sleep(0.1)

        if not request_queue.empty():
            request_ctr += 1
            # print("Sent {} requests".format(request_ctr), end="\r")
            yield request_queue.get()

def run():
    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(MY_CONFIG.API_KEY_PATH, MY_CONFIG.SSL_CA_CERT_PATH)

    # Set up for stream
    request_queue = Queue()
    request_generator = generate_request(request_queue)
    tlm_file = open("tlm_and_cmd_stream_example_tlm.bin", "wb")
    total_req, total_res, total_tlms, total_acks, total_cmds, total_streamevents = 0, 0, 0, 0, 0, 0

    # Create init request and add to queue
    stream_config_request = stellarstation_pb2.SatelliteStreamRequest(
        satellite_id = str(MY_CONFIG.SATELLITE_ID),
        enable_events = True,
        enable_flow_control = True)
    request_queue.put(stream_config_request)

    # Queue a burst of dummy commands (You can send many in a single request - in this case I am sending 10 commands 0xAA)
    request = stellarstation_pb2.SatelliteStreamRequest(
        satellite_id = str(MY_CONFIG.SATELLITE_ID),
        send_satellite_commands_request = stellarstation_pb2.SendSatelliteCommandsRequest(
            command = [bytes.fromhex("AABBCCDDEEFF")] * 10,
            channel_set_id = str(MY_CONFIG.CH_SET_ID)))
    request_queue.put(request)

    print("Starting stream for Satellite ID ({}), Channel ID ({}); {}".format(MY_CONFIG.SATELLITE_ID, MY_CONFIG.CH_SET_ID, datetime.now()))

    # Process responses
    stop_streaming_critera = [toolkit.PlanLifecycleEventStatus.COMPLETED, toolkit.PlanLifecycleEventStatus.FAILED]
    plan_status = None
    for response in client.OpenSatelliteStream(request_generator):
        total_res += 1
        
        # It is necessary to acknowledge telemetry messages
        if response.HasField("receive_telemetry_response"):
            total_tlms += 1

            ack_req = stellarstation_pb2.SatelliteStreamRequest(
                satellite_id = str(MY_CONFIG.SATELLITE_ID),
                telemetry_received_ack = stellarstation_pb2.ReceiveTelemetryAck(message_ack_id = response.receive_telemetry_response.message_ack_id))
            
            request_queue.put(ack_req)

            for tlm in response.receive_telemetry_response.telemetry:
                tlm_file.write(tlm.data)
        elif response.HasField("stream_event"):
            total_streamevents += 1

            try:
                plan_status = toolkit.PlanLifecycleEventStatus(response.stream_event.plan_monitoring_event.ground_station_event.plan.status)
            except:
                pass
        
        print("Plan Status = {}; Responses = {}: Tlm = {}, StreamEvents = {}".format(plan_status.name, total_res, total_tlms, total_streamevents), end="\r")
        
        if plan_status in stop_streaming_critera:
            break
    
    print()
    print("Ending stream; {}".format(datetime.now()))

if __name__ == '__main__':
    run()