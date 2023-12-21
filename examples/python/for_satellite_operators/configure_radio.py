# Copyright 2022 Infostellar, Inc.
# Opens a stream to configure the transceiver and/or receiver.
# Only requests within an active pass (AoS to LoS) will work.

import os
from time import sleep
from queue import Queue
import threading

from google.protobuf.wrappers_pb2 import FloatValue, BoolValue

from consolemenu import *
from consolemenu.items import *

from stellarstation.api.v1 import stellarstation_pb2
from stellarstation.api.v1.radio.radio_pb2 import AM, FM, PM, PCM_PSK_PM, PCM_PM_BI_PHASE_L
from stellarstation.api.v1.radio.radio_pb2 import BPSK, QPSK, OQPSK, PSK8, PSK16, QAM16, APSK16, MFSK, AFSK, FSK

import toolkit

class ConfigurationRequest():
    PROTO_TRUE = BoolValue(value = True)
    PROTO_FALSE = BoolValue(value = False)

    def __init__(self):
        pass
    
    def _set_bitrate(self, bitrate):
        self.bitrate = bitrate
    
    def _set_modulation_type(self, mod_type):
        self.modulation_type = mod_type
    
    def prompt_set_bitrate(self):
        bad_prompt = True
        while bad_prompt:
            bitrate = input("Enter a bitrate as an integer: ")
            try:
                bitrate = int(bitrate)
                self._set_bitrate(bitrate)
                bad_prompt = False
            except:
                try_again = input("That was not an integer. Try again? (y/n): ")
                if try_again == "n":
                    return

    def prompt_choose_modulation(self):
        bad_prompt = True
        while bad_prompt:
            print("Available modulation types: {}".format(", ".join(self.MODULATION_TYPES.keys())))
            
            mod_type = input("Enter a modulation type: ")
            
            if mod_type in self.MODULATION_TYPES.keys():
                self._set_modulation_type(mod_type)
                bad_prompt = False
            else:
                try_again = input("That was not a valid modulation type. Try again? (y/n): ")
                if try_again == "n":
                    return
    
    def clear_fields(self):
        raise NotImplementedError("clear_fields")
    
    def get_str(self):
        raise NotImplementedError("get_str")
    
    def send_as_request(self):
        raise NotImplementedError("send_as_request")

class TransmitterConfigurationRequest(ConfigurationRequest):
    MODULATION_TYPES = {
        "AM":AM,
        "FM":FM,
        "PM":PM,
        "PCMPSK":PCM_PSK_PM,
        "PSMPMBI":PCM_PM_BI_PHASE_L,
    }

    def __init__(self, request_queue):
        self.type = "Transmitter"
        self.request_queue = request_queue

        self.recently_sent_request = False
        self.bitrate = None
        self.carrier = None
        self.sweep = None
        self.modulation_bool = None
        self.modulation_type = None
        self.idle_pattern = None
    
    def __str__(self):
        str_list = ["{} Configuration Request".format(self.type)]

        if not self.recently_sent_request:
            str_list.append("Currently Set Fields...")
            if self.bitrate != None: str_list.append("\tSet Bitrate: {}bps".format(self.bitrate))
            if self.carrier != None: str_list.append("\tCarrier: {}".format("ON" if self.carrier == True else "OFF"))
            if self.sweep != None: str_list.append("\tSweep: {}".format("ON" if self.sweep == True else "OFF"))
            if self.modulation_bool != None: str_list.append("\tModulation: {}".format("ON" if self.modulation_bool == True else "OFF"))
            if self.modulation_type != None: str_list.append("\tModulation Type: {}".format(self.modulation_type))
            if self.idle_pattern != None: str_list.append("\tIdle Pattern: {}".format("ON" if self.idle_pattern == True else "OFF"))
        else:
            str_list.append("Sent request!")
        
        self.recently_sent_request = False

        return "\n".join(str_list)
    
    def enable_carrier(self):
        self.carrier = True
    
    def disable_carrier(self):
        self.carrier = False
    
    def enable_sweep(self):
        self.sweep = True
    
    def disable_sweep(self):
        self.sweep = False
    
    def enable_modulation(self):
        self.modulation_bool = True
    
    def disable_modulation(self):
        self.modulation_bool = False
    
    def enable_idle_pattern(self):
        self.idle_pattern = True
    
    def disable_idle_pattern(self):
        self.idle_pattern = False
    
    def clear_fields(self):
        self.recently_sent_request = False
        self.bitrate = None
        self.carrier = None
        self.sweep = None
        self.modulation_bool = None
        self.modulation_type = None
        self.idle_pattern = None
    
    def get_str(self):
        return str(self)
    
    def send_as_request(self):
        config = stellarstation_pb2.TransmitterConfigurationRequest()
        
        if self.bitrate:
            config.bitrate.CopyFrom(FloatValue(value = float(self.bitrate)))

        if self.carrier == True:
            config.enable_carrier.CopyFrom(ConfigurationRequest.PROTO_TRUE)
        elif self.carrier == False:
            config.enable_carrier.CopyFrom(ConfigurationRequest.PROTO_FALSE)
        
        if self.sweep == True:
            config.enable_if_sweep.CopyFrom(ConfigurationRequest.PROTO_TRUE)
        elif self.sweep == False:
            config.enable_if_sweep.CopyFrom(ConfigurationRequest.PROTO_FALSE)
        
        if self.modulation_bool == True:
            config.enable_if_modulation.CopyFrom(ConfigurationRequest.PROTO_TRUE)
        elif self.modulation_bool == False:
            config.enable_if_modulation.CopyFrom(ConfigurationRequest.PROTO_FALSE)
        
        if self.idle_pattern == True:
            config.enable_idle_pattern.CopyFrom(ConfigurationRequest.PROTO_TRUE)
        elif self.idle_pattern == False:
            config.enable_idle_pattern.CopyFrom(ConfigurationRequest.PROTO_FALSE)
        
        if self.modulation_type != None:
            try:
                config.modulation = self.MODULATION_TYPES[self.modulation_type]
            except:
                raise ValueError("Bad modulation setting ({} not in {})".format(self.modulation_type, self.MODULATION_TYPES))
        
        gs_config_request = stellarstation_pb2.GroundStationConfigurationRequest(transmitter_configuration_request = config)

        self.clear_fields()

        self.recently_sent_request = True

        self.request_queue.put(gs_config_request)

