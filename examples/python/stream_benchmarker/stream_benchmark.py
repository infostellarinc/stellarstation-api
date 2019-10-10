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

import argparse
import datetime
import grpc
import os
import threading
import time
from datetime import timedelta
from google.auth import jwt as google_auth_jwt
from google.auth.transport import grpc as google_auth_transport_grpc
from queue import Queue
from threading import Lock
from threading import Timer
from stellarstation.api.v1 import stellarstation_pb2
from stellarstation.api.v1 import stellarstation_pb2_grpc
from stellarstation.api.v1 import transport_pb2 as transport

class RepeatTimer(Timer):
    def run(self):
        while not self.finished.wait(self.interval):
            self.function(*self.args, **self.kwargs)

class Metric:
    def __init__(self, time_first_byte_received, time_last_byte_received, data_size, metric_time):
        self.time_first_byte_received = time_first_byte_received
        self.time_last_byte_received = time_last_byte_received
        self.data_size = data_size
        self.metric_time = metric_time

class OutputDetails:
    def __init__(self, time_now, most_recent_time_last_byte_received, average_first_byte_latency, 
                average_last_byte_latency, total_data_size, metrics_count, average_data_size, mbps, out_of_order_data):
        self.time_now = time_now
        self.most_recent_time_last_byte_received = most_recent_time_last_byte_received
        self.average_first_byte_latency = average_first_byte_latency
        self.average_last_byte_latency = average_last_byte_latency
        self.total_data_size = total_data_size
        self.metrics_count = metrics_count
        self.average_data_size = average_data_size
        self.mbps = mbps
        self.out_of_order_data = out_of_order_data

class MetricsData:
    def __init__(self):
        self.initial_time = datetime.datetime.utcnow()
        self.metrics = []
        self.lock = Lock()
        self.most_recent_time_last_byte_received = datetime.datetime.min

    def add_metric(self, metric):
        self.lock.acquire()
        try:
            self.metrics.append(metric)
        finally:
            self.lock.release()

    def set_initial_time(self, initial_time):
        self.lock.acquire()
        try:
            self.initial_time = initial_time
        finally:
            self.lock.release()

    def compile_summary(self):
        metrics = []
        initial_time = datetime.datetime.utcnow()

        self.lock.acquire()
        try:
            metrics = self.metrics
            initial_time = self.initial_time
            self.metrics = []
            self.initial_time = datetime.datetime.utcnow()
        finally:
            self.lock.release()

        
        output_details = compile_details(metrics, initial_time, self.most_recent_time_last_byte_received)
        self.most_recent_time_last_byte_received = output_details.most_recent_time_last_byte_received
        return output_details

class Printer:
    def __init__(self, output_directory):
        self.file_out = None
        if output_directory:
            if not os.path.exists(output_directory):
                os.makedirs(output_directory)

            output_file = output_directory + "/benchmark-" + \
                datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S" + ".txt")
            self.file_out = open(output_file, "a")
    
    def print(self, line):
        print(line)
        if self.file_out:
            self.file_out.write(line)
            self.file_out.write('\n')

    def print_header(self):
            header = "DATE\tMost recent received\tAvg seconds to last byte\t" + \
                    "Avg seconds to first byte\tTotal bytes\tNum messages\tAvg bytes\tMbps\tOut of order count"
            self.print(header)

    def print_summary(self, metrics_data):
        output_details = metrics_data.compile_summary()

        outLine = '\t'.join([str(output_details.time_now),
                            str(output_details.most_recent_time_last_byte_received),
                            str(output_details.average_first_byte_latency),
                            str(output_details.average_last_byte_latency),
                            str(output_details.total_data_size),
                            str(output_details.metrics_count),
                            str(output_details.average_data_size),
                            str(output_details.mbps),
                            str(output_details.out_of_order_data)])
        self.print(outLine)

    def close(self):
        if self.file_out:
            self.file_out.close()

def run(credentials, environment, satelliteId, interval, print_summary):
    # Setup the gRPC client.
    jwt_creds = google_auth_jwt.OnDemandCredentials.from_signing_credentials(
        credentials)
    channel = google_auth_transport_grpc.secure_authorized_channel(
        jwt_creds, None, environment)
    client = stellarstation_pb2_grpc.StellarStationServiceStub(channel)

    # Open satellite stream
    request_queue = Queue()
    request_iterator = generate_request(request_queue, satelliteId)

    # Initialize variables
    metrics_data = MetricsData()
    done = False
    got_first_time = False

    timer = RepeatTimer(int(interval), print_summary, args=(metrics_data,))

    while not done:
        print('Listening for messages')
        try:
            for response in client.OpenSatelliteStream(request_iterator):
                if response.HasField("receive_telemetry_response"):
                    num_bytes_in_message = len(response.receive_telemetry_response.telemetry.data)
                    if not got_first_time:
                        print('Receiving messages')
                        got_first_time = True
                        metrics_data.set_initial_time(datetime.datetime.utcnow())
                        timer.start()
                    metric = Metric(response.receive_telemetry_response.telemetry.time_first_byte_received,
                         response.receive_telemetry_response.telemetry.time_last_byte_received,
                         num_bytes_in_message,
                            datetime.datetime.utcnow())
                    metrics_data.add_metric(metric)

        # When CTRL-C is pressed the benchmark ends
        except KeyboardInterrupt:
            done = True
        except Exception as e:
            print("Encountered error")
            print(e)
            break
        finally:
            timer.cancel()



