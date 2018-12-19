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
"""The integration-tests for gRPC stubs for StellarStation API."""


def test_satellite_service_stub(stub_factory):
    client = stub_factory.get_satellite_service_stub()
    assert client


def test_groundstation_service_stub(stub_factory):
    client = stub_factory.get_gs_service_stub()
    assert client
