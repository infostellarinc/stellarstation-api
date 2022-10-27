using Grpc.Net.Client;
using Grpc.Auth;
using Google.Apis.Auth.OAuth2;

namespace Stellarstation {
    public static class Auth {
        public static GrpcChannel GenerateChannel(string apiAddress, string apiKeyPath) {
            using (var stream = new FileStream(apiKeyPath, FileMode.Open, FileAccess.Read)) {
                var creds = ServiceAccountCredential.FromServiceAccountData(stream).ToChannelCredentials();
                return GrpcChannel.ForAddress(apiAddress, new GrpcChannelOptions {
                    Credentials = creds,
                    MaxSendMessageSize = 512 * 1024 * 1024,
                    MaxReceiveMessageSize = 512 * 1024 * 1024
                });
            }
        }
    }
}