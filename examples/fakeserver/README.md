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

Users of this repository can just use the `run` task.

```bash
$ ./gradlew :examples:fakeserver:run
```

Make sure to initialize a client with the API key located [here](./src/misc/api-key.json).

You can run the fakeserver even without using this repository via Docker.

First extract the api-key from the Docker image.

```bash
$ docker pull quay.io/infostellarinc/fake-apiserver
$ docker run -v `pwd`:/out --entrypoint extract-key -it --rm infostellarinc/fake-apiserver
```

A file named `api-key.json` will be present in the current directory which you can use to initialize
a client.

Then just run the server with docker.

```bash
$ docker run -p 8081:8081 -it --rm quay.io/infostellarinc/fake-apiserver
```

The server will be listening for plaintext on port 8081. See [printing-client](../printing-client)
for simple example code that exercises the server code.