# StellarStation API

The public API definition for [StellarStation](https://www.stellarstation.com/) and supported client
libraries / helpers.

Feel free to send PRs to improve documentation when things are unclear or file issues with questions on usage.

## Usage

The StellarStation API is based on [gRPC](https://grpc.io). An API client can be written in any
language supported by gRPC by following one of the language-specific guides [here](https://grpc.io/docs/).

The main protocol definition used to generate language specific stub code is [here](./api/src/main/proto/stellarstation/api/v1/stellarstation.proto).

Language-specific documentation:

- [Java](https://javadoc.io/doc/com.stellarstation.api/stellarstation-api/)
- [Go](https://godoc.org/github.com/infostellarinc/go-stellarstation/api/v1)

When using `proto` files from this repository directly in client code, make sure to only use [tagged releases](https://github.com/infostellarinc/stellarstation-api/releases).
Using `proto` files from any non-tagged revision will likely not work correctly or maintain backwards compatibility.

The API follows semantic versioning - any breaking, backwards incompatible change will be made while increasing the
major version.

### Java

We provide precompiled client stubs for Java. Java users can just add a dependency on
the stubs and don't need to compile the protocol into code themselves.

Gradle users should add the `stellarstation-api` artifact to their `dependencies`, e.g.,

```groovy
dependencies {
    compile 'com.stellarstation.api:stellarstation-api:0.2.0'
}
```

Maven users would add to their `pom.xml`

```xml
<dependencies>
  <dependency>
    <groupId>com.stellarstation.api</groupId>
    <artifactId>stellarstation-api</artifactId>
    <version>0.2.0</version>
  </dependency>
</dependencies>
```

A full example of a Java API client can be found [here](./examples/java/printing-client).

We publish `SNAPSHOT` builds to https://oss.jfrog.org/libs-snapshot/ for access to preview features.
The same caveats as using non-tagged releases applies - not all functions in `SNAPSHOT` builds may
be implemented yet and there is no guarantee of backwards compatibility for `SNAPSHOT` builds. It is
generally not recommended to use `SNAPSHOT` builds without first consulting with your StellarStation rep.

#### Note for Alpine Linux users

For anyone trying to use the Java API client in an Alpine Linux container, they will find it doesn't
work due to a limitation of gRPC with Java 8. There are many ways to work around this, such as
using [jetty-alpn](https://www.eclipse.org/jetty/documentation/current/alpn-chapter.html) or
installing a version of Java 9+, but our recommendation for Java 8 users is to use 
[distroless](https://github.com/GoogleContainerTools/distroless/blob/master/java/README.md), which
is similarly compact but will work fine with gRPC.

### Python

We provide precompiled client stubs for Python. Python users can install them with `pip`.

```bash
$  pip install --upgrade stellarstation
```

A full example of a Python API client can be found [here](./examples/python/printing-client).

### Go

We provide precompiled client stubs for Go, found [here](https://github.com/infostellarinc/go-stellarstation).

```go
import stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
```

### NodeJS

We provide precompiled client stubs for NodeJS. NodeJS users can install them with `npm`.

```bash
$ npm install @infostellarinc/stellarstation-api
```

## Authentication

Authentication to the StellarStation API is done using JWT bearer tokens (https://jwt.io). When
initializing an API client, make sure to register call credentials using the private key downloaded
from the StellarStation Console. Details for registering call credentials on a gRPC stub can be
found [here](https://grpc.io/docs/guides/auth.html). Note that if the key has been revoked on the
console, it will not be usable to authenticate with the API.

### Java
For Java, the `grpc-auth` and `google-auth-library-oauth2-http` libraries can be used to easily setup
authentication of an API client.

```java
// Load the private key downloaded from the StellarStation Console.
ServiceAccountJwtAccessCredentials credentials =
    ServiceAccountJwtAccessCredentials.fromStream(
        Resources.getResource("stellarstation-private-key.json").openStream(),
        URI.create("https://api.stellarstation.com"));

// Setup the gRPC client.
ManagedChannel channel =
    ManagedChannelBuilder.forAddress("api.stellarstation.com", 443)
        .build();
StellarStationServiceStub client =
    StellarStationServiceGrpc.newStub(channel)
        .withCallCredentials(MoreCallCredentials.from(credentials));
```
### Python
`google-auth` for Python can be used for authentication of an API client.


```python
# Load the private key downloaded from the StellarStation Console.
credentials = google_auth_jwt.Credentials.from_service_account_file(
  'stellarstation-private-key.json',
  audience='https://api.stellarstation.com')

# Setup the gRPC client.
jwt_creds = google_auth_jwt.OnDemandCredentials.from_signing_credentials(
  credentials)
channel = google_auth_transport_grpc.secure_authorized_channel(
  jwt_creds, None, 'api.stellarstation.com:443')
client = stellarstation_pb2_grpc.StellarStationServiceStub(channel)
```


### Other languages
Other languages have similar methods for loading Service Account JWT Access Credentials.
For example,

- C++ - https://github.com/grpc/grpc/blob/583f39ad94c0a14a50916e86a5ccd8c3c77ae2c6/include/grpcpp/security/credentials.h#L144
- Go - https://github.com/grpc/grpc-go/blob/96cefb43cfc8b2cd3fed9f19f59830bc69e30093/credentials/oauth/oauth.go#L60
