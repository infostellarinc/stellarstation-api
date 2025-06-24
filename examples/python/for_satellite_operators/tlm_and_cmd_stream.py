# Copyright 2023 Infostellar, Inc.
# Opens a stream to both receive telemetry and send commands.

import os
from datetime import datetime
from queue import Queue
from time import sleep
import grpc

from stellarstation.api.v1 import stellarstation_pb2
from google.protobuf.timestamp_pb2 import Timestamp

import toolkit

STREAM_DONE = object()

def generate_request(request_queue):
    # iter(request_queue.get) will block until a message is sent into the queue
    # we break from the loop by listening for STREAM_DONE to be sent into the queue
    for request in iter(request_queue.get, None):
        if request is STREAM_DONE:
            break
        yield request

def run():
    STELLARSTATION_API_KEY_PATH = os.getenv('STELLARSTATION_API_KEY_PATH')
    STELLARSTATION_API_SATELLITE_ID = os.getenv('STELLARSTATION_API_SATELLITE_ID')
    STELLARSTATION_API_CHANNEL_ID = os.getenv('STELLARSTATION_API_CHANNEL_ID')

    assert STELLARSTATION_API_KEY_PATH, "Did you properly define this environment variable on your system?"
    assert STELLARSTATION_API_SATELLITE_ID, "Did you properly define this environment variable on your system?"
    assert STELLARSTATION_API_CHANNEL_ID, "Did you properly define this environment variable on your system?"
    
    STELLARSTATION_API_URL = os.getenv('STELLARSTATION_API_URL','stream.qa.stellarstation.com')
    assert STELLARSTATION_API_URL, "Did you properly define this environment variable on your system?"

    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(STELLARSTATION_API_KEY_PATH, STELLARSTATION_API_URL)

    # Set up for stream
    tlm_file = open("tlm_and_cmd_stream_example_tlm.bin", "wb")
    total_responses = 0
    total_telemetry_messages = 0
    total_stream_events = 0
    total_acks_sent = 0
    total_messages_sent = 0
    total_bytes_received = 0
    command_request_count = 0
    stream_id = None
    last_ack_id = None

    # All messages to the streamer will go through this queue.
    request_queue = Queue()
    request_generator = generate_request(request_queue)

    # Process responses
    stop_streaming_critera = [toolkit.PlanLifecycleEventStatus.FAILED]
    plan_status = toolkit.PlanLifecycleEventStatus.UNKNOWN
    end_message_received = False
    stream_attempts = 0

    # Running in a loop to show how you can reconnect
    # if there's a problem with GRPC/Streamer
    #
    # if we've received the end of the telemetry data, the plan fails,
    # or we've tried 3 times we'll stop
    while end_message_received is False and \
            plan_status not in stop_streaming_critera and \
            stream_attempts < 3:
        stream_attempts += 1

        # We'll prepare the initial setup request and add to queue
        #
        #
        # enable_events, enable_flow_control do not need to be sent
        # after the setup message.
        stream_setup_request = stellarstation_pb2.SatelliteStreamRequest(
            # satellite_id is required for every message sent
            satellite_id=STELLARSTATION_API_SATELLITE_ID,
            # enable events allow stream events to be received
            enable_events=True,

            # These are for stream recovery.
            #
            # stream_id and last_ack_id will be None in the first loop,
            # but will let us attempt to recover the stream if an error occurs
            stream_id=stream_id,
            # On recovery, streamer will rewind the stream to the message after last_ack_id
            resume_stream_message_ack_id=last_ack_id,
            # This is required in order to do stream recovery
            # It also helps us verify data is received by your client
            enable_flow_control=True

            # If you have the plan id, it'll limit your stream to only data for that plan
            # plan_id=plan_id

            # if you have the groundstation id, it'll limit your stream to only data from that groundstation
            # groundstation_id=groundstation_id
        )
        request_queue.put(stream_setup_request)
        total_messages_sent += 1

        if command_request_count == 0:
            # Queue a burst of dummy commands
            # You can send many in a single request
            # in this case we send 10 commands
            command_request = stellarstation_pb2.SatelliteStreamRequest(
                # The groundstation will try to respond with this request_id
                # to confirm the command was sent.
                # Using a UUID is recommended.
                request_id="command_request_id_{}".format(
                    command_request_count),
                satellite_id=STELLARSTATION_API_SATELLITE_ID,
                send_satellite_commands_request=stellarstation_pb2.SendSatelliteCommandsRequest(
                    command=[bytes.fromhex("AABBCCDDEEFF")] * 10,
                    channel_set_id=STELLARSTATION_API_SATELLITE_ID))

            request_queue.put(command_request)
            command_request_count += 1
            total_messages_sent += 1

        print("Starting stream for Satellite ID ({}), Channel ID ({}); {}".format(
            STELLARSTATION_API_SATELLITE_ID, STELLARSTATION_API_CHANNEL_ID, datetime.now()))

        try:
            # OpenSatelliteStream will start the stream,
            # all messages received will come as a response
            # all messages we want to send go through request_queue and request_generator
            for response in client.OpenSatelliteStream(request_generator):
                total_responses += 1

                # stream_id allows you to attempt a stream recovery, but
                # also provides a useful identifier for the Stellarstation
                # team to help debug any issues
                if stream_id is None and len(response.stream_id) > 0:
                    stream_id = response.stream_id

                # check if we received telemetry or a stream event
                if response.HasField("receive_telemetry_response"):
                    total_telemetry_messages += 1

                    # First we'll send an ack that we received the message
                    # Acks are required to verify your client has received the data
                    ack_request = stellarstation_pb2.SatelliteStreamRequest(
                        satellite_id=STELLARSTATION_API_SATELLITE_ID,
                        telemetry_received_ack=stellarstation_pb2.ReceiveTelemetryAck(
                            message_ack_id=response.receive_telemetry_response.message_ack_id,
                            # received_timestamp is not required,
                            # but provides stellarstation with debugging information
                            received_timestamp=Timestamp().GetCurrentTime()
                        ))

                    request_queue.put(ack_request)
                    last_ack_id = response.receive_telemetry_response.message_ack_id
                    total_messages_sent += 1
                    total_acks_sent += 1

                    # Record the telemetry to file
                    for tlm in response.receive_telemetry_response.telemetry:
                        total_bytes_received += len(tlm.data)
                        tlm_file.write(tlm.data)

                    # A message with 1 telemetry and 0 data is a way
                    # we mark the End message. This may change in
                    # the future for clearer detection for the end of a stream
                    #
                    # This message is sent when the groundstation is "cleaning up"
                    # So it may arrive after the PlanLifecycleEventStatus
                    # is marked as complete
                    if len(response.receive_telemetry_response.telemetry) == 1 and len(response.receive_telemetry_response.telemetry[0].data) == 0:
                        end_message_received = True

                elif response.HasField("stream_event"):
                    total_stream_events += 1

                    try:
                        # There are various types of stream events
                        # There's monitoring events as well as life cycle events
                        # Here we're looking for plan status updates
                        if response.stream_event.HasField("plan_monitoring_event") and \
                                response.stream_event.plan_monitoring_event.HasField("ground_station_event"):
                            plan_status = toolkit.PlanLifecycleEventStatus(
                                response.stream_event.plan_monitoring_event.ground_station_event.plan.status)
                    except:
                        pass

                print("Plan Status = {}: Total Responses = {}, Telemetry Messages = {}, MessagesSent = {}, Acks Sent = {}, StreamEvents = {}, Total Bytes = {}".format(
                    plan_status.name,
                    total_responses,
                    total_telemetry_messages,
                    total_messages_sent,
                    total_acks_sent,
                    total_stream_events,
                    total_bytes_received
                ), end="\r")

                if plan_status in stop_streaming_critera or end_message_received:
                    break
        except grpc.RpcError as e:
            print("GRPC error while streaming: {}".format(e))
            # Sleep before retrying
            sleep(1)
        except Exception as e:
            print("Unhandled error while streaming: {}".format(e))
            # unknown exceptions won't be retried
            break

    # Send STREAM_DONE so the request generator can shut down
    request_queue.put(STREAM_DONE)
    print()
    print("Ending stream (id = {}): total bytes = {}, finished at = {}".format(
        stream_id, total_bytes_received, datetime.now()))


if __name__ == '__main__':
    run()
