# Copyright 2022 Infostellar, Inc.
# Requests and prints the past 30 days-worth of completed plans' telemetry file URLs for your satellite.

import os

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1 import stellarstation_pb2

import toolkit

def get_plans(client, sat_id, days=-30):
    start = Timestamp()
    start.GetCurrentTime()
    end = Timestamp()
    end.GetCurrentTime()

    start.FromSeconds(int(start.ToSeconds()) + (days * 24 * 3600))

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

    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(STELLARSTATION_API_KEY_PATH, "")

    # Get the plans
    plans = get_plans(client, STELLARSTATION_API_SATELLITE_ID)

    # Get plans that are COMLETED
    completed_plans = [plan for plan in plans if toolkit.PlanStatus(plan.status).name == "SUCCEEDED"]

    # Print the plans
    for i, plan in enumerate(completed_plans):
        # Each plan in this list of plans is simply a protobuf 'Plan' message (defined in stellarstation.proto)
        if plan.telemetry_metadata:
            print("({} of {}):{}\n".format(i + 1, len(completed_plans), plan.telemetry_metadata))
        else:
            print("({} of {}):{}\n".format(i + 1, len(completed_plans), "DNE"))

if __name__ == '__main__':
    run()
