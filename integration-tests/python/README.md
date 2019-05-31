# Integration tests by Python

This is a set of Python code that tests StellarStation API aiming to show how you can write API clients 
for StellarStation API.

Since this code is intended to show real world usage of the API, you will need to make some changes
if you want the tests to pass in your environment:
  - You need to change satellite ID and ground station ID in the code to match the satellites and ground
stations that you own.
  - If you don't own any satellites or ground stations, you may need to disable the corresponding tests.
  - open-stream tests depend on command and telemetry messages, as well as server infrastructure, specific to
our test environment.


# How to run tests

The test code requires Python3 and pytest. You need to install dependencies with `pip` command. 
We would recommend to build the environment with venv or virtualenv not to break your current environment.  


Install dependencies
```bash
$ cd integration-tests/python 
$ pip install -r requirements.txt
```


You need to obtain an API key for StellarStation and set it as an environmental variable, 
`STELLARSTATION_API_KEY`. 
  
```bash
$ export STELLARSTATION_API_KEY=stellarstation-api-key.json
$ pytest .
```
