# Integration tests by Python

This is a set of Python code that tests StellarStation API aiming to show how you can write API clients 
for StellarStation API.

Since these code are intended to server as sample code of the API, tests will not pass in your environment.
If you want to run and pass tests, 
  - You need to change satellite ID and ground station ID in the code.
  - You may need to disable tests for satellites API or ground station API when you don't have 
  any of theses.
  - Commands and telemetries used in open-stream tests are specific to our test environment.


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
