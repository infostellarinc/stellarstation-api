# StellarStation example client - PrintingClient

A simple example client that prints received telemetry and sends commands at a fixed interval. Shows
the basics of writing code to integrate with the StellarStation API. Only works with [Fakeserver](../../fakeserver).

## Create and activate a venv for printing-client

```bash
$ python3 -m venv venv
$ source venv/bin/activate
```

## Install requirements
```bash
$ pip install -r requirements.txt
```

## Try it out!
Now, you'll see response from fakeserver by running printing-client.
```bash
$  python3 printing-client.py
```
