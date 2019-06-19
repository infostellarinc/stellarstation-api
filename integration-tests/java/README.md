# Java integration tests

This is a set of Java code that tests StellarStation API aiming to show how you can write API clients
for StellarStation API.

Since this code is intended to show real world usage of the API, you will need to make some changes
if you want the tests to pass in your environment:
  - You need to change satellite ID and ground station ID in the test code to match the satellites and ground
stations that you own.
  - If you don't own any satellites or ground stations, you may need to disable the corresponding tests.
  - open-stream tests depend on command and telemetry messages, as well as server infrastructure, specific to
our test environment.


# How to run tests


### Set your API key
You need to obtain an API key for StellarStation and set it in a configuration file. 
Open `src/main/resources/application.conf` and replace `PATH_TO_YOUR_API_KEY` to your API key.

For example, if you saved the key as `stellarstation-private-key.json ` in `/home/kevin/stellarstation`, they value
should be `/home/kevin/stellarstation/stellarstation-private-key.json`.     
 

### Run tests
You can run tests from the top level directory with the following command.

```bash
$ ./gradlew integration-tests:java:integrationTest
``` 
