# Copyright 2022 Infostellar, Inc.
# Requests and prints the past 30 days-worth of completed plans' telemetry file URLs for your satellite.

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1 import stellarstation_pb2

import toolkit
import MY_CONFIG

def get_plans(client, days=-30):
    start = Timestamp()
    start.GetCurrentTime()
    end = Timestamp()
    end.GetCurrentTime()

    start.FromSeconds(int(start.ToSeconds()) + (days * 24 * 3600))

    request = stellarstation_pb2.ListPlansRequest(
            satellite_id = str(MY_CONFIG.SATELLITE_ID),
            aos_after = start,
            aos_before = end)
    
    listPlansResponse = client.ListPlans(request)

    plans = listPlansResponse.plan

    return plans

def run():
    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(MY_CONFIG.API_KEY_PATH, MY_CONFIG.SSL_CA_CERT_PATH)

    # Get the plans
    plans = get_plans(client)

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
