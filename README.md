# StellarStation API

The public API definition for [StellarStation](https://www.stellarstation.com/) and supported client libraries / helpers.

This repository is currently under construction and is provided for reference. API implementation is
in progress and documentation will continue to evolve. Feel free to send PRs to improve
documentation when things are unclear or file issues with questions on usage.

## Authentication

Authentication to the StellarStation API is done using JWT bearer tokens (https://jwt.io). When
initializing an API client, make sure to register call credentials using the private key downloaded
from the StellarStation Console. Details for registering call credentials on a gRPC stub can be
found [here](https://grpc.io/docs/guides/auth.html). Note that if the key has been revoked on the
console, it will not be usable to authenticate with the API.

The `grpc-auth` and `google-auth-library-oauth2-http` libraries can be used to easily setup
authentication of an API client.

```java
// Load the private key downloaded from the StellarStation Console.
ServiceAccountJwtAccessCredentials credentials =
    ServiceAccountJwtAccessCredentials.fromStream(
        Resources.getResource("api-key.json").openStream(),
        URI.create("https://api.stellarstation.com"));

// Setup the gRPC client.
ManagedChannel channel =
    ManagedChannelBuilder.forAddress("localhost", 8081)
        .build();
StellarStationServiceStub client =
    StellarStationServiceGrpc.newStub(channel)
        .withCallCredentials(MoreCallCredentials.from(credentials));
```

A full example of an API client can be found [here](./examples/fakeserver).

## Usage

When using `proto` files from this repository directly in client code, make sure to only use [tagged releases](https://github.com/infostellarinc/stellarstation-api/releases).
Using `proto` files from any non-tagged revision will likely not work correctly.
