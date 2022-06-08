# StellarStation example client - PrintingClient

A simple example client that prints received telemetry and sends commands at a fixed interval. Shows
the basics of writing code to integrate with the StellarStation API. Only works with [Fakeserver](../../fakeserver).

## Setting up the environment:

## Install python 3 dependencies and start virtual environment:

- For linux users:

```bash
$ sudo apt update && sudo apt install python3 python3-venv
```

- For mac users: you need to install homebrew
```bash
$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

- After that or if you already have homebrew installed:
```bash
$ brew install python3
$ pip3 install virtualenv
```

- Install dependencies and start virtual environment:

```bash
$ python3 -m venv venv
$ source venv/bin/activate
$ pip3 install -r requirements.txt
```

## Try it out!
Now, you'll see response from fakeserver by running printing-client.
```bash
$  python3 printing-client.py
```