class ReceiverConfigurationRequest(ConfigurationRequest):
    MODULATION_TYPES = {
        "BPSK":BPSK,
        "QPSK":QPSK,
        "OQPSK":OQPSK,
        "PSK8":PSK8,
        "PSK16":PSK16,
        "QAM16":QAM16,
        "APSK16":APSK16,
        "MFSK":MFSK,
        "AFSK":AFSK,
        "FSK":FSK,
    }

    def __init__(self, request_queue):
        self.type = "Receiver"
        self.request_queue = request_queue

        self.recently_sent_request = False
        self.bitrate = None
        self.modulation_type = None
    
    def __str__(self):
        str_list = ["{} Configuration Request".format(self.type)]

        if not self.recently_sent_request:
            str_list.append("Currently Set Fields...")
            if self.bitrate != None: str_list.append("\tSet Bitrate: {}bps".format(self.bitrate))
            if self.modulation_type != None: str_list.append("\tModulation Type: {}".format(self.modulation_type))
        else:
            str_list.append("Sent request!")
        
        self.recently_sent_request = False

        return "\n".join(str_list)

    def clear_fields(self):
        self.recently_sent_request = False
        self.bitrate = None
        self.modulation_type = None
    
    def get_str(self):
        return str(self)
    
    def send_as_request(self):
        config = stellarstation_pb2.ReceiverConfigurationRequest()
        
        if self.bitrate:
            config.bitrate.CopyFrom(FloatValue(value = float(self.bitrate)))
        
        if self.modulation_type != None:
            try:
                config.modulation = self.MODULATION_TYPES[self.modulation_type]
            except:
                raise ValueError("Bad modulation setting ({} not in {})".format(self.modulation_type, self.MODULATION_TYPES))
        
        gs_config_request = stellarstation_pb2.GroundStationConfigurationRequest(receiver_configuration_request = config)

        self.clear_fields()

        self.recently_sent_request = True

        self.request_queue.put(gs_config_request)

def generate_request(request_queue, thread_sts_queue):
    while True:
        sleep(0.1)

        if not request_queue.empty():
            req = request_queue.get()
            if req == None:
                thread_sts_queue.put("Received flag to shut down stream thread. Breaking from generate_request...")
                break
            else:
                # thread_sts_queue.put("Sent Request: {}".format(str(req)))
                yield req

def run_streamer(api_key_path, api_url_path, request_queue, thread_sts_queue):
    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(api_key_path, api_url_path)

    request_generator = generate_request(request_queue, thread_sts_queue)

    try:
        for response in client.OpenSatelliteStream(request_generator):
            # thread_sts_queue.put("Received response: {}".format(response))
            pass
    except:
        thread_sts_queue.put("Shutting down streamer thread.")

