# Copyright 2018 Infostellar, Inc.
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
"""The integration tests for ListPlans API in GroundStationService."""

from datetime import datetime

from stellarstation.api.v1.groundstation import groundstation_pb2

GS_ID = '27'


class TestListPlans(object):
    def test_list_plans(self, stub_factory):
        client = stub_factory.get_gs_service_stub()

        request = groundstation_pb2.ListPlansRequest()
        request.ground_station_id = GS_ID
        request.aos_after.FromDatetime(datetime(2018, 12, 1, 0, 0))
        request.aos_before.FromDatetime(datetime(2018, 12, 31, 0, 0))

        result = client.ListPlans(request)
        assert result
        assert len(result.plan) > 0
