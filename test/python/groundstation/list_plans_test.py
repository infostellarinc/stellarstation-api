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
import unittest

from google.protobuf.timestamp_pb2 import Timestamp
from stellarstation.api.v1.groundstation import groundstation_pb2

from conn.factory import StubFactory

GS_ID = '27'


class TestStringMethods(unittest.TestCase):
    def setUp(self):
        self.factory = StubFactory()

    def test_list_plans(self):
        client = self.factory.get_gs_service_stub()

        fromTime = Timestamp(seconds=int(datetime(2018, 12, 1, 0, 0).timestamp()))
        toTime = Timestamp(seconds=int(datetime(2018, 12, 31, 0, 0).timestamp()))

        request = groundstation_pb2.ListPlansRequest(
            ground_station_id=GS_ID,
            aos_after=fromTime,
            aos_before=toTime
        )
        result = client.ListPlans(request)
        self.assertIsNotNone(result)


if __name__ == '__main__':
    unittest.main()
