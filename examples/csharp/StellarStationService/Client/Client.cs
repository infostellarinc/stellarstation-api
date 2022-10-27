using Grpc.Net.Client;

namespace Stellarstation {
    public class Client {
        Stellarstation.Api.V1.StellarStationService.StellarStationServiceClient cl;

        public Client(GrpcChannel chan) {
            cl = new Stellarstation.Api.V1.StellarStationService.StellarStationServiceClient(chan);
        }

        public Stellarstation.Api.V1.ListUpcomingAvailablePassesResponse ListUpcomingAvailablePasses(int satId, DateTime startTime, DateTime stopTime) {
            var start = new Google.Protobuf.WellKnownTypes.Timestamp{ 
                Seconds = (long)(startTime.ToUniversalTime() - DateTime.MinValue).TotalSeconds 
            };
            var stop = new Google.Protobuf.WellKnownTypes.Timestamp{ 
                Seconds = (long)(stopTime.ToUniversalTime() - DateTime.MinValue).TotalSeconds 
            };

            var req = new Stellarstation.Api.V1.ListPlansRequest {
                SatelliteId = satId.ToString(),
                AosAfter = start,
                AosBefore = stop,
            };

            var req2 = new Stellarstation.Api.V1.ListUpcomingAvailablePassesRequest {
                SatelliteId = satId.ToString(),

            };

            return cl.ListUpcomingAvailablePasses(req2);
        }
    }
}