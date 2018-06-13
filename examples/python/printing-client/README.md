# StellarStation example client - PrintingClient

A simple example client that prints received telemetry and sends commands at a fixed interval. Shows
the basics of writing code to integrate with the StellarStation API. Only works with [Fakeserver](../../fakeserver).

## Before you begin
### Prerequisites
When you start with gRPC in Python, you need pip, gRPC and gRPC tools. Ensure you have correctly installed those modules by following instructions in official [gRPC site](https://grpc.io/docs/quickstart/python.html#install-grpc-tools).


### Install Google authentication library for Python
In roder for authentication, install google-auth by running:
```bash
$ pip install --upgrade google-auth
```

### Compile proto file
After installation of required modules, you need to generate stub files from .proto file. Python's gRPC tools include the protol buffer compiler. To generate stubs for fakeserver, run:
```bash
$  python -m grpc_tools.protoc -I../../../api/src/main/proto/stellarstation/api/v1/ --python_out=. --grpc_python_out=. stellarstation.proto
```

### Try it out!
Now, you'll see response from fakeserver by running printing-client.
```bash
$  python printing-client.py
```
