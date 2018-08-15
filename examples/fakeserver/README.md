# StellarStation API Fakeserver

A fake implementation of the StellarStation API to use to check behavior of client code before
integrating with the service.

## Features

- Verifies a fake API key
- Returns 1MB of random telemetry every second
- Echos back received commands
- Cancels the connection after 5 minutes
- Only allows requests for satellite ID `"5"`, other IDs show behavior for non-existent satellites

## Running

### Running the server with Gradle

Users of this repository can just use the `run` task.

```bash
$ ./gradlew :examples:fakeserver:run
```

If using a lower version of Java than Java 10, you will need to specify the build file explicitly.

```bash
$ ./gradlew -b examples/fakeserver/build.gradle run
```

Make sure to initialize a client with the API key located [here](./src/misc/api-key.json).

### Running the server with Docker

You can run the fakeserver even without using this repository via Docker.

First extract the api-key from the Docker image.

```bash
$ docker pull quay.io/infostellarinc/fake-apiserver
$ docker run -v `pwd`:/out --entrypoint sh -it --rm quay.io/infostellarinc/fake-apiserver /extract-key
```

A file named `api-key.json` will be present in the current directory which you can use to initialize
a client.

Then just run the server with docker.

```bash
$ docker run -p 8080:8080 -p 8081:8081 -it --rm quay.io/infostellarinc/fake-apiserver
```

The server will be listening for plaintext on port 8081 and TLS on port 8080. Connecting on TLS will
require either disabling TLS verification or adding [tls.crt](./src/main/resources/tls.crt) as a
trusted certificate. See [printing-client](../java/printing-client) for simple example code that 
exercises the server code.

### Releasing docker image

Currently, releasing the docker image is a manual process until https://github.com/GoogleContainerTools/jib/issues/601
is resolved. In the meantime,

```bash
$ docker login quay.io  # Only need to do this once on a machine
$ ./gradlew :examples:fakeserver:jibDockerBuild
$ docker run -it --rm quay.io/infostellarinc/fake-apiserver  # Sanity check the server starts up
$ docker push quay.io/infostellarinc/fake-apiserver
``` 
