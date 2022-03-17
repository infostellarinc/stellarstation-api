# Streaming API Benchmarking Tool

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

- After that of if you already have homebrew installed:
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

```
python3 stream_benchmark.py
```

### To run the app with [fake server]("../../../../fakeserver/"):
- you need to copy this [API key](../../fakeserver/src/main/jib/var/keys/api-key.json) to the current directory.
```sh
cp ../../fakeserver/src/main/jib/var/keys/api-key.json ./
```
- and then run the following command:

```
python3 stream_benchmark.py --endpoint=localhost:8080
```

### You can add the following flags to the run command:

- Set API key with --key arg (default 'api-key.json')
- Set ENDPOINT with --endpoint arg (default 'api.stellarstation.com:443')
- Set SATTELITE ID with --id arg (default '5')
- Set Reporting INTERVAL with --interval arg (default '10' seconds)
- Set Output DIRECTORY with --directory arg (default None)

## Notes:

- Output will write to a file if the Output DIRECTORY is set.
- CTRL-C will stop the test and write output to a file.
- Currently stream_benchmark.py is not compatible with python 3.8 and greater.
you need to use python 3.7 or older versions of python 3 to run it.