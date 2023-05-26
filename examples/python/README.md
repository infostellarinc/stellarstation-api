

# Getting Started with the StellarStation API Python Stubs

## About the StellarStation API
The StellarStation API facilitates communication with your satellite. As the capabilities of the StellarStation platform grows, so too does the API.

We use gRPC. Here are some reasons why:
1. It is one of the fastest ways to stream large amounts of data.
2. API stubs can be generated for [many languages](https://grpc.io/docs/languages/).
3. We can make our service's protobuf definitions directly accessible for you on [Github](https://github.com/infostellarinc/stellarstation-api/blob/master/api/src/main/proto/stellarstation/api/v1/stellarstation.proto).


## For Linux Users
### Installation
If you haven't already done so, please install [Python](https://www.python.org/downloads/).
We recommend Python 3.10 or later, upgrading pip, and using a Python virtual environment.
We also recommend updating and upgrading whatever Linux package tool you are using (apt, apt-get, etc).

1. Navigate to the base directory where you wish to run Python code that will interface with StellarStation and open a terminal.

2. In the terminal window, execute the following command to create a new Python virtual environment:
```bash
$ python3 -m venv .venv
```
3. Activate the virtual environment by executing:
```bash
$ source venv/bin/activate
```

4. Install the StellarStation API Python package by executing:
```bash
$ python3 -m pip install stellarstation
```

### Running the Examples
Again, using the terminal...

1. If you don't already have your API key, do the following:
Get it from your StellarStation organization administrator,
or,
Sign into [StellarStation](https://www.stellarstation.com/console), go to settings, create your API key, and save it somewhere safe on your system.

2. It's good practice to not share around sensitive information like keys and IDs. Set the following environment variables in your system:
STELLARSTATION_API_KEY_PATH         -> *a string, for example: ~/keys/your_personal_key.json*
STELLARSTATION_API_SATELLITE_ID     -> *an integer, for example: 123*
STELLARSTATION_API_CHANNEL_ID       -> *an integer, for example: 123*

For example, if you're using Debian, you would either execute the following in your terminal or add the following to your bash script (for permanency):
```bash
$ export STELLARSTATION_API_KEY_PATH=~/keys/your_personal_key.json
$ export STELLARSTATION_API_SATELLITE_ID=123
$ export STELLARSTATION_API_CHANNEL_ID=123
```

If you added them to your bash script, don't forget to resource with:
For example, if you're using Debian, that might look something like this,
```bash
$ source ~/.bashrc
```

3. Get the [StellarStation API code](https://github.com/infostellarinc/stellarstation-api) onto your computer either by downloading it or cloning the repository.
**We highly recommend cloning the repository using Git. You can find the installation instructions [here](https://github.com/git-guides/install-git)**
If using HTTPS,
```bash
$ git clone https://github.com/infostellarinc/stellarstation-api.git
```

If using SSL,
```bash
$ git clone git@github.com:infostellarinc/stellarstation-api.git
```

4. Navigate to the /stellarstation-api/examples/python directory.

5. Install the required tools.
```bash
$ python3 -m pip install -r requirements.txt
```

6. Run one of the examples.
```bash
$ python3 for_satellite_operators/list_reserved_plans.py
```


## For Windows Users
### Installation

### Running the Examples