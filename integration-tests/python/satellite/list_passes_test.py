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
"""The integration tests for ListUpcomingAvailablePasses API in StellarStationService."""

from stellarstation.api.v1 import stellarstation_pb2

SATELLITE_ID = '98'


class TestListUpcomingAvailablePasses(object):
    def test_list_passes(self, stub_factory):
        client = stub_factory.get_satellite_service_stub()

        request = stellarstation_pb2.ListUpcomingAvailablePassesRequest()
        request.satellite_id = SATELLITE_ID

        result = client.ListUpcomingAvailablePasses(request)
        assert result

        # We need to use getattr to get the array of passes from the result instead of "result.pass"
        # since "pass" is reserved word in Python.
        passes = getattr(result, 'pass')
        assert len(passes) > 0
