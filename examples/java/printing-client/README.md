# StellarStation example client - PrintingClient

A simple example client that prints received telemetry and sends commands at a fixed interval. Shows
the basics of writing code to integrate with the StellarStation API. Only works with [Fakeserver](../../fakeserver).

## Running the server with Gradle

Users of this repository can just use the `run` task.

```bash
$ ./gradlew :examples:java:printing-client:run
```

If using a lower version of Java than Java 10, you will need to specify the build file explicitly.

```bash
$ ./gradlew -b examples/java/printing-client/build.gradle run
```