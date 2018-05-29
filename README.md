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

// TODO(rag): Provide details on how to register call credentials for StellarStation private keys.

## Usage

When using `proto` files from this repository directly in client code, make sure to only use [tagged releases](https://github.com/infostellarinc/stellarstation-api/releases).
Using `proto` files from any non-tagged revision will likely not work correctly.
