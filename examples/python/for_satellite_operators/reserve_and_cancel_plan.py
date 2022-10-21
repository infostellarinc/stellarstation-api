# Copyright 2022 Infostellar, Inc.
# Reserves the first available plan, confirms to the user that it has been reserved by getting upcoming plans and printing it,
#   then cancels the plan (to clean up after this example runs).

from time import sleep

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1 import stellarstation_pb2

import toolkit
import MY_CONFIG

def get_plans(client, days=1):
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
    WAIT_TIME_SEC = 5

    # A client is necessary to receive services from StellarStation.
    client = toolkit.get_grpc_client(MY_CONFIG.API_KEY_PATH, MY_CONFIG.SSL_CA_CERT_PATH)

    # Get passes that a plan can be scheduled for
    # Each pass in this list of plans is simply a protobuf 'Pass' message (defined in stellarstation.proto)
    request = stellarstation_pb2.ListUpcomingAvailablePassesRequest(satellite_id = str(MY_CONFIG.SATELLITE_ID))
    response = client.ListUpcomingAvailablePasses(request)
    available_passes = getattr(response, "pass")

    # Reserve a plan for the first pass using the first available channel set's reservation token
    plan_ids_before_resevation = [plan.id for plan in get_plans(client)]
    assert available_passes[0], "There are no available passes. Contact Infostellar to check configuration settings."
    assert available_passes[0].channel_set_token[0], "There are no channel sets configured for this satellite. Contact Infostellar to check configuration settings."
    first_pass = available_passes[0]
    print("The first pass is over Ground Station of ID {}, with AoS={} and LoS={} UTC".format(
            first_pass.ground_station_id,
            first_pass.aos_time.ToDatetime(),
            first_pass.los_time.ToDatetime()))
    print("Attempting to reserve a plan on that first pass...")
    reservation_token = first_pass.channel_set_token[0].reservation_token
    request = stellarstation_pb2.ReservePassRequest(reservation_token = reservation_token, priority = "HIGH")
    response = client.ReservePass(request)
    scheduled_plan = response.plan

    # Give the servers a bit of time to update
    print("Reservation request sent. Waiting {} seconds for servers to update...".format(WAIT_TIME_SEC))
    sleep(WAIT_TIME_SEC)

    # Assert that the plan was scheduled and print the ID
    plan_ids_after_resevation = [plan.id for plan in get_plans(client)]
    assert scheduled_plan.id not in plan_ids_before_resevation
    assert scheduled_plan.id in plan_ids_after_resevation
    print("Successfully scheduled plan ID ({})".format(scheduled_plan.id))
    # print("--Scheduled Plan Details-------------------------------------")
    # print(scheduled_plan)

    # Cancel the plan to clean up
    print("Canceling the plan to clean up...".format(scheduled_plan.id))
    request = stellarstation_pb2.CancelPlanRequest(plan_id = scheduled_plan.id)
    client.CancelPlan(request)

    # Give the servers a bit of time to update
    print("Cancelation request sent. Waiting {} seconds for servers to update...".format(WAIT_TIME_SEC))
    sleep(WAIT_TIME_SEC)

    # Checking cancelation successful
    my_canceled_plan = next((plan for plan in get_plans(client)), None)
    assert my_canceled_plan
    assert toolkit.PlanStatus(my_canceled_plan.status).name == "CANCELED"
    print("Successfully canceled plan of ID ({})".format(scheduled_plan.id))

    print("Example finished. Exiting...")

if __name__ == '__main__':
    run()
