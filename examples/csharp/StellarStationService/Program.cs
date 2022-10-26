using System.Threading.Tasks;
using Grpc.Net.Client;
using Grpc.Auth;
using Google.Apis.Auth.OAuth2;

// See https://aka.ms/new-console-template for more information
Console.WriteLine("Hello, World!");

var credential = GoogleCredential.GetApplicationDefault();


var newCreds = GoogleGrpcCredentials.ToChannelCredentials(credential);
using var channel = GrpcChannel.ForAddress("https://localhost:7042", new GrpcChannelOptions
    {
        Credentials = newCreds,
        MaxReceiveMessageSize = 5 * 1024 * 1024, // 5 MB
        MaxSendMessageSize = 2 * 1024 * 1024 // 2 MB
    });
var client = new Stellarstation.Api.V1.StellarStationService.StellarStationServiceClient(channel);

var startTime = DateTime.UtcNow - DateTime.MinValue;
var start = new Google.Protobuf.WellKnownTypes.Timestamp{ Seconds = (long)startTime.TotalSeconds, };
var stopTime = startTime.Add(TimeSpan.FromDays(3));
var stop = new Google.Protobuf.WellKnownTypes.Timestamp{ Seconds = (long)startTime.TotalSeconds, };

var req = new Stellarstation.Api.V1.ListPlansRequest {
    SatelliteId = "174",
    AosAfter = start,
    AosBefore = stop,
};