def compile_details(metrics, initial_time, most_recent_time_last_byte_received):
    now = datetime.datetime.utcnow()
    
    average_first_byte_latency = timedelta()
    average_last_byte_latency = timedelta()
    average_data_size = 0

    total_first_byte_latency = timedelta()
    total_last_byte_latency = timedelta()
    total_data_size = 0

    metrics_count = 0
    out_of_order_data = 0
    previous_first_byte_time = datetime.datetime.min

    for metric in metrics:
        first_byte_time = metric.time_first_byte_received.ToDatetime()
        last_byte_time = metric.time_last_byte_received.ToDatetime()
        data_size = metric.data_size
        metric_time = metric.metric_time

        first_byte_latency = metric_time - first_byte_time
        last_byte_latency = metric_time - last_byte_time

        total_first_byte_latency += first_byte_latency
        total_last_byte_latency += last_byte_latency
        total_data_size += data_size
        metrics_count += 1
        
        if metric.time_last_byte_received.ToDatetime() > most_recent_time_last_byte_received:
            most_recent_time_last_byte_received = metric.time_last_byte_received.ToDatetime()

        if previous_first_byte_time > first_byte_time:
            out_of_order_data += 1

        previous_first_byte_time = first_byte_time

    if metrics_count > 0:
        average_first_byte_latency = total_first_byte_latency / metrics_count
        average_last_byte_latency = total_last_byte_latency / metrics_count
        average_data_size = total_data_size / metrics_count

    mbps = (total_data_size / (now - initial_time).seconds) * 8 / 1024 / 1024

    outputDetails = OutputDetails(datetime.datetime.utcnow(),
                                most_recent_time_last_byte_received, 
                                average_first_byte_latency, 
                                average_last_byte_latency, 
                                total_data_size, 
                                metrics_count, 
                                average_data_size, 
                                mbps,
                                out_of_order_data)

    return outputDetails


# This generator yields the requests to send on the stream opened by OpenSatelliteStream.
# The client side of the stream will be closed when this generator returns
# (in this example, it never returns).

def generate_request(queue, satelliteId):
    # Send the first request to activate the stream. Telemetry will start
    # to be received at this point.
    satellite_stream_request = stellarstation_pb2.SatelliteStreamRequest(
        satellite_id=satelliteId)
    satellite_stream_request.accepted_framing.extend(
        [transport.AX25, transport.BITSTREAM])
    yield satellite_stream_request

    try:
        while True:
            commands = queue.get()
            command_request = stellarstation_pb2.SendSatelliteCommandsRequest(
                command=commands)

            satellite_stream_request = stellarstation_pb2.SatelliteStreamRequest(
                satellite_id=satelliteId, send_satellite_commands_request=command_request)

            satellite_stream_request.accepted_framing.extend(
                [transport.AX25, transport.BITSTREAM])

            yield satellite_stream_request
            queue.task_done()
            while True:
                time.sleep(3)
    except Exception as e:
        print("Encountered error")
        print(e)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-int",
                        "--interval",
                        help="reporting interval in seconds",
                        default="10")
    parser.add_argument("-i",
                        "--id",
                        help="satellite id",
                        default="5")
    parser.add_argument("-k",
                        "--key",
                        help="API key file",
                        default="stellarstation-private-key.json")
    parser.add_argument("-e",
                        "--endpoint",
                        help="API endpoint",
                        default="api.stellarstation.com:443")
    parser.add_argument("-d",
                        "--directory",
                        help="output directory",
                        default="")
    args = parser.parse_args()

    # Load the private key downloaded from the StellarStation Console.
    credentials = google_auth_jwt.Credentials.from_service_account_file(
        args.key,
        audience="https://api.stellarstation.com")
    
    printer = Printer(args.directory)
    try:
        printer.print_header()
        run(credentials, args.endpoint, args.id, args.interval, printer.print_summary)
    finally:
        printer.close()


if __name__ == '__main__':
    main()
