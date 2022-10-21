# Copyright 2022 Infostellar, Inc.
# Requests and prints the next 3 days-worth of reserved plans for your satellite.

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1 import stellarstation_pb2

import toolkit
import MY_CONFIG

def get_plans(client, days=3):
    start = Timestamp()
    start.GetCurrentTime()

    end = Timestamp()
    end.GetCurrentTime()
    end.FromSeconds(int(start.ToSeconds()) + (days * 24 * 3600))

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
