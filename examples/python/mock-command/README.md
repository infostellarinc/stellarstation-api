# Streaming Mock Command Tool
This example:
- Opens stream for the provided satellite ID.
- Sends a fixed mock command (CMD) to the radio transmitter every time a TLM message is received.

## Setting up the environment:

### Install python 3 dependencies and start virtual environment:

- For linux users:

```sh
sudo apt update && sudo apt install python3 python3-venv
```

- For mac users: you need to install homebrew
```sh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

- After that or if you already have homebrew installed:
```sh
brew install python3
pip3 install virtualenv
```

- Install dependencies and start virtual environment:

```sh
python3 -m venv venv
source venv/bin/activate
pip3 install -r requirements.txt
```

## To run the app:
Use the following command:

```sh
python3 mock-command.py
```

### To run the app with [fake server]("../../../../fakeserver/"):
- you need to copy this [API key](../../fakeserver/src/main/jib/var/keys/api-key.json) to the current directory.
```sh
cp ../../fakeserver/src/main/jib/var/keys/api-key.json ./api-key.json
```
- you need to copy this [CA Cert](../../fakeserver/src/main/resources/tls.crt) inside the current directory.
```sh
cp ../../fakeserver/src/main/resources/tls.crt ./tls.crt
```
- and then run the following command:

```sh
python3 mock-command.py --endpoint=localhost:8080
```

### You can add the following flags to the run command:

- Set SATTELITE ID with --id arg (default '5')
- Set CHANNEL SET ID with --channelset arg (default 'test')
- Set API key path with --key arg (default './api-key.json')
- Set ENDPOINT with --endpoint arg (default 'api.stellarstation.com:443')
- Set SSL CA Certificate path with --sslcert arg (default './tls.crt')