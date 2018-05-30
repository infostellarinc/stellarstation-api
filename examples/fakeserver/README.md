# StellarStation API Fakeserver

A fake implementation of the StellarStation API to use to check behavior of client code before
integrating with the service.

## Features

- Verifies a fake API key
- Returns 1MB of random telemetry every second
- Echos back received commands
- Cancels the connection after 5 minutes
- Only allows requests for satellite ID `"5"`, other IDs show behavior for non-existent satellites

## Running

// TODO(rag): Add instructions for running with docker after pushing an image.
