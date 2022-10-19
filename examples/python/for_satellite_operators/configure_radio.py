# Copyright 2022 Infostellar, Inc.
# Opens a stream to configure the transceiver and/or receiver.
# Only requests within an active pass (AoS to LoS) will work.

from time import sleep
from queue import Queue
import threading
from urllib import request

from google.protobuf.wrappers_pb2 import FloatValue, BoolValue

from consolemenu import *
from consolemenu.items import *

from stellarstation.api.v1 import stellarstation_pb2
from stellarstation.api.v1.radio.radio_pb2 import AM, FM, PM, PCM_PSK_PM, PCM_PM_BI_PHASE_L
from stellarstation.api.v1.radio.radio_pb2 import BPSK, QPSK, OQPSK, PSK8, PSK16, QAM16, APSK16, MFSK, AFSK, FSK

import toolkit
import MY_CONFIG

class GsConfigRequest():
    TRANSMITTER_MODULATION_TYPES = {
        "AM":AM,
        "FM":FM,
        "PM":PM,
        "PCMPSK":PCM_PSK_PM,
        "PSMPMBI":PCM_PM_BI_PHASE_L,
    }

    RECEIVER_MODULATION_TYPES = {
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

    def __init__(self, transmitter_or_receiver):
        assert (transmitter_or_receiver == "transmitter" or transmitter_or_receiver == "receiver")

        self.transmitter_or_receiver = transmitter_or_receiver
        self.bitrate = None
        self.carrier = None
        self.sweep = None
        self.modulation_bool = None
        self.modulation_type = None
        self.idle_pattern = None

        self.available_modulation_types = None
        if self.transmitter_or_receiver == "transmitter":
            self.available_modulation_types = GsConfigRequest.TRANSMITTER_MODULATION_TYPES
        else:
            self.available_modulation_types = GsConfigRequest.RECEIVER_MODULATION_TYPES

    def __str__(self):
        str_list = []
        if self.transmitter_or_receiver != None: str_list.append("{} CONFIGURATION REQUEST".format(self.transmitter_or_receiver.upper()))
        str_list.append("Currently Set Fields...")
        if self.bitrate != None: str_list.append("\tSet Bitrate: {}bps".format(self.bitrate))
        if self.carrier != None: str_list.append("\tCarrier: {}".format("ON" if self.carrier == True else "OFF"))
        if self.sweep != None: str_list.append("\tSweep: {}".format("ON" if self.sweep == True else "OFF"))
        if self.modulation_bool != None: str_list.append("\tModulation: {}".format("ON" if self.modulation_bool == True else "OFF"))
        if self.modulation_type != None: str_list.append("\tModulation Type: {}".format(self.modulation_type))
        if self.idle_pattern != None: str_list.append("\tIdle Pattern: {}".format("ON" if self.idle_pattern == True else "OFF"))

        return "\n".join(str_list)
    
    def _set_bitrate(self, bitrate):
        self.bitrate = bitrate
    
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
    
    def _set_modulation_type(self, mod_type):
        self.modulation_type = mod_type

    def prompt_choose_modulation(self):
        bad_prompt = True
        while bad_prompt:
            print("Available modulation types: {}".format(", ".join(self.available_modulation_types.keys())))
            
            mod_type = input("Enter a modulation type: ")
            
            if mod_type in self.available_modulation_types.keys():
                self._set_modulation_type(mod_type)
                bad_prompt = False
            else:
                try_again = input("That was not a valid modulation type. Try again? (y/n): ")
                if try_again == "n":
                    return
    
    def enable_idle_pattern(self):
        self.idle_pattern = True
    
    def disable_idle_pattern(self):
        self.idle_pattern = False
    
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
    
    def clear_fields(self):
        self.bitrate = None
        self.carrier = None
        self.sweep = None
        self.modulation_bool = None
        self.modulation_type = None
        self.idle_pattern = None
    
    def get_str(self):
        return str(self)
    
    def to_request(self):
        if self.transmitter_or_receiver == "transmitter":
            config = stellarstation_pb2.TransmitterConfigurationRequest()
        elif self.transmitter_or_receiver == "receiver":
            config = stellarstation_pb2.ReceiverConfigurationRequest()
        else:
            raise ValueError("self.transmitter_or_receiver parameter not set")
        
        PROTO_TRUE = BoolValue(value = True)
        PROTO_FALSE = BoolValue(value = False)
        
        if self.bitrate:
            config.bitrate.CopyFrom(FloatValue(value = float(self.bitrate)))

        if self.carrier == True:
            config.enable_carrier.CopyFrom(PROTO_TRUE)
        elif self.carrier == False:
            config.enable_carrier.CopyFrom(PROTO_FALSE)
        
        if self.sweep == True:
            config.enable_if_sweep.CopyFrom(PROTO_TRUE)
        elif self.sweep == False:
            config.enable_if_sweep.CopyFrom(PROTO_FALSE)
        
        if self.modulation_bool == True:
            config.enable_if_modulation.CopyFrom(PROTO_TRUE)
        elif self.modulation_bool == False:
            config.enable_if_modulation.CopyFrom(PROTO_FALSE)
        
        if self.idle_pattern == True:
            config.enable_idle_pattern.CopyFrom(PROTO_TRUE)
        elif self.idle_pattern == False:
            config.enable_idle_pattern.CopyFrom(PROTO_FALSE)
        
        if self.modulation_type != None:
            mod_types = GsConfigRequest.TransmitterModulationTypes if self.transmitter_or_receiver == "transmitter" else GsConfigRequest.ReceiverModulationTypes
            try:
                config.modulation = mod_types[self.modulation_type]
            except:
                raise ValueError("{} not acceptable {} configuration setting\nAcceptable settings: {}".format(
                        self.modulation_type, self.transmitter_or_receiver, ", ".join(mod_types)))
        
        if self.transmitter_or_receiver == "transmitter":
            gs_config_request = stellarstation_pb2.GroundStationConfigurationRequest(transmitter_configuration_request = config)
        else:
            gs_config_request = stellarstation_pb2.GroundStationConfigurationRequest(transmitter_configuration_request = config)

        return gs_config_request

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

def run_streamer(request_queue, thread_sts_queue):
    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(MY_CONFIG.API_KEY_PATH, MY_CONFIG.SSL_CA_CERT_PATH)

    request_generator = generate_request(request_queue, thread_sts_queue)

    try:
        for response in client.OpenSatelliteStream(request_generator):
            # thread_sts_queue.put("Received response: {}".format(response))
            pass
    except:
        thread_sts_queue.put("Shutting down streamer thread.")

def run():
    request_queue = Queue()
    thread_sts_queue = Queue()

    stream_config_request = stellarstation_pb2.SatelliteStreamRequest(
        satellite_id = str(MY_CONFIG.SATELLITE_ID),
        enable_events = True,
        enable_flow_control = True)
    request_queue.put(stream_config_request)

    streamer_thread = threading.Thread(target=run_streamer, args=(request_queue, thread_sts_queue,), daemon=True)
    streamer_thread.start()

    main_menu = ConsoleMenu("Main Menu (Radio Configuration)", "This example code exhibits a CLI that allows the user to build and send transceiver configuration commands.")

    # Transmitter
    outgoing_transmitter_req = GsConfigRequest("transmitter")
    config_transmitter_menu = ConsoleMenu("Submenu (Configure the Transmitter)", outgoing_transmitter_req.get_str)
    transmitter_config_menu_func_items = [
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
        FunctionItem("Finalize and Send Configuration Request", request_queue.put, [outgoing_transmitter_req.to_request]),
        FunctionItem("Reset Configuration Request Fields", outgoing_transmitter_req.clear_fields),
    ]
    for func_item in transmitter_config_menu_func_items:
        config_transmitter_menu.append_item(func_item)   
    config_transmitter_submenu = SubmenuItem("Configure Transmitter", config_transmitter_menu, main_menu)
    
    # Receiver
    outgoing_receiver_req = GsConfigRequest("receiver")
    config_receiver_menu = ConsoleMenu("Submenu (Configure the Receiver)", outgoing_receiver_req.get_str)
    receiver_config_menu_func_items = [
        FunctionItem("Set Bitrate", outgoing_receiver_req.prompt_set_bitrate),
        FunctionItem("Enable Carrier", outgoing_receiver_req.enable_carrier),
        FunctionItem("Disable Carrier", outgoing_receiver_req.disable_carrier),
        FunctionItem("Enable Sweep", outgoing_receiver_req.enable_sweep),
        FunctionItem("Disable Sweep", outgoing_receiver_req.disable_sweep),
        FunctionItem("Enable Modulation", outgoing_receiver_req.enable_modulation),
        FunctionItem("Disable Modulation", outgoing_receiver_req.disable_modulation),
        FunctionItem("Choose Modulation Type", outgoing_receiver_req.prompt_choose_modulation),
        FunctionItem("Enable Idle Pattern", outgoing_receiver_req.enable_idle_pattern),
        FunctionItem("Disable Idle Pattern", outgoing_receiver_req.disable_idle_pattern),
        FunctionItem("Finalize and Send Configuration Request", request_queue.put, [outgoing_receiver_req.to_request]),
        FunctionItem("Reset Configuration Request Fields", outgoing_receiver_req.clear_fields),
    ]
    for func_item in receiver_config_menu_func_items:
        config_receiver_menu.append_item(func_item)
    config_receiver_submenu = SubmenuItem("Configure Receiver", config_receiver_menu, main_menu)

    main_menu.append_item(config_transmitter_submenu)
    main_menu.append_item(config_receiver_submenu)
    main_menu.show()

    request_queue.put(None)
    streamer_thread.join()

    # while not thread_sts_queue.empty():
    #     sts = thread_sts_queue.get()
    #     print(sts)

if __name__ == '__main__':
    run()