def run():
    STELLARSTATION_API_KEY_PATH = os.getenv('STELLARSTATION_API_KEY_PATH')
    STELLARSTATION_API_SATELLITE_ID = os.getenv('STELLARSTATION_API_SATELLITE_ID')

    assert STELLARSTATION_API_KEY_PATH, "Did you properly define this environment variable on your system?"
    assert STELLARSTATION_API_SATELLITE_ID, "Did you properly define this environment variable on your system?"

    STELLARSTATION_API_URL = os.getenv('STELLARSTATION_API_URL','stream.qa.stellarstation.com')
    assert STELLARSTATION_API_URL, "Did you properly define this environment variable on your system?"
    
    request_queue = Queue()
    thread_sts_queue = Queue()

    stream_config_request = stellarstation_pb2.SatelliteStreamRequest(
        satellite_id = STELLARSTATION_API_SATELLITE_ID,
        enable_events = True,
        enable_flow_control = True)
    request_queue.put(stream_config_request)

    streamer_thread = threading.Thread(target=run_streamer, args=(STELLARSTATION_API_KEY_PATH, STELLARSTATION_API_URL, request_queue, thread_sts_queue,), daemon=True)
    streamer_thread.start()

    main_menu = ConsoleMenu("Main Menu (Radio Configuration)", "This example code exhibits a CLI that allows the user to build and send transceiver configuration commands.")

    # Transmitter
    outgoing_transmitter_req = TransmitterConfigurationRequest(request_queue)
    config_transmitter_menu = ConsoleMenu("Configure Transmitter", outgoing_transmitter_req.get_str)
    transmitter_config_menu_func_items = [
        FunctionItem("Clear Fields", outgoing_transmitter_req.clear_fields),
        FunctionItem("Set Bitrate", outgoing_transmitter_req.prompt_set_bitrate),
        FunctionItem("Enable Carrier", outgoing_transmitter_req.enable_carrier),
        FunctionItem("Disable Carrier", outgoing_transmitter_req.disable_carrier),
        FunctionItem("Enable Sweep", outgoing_transmitter_req.enable_sweep),
        FunctionItem("Disable Sweep", outgoing_transmitter_req.disable_sweep),
        FunctionItem("Enable Modulation", outgoing_transmitter_req.enable_modulation),
        FunctionItem("Disable Modulation", outgoing_transmitter_req.disable_modulation),
        FunctionItem("Choose Modulation Type", outgoing_transmitter_req.prompt_choose_modulation),
        FunctionItem("Enable Idle Pattern", outgoing_transmitter_req.enable_idle_pattern),
        FunctionItem("Disable Idle Pattern", outgoing_transmitter_req.disable_idle_pattern),
        FunctionItem("Send Request", outgoing_transmitter_req.send_as_request),
    ]
    for func_item in transmitter_config_menu_func_items:
        config_transmitter_menu.append_item(func_item)   
    config_transmitter_submenu = SubmenuItem("Configure Transmitter", config_transmitter_menu, main_menu)
    
    # Receiver
    outgoing_receiver_req = ReceiverConfigurationRequest(request_queue)
    config_receiver_menu = ConsoleMenu("Configure Receiver", outgoing_receiver_req.get_str)
    receiver_config_menu_func_items = [
        FunctionItem("Clear Fields", outgoing_receiver_req.clear_fields),
        FunctionItem("Set Bitrate", outgoing_receiver_req.prompt_set_bitrate),
        FunctionItem("Choose Modulation Type", outgoing_receiver_req.prompt_choose_modulation),
        FunctionItem("Send Request", outgoing_receiver_req.send_as_request),
    ]
    for func_item in receiver_config_menu_func_items:
        config_receiver_menu.append_item(func_item)
    config_receiver_submenu = SubmenuItem("Configure Receiver", config_receiver_menu, main_menu)

    main_menu.append_item(config_transmitter_submenu)
    main_menu.append_item(config_receiver_submenu)
    main_menu.show()

    request_queue.put(None)

    print("Shutting down CLI and stream...")

    streamer_thread.join(timeout=5.0)

    # print("\nvvv Debugging - Status Updates from Streaming Thread vvv")
    # while not thread_sts_queue.empty():
    #     sts = thread_sts_queue.get()
    #     print(sts)

if __name__ == '__main__':
    run()