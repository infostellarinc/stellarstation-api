# StellarStation example client - PrintingClient

A simple example client that prints received telemetry and sends commands at a fixed interval. Shows
the basics of writing code to integrate with the StellarStation API. Only works with [Fakeserver](../../fakeserver).


## Install Google authentication library for Python
For authentication, install google-auth by running:
```bash
$ pip install --upgrade google-auth
```

## Install StellarStation API library
After installation of required modules, you need stubs generated from .proto file. To install precompiled client stubs for Python, run:

```bash
$  pip install --upgrade stellarstation
```

## Try it out!
Now, you'll see response from fakeserver by running printing-client.
```bash
$  python printing-client.py
```
