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
"""The integration tests for ListPlans API in StellarStationService."""

from datetime import datetime

from stellarstation.api.v1 import stellarstation_pb2

SATELLITE_ID = '98'


class TestListPlans(object):
    def test_list_plans(self, stub_factory):
        client = stub_factory.get_satellite_service_stub()

        request = stellarstation_pb2.ListPlansRequest()
        request.satellite_id = SATELLITE_ID
        request.aos_after.FromDatetime(datetime(2018, 12, 1, 0, 0))
        request.aos_before.FromDatetime(datetime(2018, 12, 15, 0, 0))

        result = client.ListPlans(request)
        assert result
        assert len(result.plan)
