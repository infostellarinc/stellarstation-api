# Copyright 2022 Infostellar, Inc.
# Requests and prints the next 3 days-worth of reserved plans for your satellite.

import os

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1 import stellarstation_pb2

import toolkit

def get_plans(client, sat_id, days=3):
    start = Timestamp()
    start.GetCurrentTime()

    end = Timestamp()
    end.GetCurrentTime()
    end.FromSeconds(int(start.ToSeconds()) + (days * 24 * 3600))

    request = stellarstation_pb2.ListPlansRequest(
            satellite_id = sat_id,
            aos_after = start,
            aos_before = end)
    
    listPlansResponse = client.ListPlans(request)

    plans = listPlansResponse.plan

    return plans

def run():
    STELLARSTATION_API_KEY_PATH = os.getenv('STELLARSTATION_API_KEY_PATH')
    STELLARSTATION_API_SATELLITE_ID = os.getenv('STELLARSTATION_API_SATELLITE_ID')

    assert STELLARSTATION_API_KEY_PATH, "Did you properly define this environment variable on your system?"
    assert STELLARSTATION_API_SATELLITE_ID, "Did you properly define this environment variable on your system?"
    
    STELLARSTATION_API_URL = os.getenv('STELLARSTATION_API_URL','stream.qa.stellarstation.com')
    assert STELLARSTATION_API_URL, "Did you properly define this environment variable on your system?"

    if not STELLARSTATION_API_KEY_PATH:
        raise ValueError("Expected a string for environment variable STELLARSTATION_API_KEY_PATH but got {STELLARSTATION_API_KEY_PATH}.")
    if not STELLARSTATION_API_SATELLITE_ID:
        raise ValueError("Expected a string for environment variable STELLARSTATION_API_SATELLITE_ID but got {STELLARSTATION_API_SATELLITE_ID}.")

    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(STELLARSTATION_API_KEY_PATH, STELLARSTATION_API_URL)

    # Get the plans
    plans = get_plans(client, STELLARSTATION_API_SATELLITE_ID)

    # Get plans that are RESERVED
    reserved_plans = [plan for plan in plans if toolkit.PlanStatus(plan.status).name == "RESERVED"]

    if len(reserved_plans) == 0:
        print("No reserved plans found.")

    # Print the plans
    for i, plan in enumerate(reserved_plans):
        # Each plan in this list of plans is simply a protobuf 'Plan' message (defined in stellarstation.proto)
        print("--Plan ({} of {})---------------------------------------------------------".format(i + 1, len(reserved_plans)))

        print("Plan ID: {}\nStatus: {}\nAoS (UTC): {}\nLoS (UTC): {}\nGround Station Lat: {}\nGround Station Lon: {}\n".format(
                plan.id,
                toolkit.PlanStatus(plan.status).name,
                plan.aos_time.ToDatetime(),
                plan.los_time.ToDatetime(),
                plan.ground_station_latitude,
                plan.ground_station_longitude))

if __name__ == '__main__':
    run()
