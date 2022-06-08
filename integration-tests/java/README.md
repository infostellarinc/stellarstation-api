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
Open `src/main/resources/application.conf` and replace `PATH_TO_YOUR_API_KEY` with the path to your API key.

For example, if you saved the key as `stellarstation-private-key.json ` in `/home/kevin/stellarstation`, the value
should be `/home/kevin/stellarstation/stellarstation-private-key.json`.     
 

### Run tests
You can run tests from the top level directory with the following command.

```bash
$ ./gradlew integration-tests:java:integrationTest
```

# How to initiate a new project from this example
This section explains how you can write your Java clients based on this example.

### Set up a new directory
Create a new directory, and copy the entire contents of the java integration directory into it.

```bash
$ mkdir my-client
$ cd my-client
$ copy -R PATH_TO_STELLARSTATION_API/integration-tests/java/* ./
```

### Replace dependencies in build.gradle
In order to run these tests in your own copy of the source code, you need to update the dependency
on stellarstation-api from an internal reference to the external one.

To do that, open `build.gradle` and replace `implementation project(':api')` in dependencies section with
`implementation 'com.stellarstation.api:stellarstation-api:0.12.0'`.


### Set your API key
You need to obtain an API key for StellarStation and set it in a configuration file.
Open `src/main/resources/application.conf` and replace `PATH_TO_YOUR_API_KEY` with the path to your API key.

For example, if you saved the key as `stellarstation-private-key.json ` in `/home/kevin/stellarstation`, the value
should be `/home/kevin/stellarstation/stellarstation-private-key.json`.


### Build the source code
You can build the source code from the top level directory with the following command.

```bash
$ ./gradlew :integrationTestClasses
```

### Develop your client
Congratulations! You are ready to build your own API client.
