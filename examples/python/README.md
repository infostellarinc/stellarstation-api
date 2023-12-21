

# Getting Started with the StellarStation API Python Stubs

## About the StellarStation API
The StellarStation API facilitates communication with your satellite. As the capabilities of the StellarStation platform grows, so too does the API.

We use gRPC. Here are some reasons why:
1. It is one of the fastest ways to stream large amounts of data.
2. API stubs can be generated for [many languages](https://grpc.io/docs/languages/).
3. We can make our service's protobuf definitions directly accessible for you on [Github](https://github.com/infostellarinc/stellarstation-api/blob/master/api/src/main/proto/stellarstation/api/v1/stellarstation.proto).

## API Keys
StellarStation takes advantage of API keys to authenticate clients.

If you don't already have your API key, get it from your StellarStation organization administrator

or, 

sign into [StellarStation](https://www.stellarstation.com/console), go to settings, create your API key, and save it somewhere safe on your system.

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
$ source .venv/bin/activate
```
This step may differ based on your platform and shell. See how venvs work, [here](https://docs.python.org/3/library/venv.html#how-venvs-work).

4. Install the StellarStation API Python package by executing:
```bash
$ python3 -m pip install stellarstation
```

### Running the Examples
Again, using the terminal...

1. It's good practice to not share around sensitive information like keys and IDs. Set the following environment variables in your system:

STELLARSTATION_API_KEY_PATH         -> *a string, for example: ~/keys/your_personal_key.json*

STELLARSTATION_API_SATELLITE_ID     -> *an integer, for example: 123*

STELLARSTATION_API_CHANNEL_ID       -> *an integer, for example: 123*

STELLARSTATION_API_URL       -> *a string, for example: stream.qa.stellarstation.com*

For example, if you're using Debian, you would either execute the following in your terminal or add the following to your bash script (for permanency):
```bash
$ export STELLARSTATION_API_KEY_PATH=~/keys/your_personal_key.json
$ export STELLARSTATION_API_SATELLITE_ID=123
$ export STELLARSTATION_API_CHANNEL_ID=123
$ export STELLARSTATION_API_URL=stream.qa.stellarstation.com
```

If you added them to your bash script don't forget to re-source.

For example, if you're using Debian, that might look something like this,
```bash
$ source ~/.bashrc
```

2. Get the [StellarStation API code](https://github.com/infostellarinc/stellarstation-api) onto your computer either by downloading it or cloning the repository.
**We highly recommend cloning the repository using Git. You can find the installation instructions [here](https://github.com/git-guides/install-git)**

If using HTTPS,
```bash
$ git clone https://github.com/infostellarinc/stellarstation-api.git
```

If using SSL,
```bash
$ git clone git@github.com:infostellarinc/stellarstation-api.git
```

3. Navigate to the /stellarstation-api/examples/python directory.

4. Install the required tools.
```bash
$ python3 -m pip install -r requirements.txt
```

5. Run one of the examples.
```bash
$ python3 for_satellite_operators/list_reserved_plans.py
```


## For Windows Users
### Installation
If you haven't already done so, please install [Python](https://www.python.org/downloads/).
We recommend Python 3.10 or later, upgrading pip, and using a Python virtual environment.

1. Navigate to the base directory where you wish to run Python code that will interface with StellarStation and open Powershell.

2. In Powershell, execute the following command to create a new Python virtual environment:
```powershell
PS C:\> python -m venv .venv
```

3. Activate the virtual environment by executing:
```powershell
PS C:\> .venv\Scripts\Activate.ps1
```
If this does not work, you may need to change the Execution Policy settings for your user. See [RemoteSigned](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies?view=powershell-5.1#remotesigned) and [Unrestricted](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies?view=powershell-5.1#unrestricted) settings.

Example to enable running scripts in an unrestricted manner for the current user only:
```powershell
PS C:\> Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser
```

This step may differ based on your platform and shell. See how venvs work, [here](https://docs.python.org/3/library/venv.html#how-venvs-work).

4. Install the StellarStation API Python package by executing:
```powershell
PS C:\> python -m pip install stellarstation
```

### Running the Examples
Again, using Powershell...

1. It's good practice to not share around sensitive information like keys and IDs. Set the following environment variables in your system:

STELLARSTATION_API_KEY_PATH         -> *a string, for example: ~/keys/your_personal_key.json*

STELLARSTATION_API_SATELLITE_ID     -> *an integer, for example: 123*

STELLARSTATION_API_CHANNEL_ID       -> *an integer, for example: 123*

STELLARSTATION_API_URL       -> *a string, for example: stream.qa.stellarstation.com*

Here's an example:
```powershell
PS C:\> $Env:STELLARSTATION_API_KEY_PATH="C:\Users\Your Username\keys\your_personal_key.json"
PS C:\> $Env:STELLARSTATION_API_SATELLITE_ID="123"
PS C:\> $Env:STELLARSTATION_API_CHANNEL_ID="123"
PS C:\> $Env:STELLARSTATION_API_URL="stream.qa.stellarstation.com"
```

,or,

Or you can add environment variables via the Control Panel. If you do this, restart the shell process you are using.

2. Get the [StellarStation API code](https://github.com/infostellarinc/stellarstation-api) onto your computer either by downloading it or cloning the repository.
**We highly recommend cloning the repository using Git. You can find the installation instructions [here](https://github.com/git-guides/install-git)**

If using HTTPS,
```powershell
PS C:\> git clone https://github.com/infostellarinc/stellarstation-api.git
```

If using SSL,
```powershell
PS C:\> git clone git@github.com:infostellarinc/stellarstation-api.git
```

3. Navigate to the /stellarstation-api/examples/python directory.

4. Install the required tools.
```powershell
PS C:\> python -m pip install -r requirements.txt
```

5. Run one of the examples.
```powershell
PS C:\> python for_satellite_operators/list_reserved_plans.py
```
