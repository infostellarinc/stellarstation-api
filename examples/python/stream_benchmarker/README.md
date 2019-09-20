# Streaming API Benchmarking Tool

Steps to run(commands Debian):

* Install python 3 dependencies and start virtual environment:

```
sudo apt update && sudo apt install python3 python3-venv
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

* Set API key with --key arg (default 'stellarstation-private-key.json')
* Set ENDPOINT with --endpoint arg (default 'api.stellarstation.com:443')
* Set SATTELITE ID with --id arg (default '5')
* Set Reporting INTERVAL with --interval arg (default '10' seconds)
* Set Output DIRECTORY with --directory arg (default None)

* Output will write to a file if the Output DIRECTORY is set

```
python3 stream_benchmark.py
```

* CTRL-C will stop the test and write output to a file

# Copyright 2019 Infostellar, Inc. All Rights Reserved.

