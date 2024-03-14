# StreamCLI
An example CLI for connecting and streaming to StellarStation from Rust and similar to [stellarcli](https://github.com/infostellarinc/stellarcli).

## Usage
To see the usage instructions
```
cargo run -- --help
```
```
Usage: streamcli [OPTIONS] --key <FILE> --satellite-id <SATELLITE_ID>

Options:
      --url <URL>
          URL to connect to for streaming [env: STELLARSTATION_API_URL=] [default: https://api.stellarstation.com]
      --key <FILE>
          Path to a StellarStation API key [env: STELLARSTATION_API_KEY=]
  -s, --satellite-id <SATELLITE_ID>
          Specify a satellite ID with which to Filter telemetry and commands
  -p, --plan-id <PLAN_ID>
          Specify a plan ID with which to filter telemetry and commands
  -r, --reconnect
          Enable trying to automatically reconnect if the stream is dropped
      --reconnect-stream-id <STREAM_ID>
          On the initial connection, use an existing stream ID to reconnect to that stream
      --reconnect-message-index <MESSAGE_INDEX>
          On the initial connection, set the next expected message index to receive
      --count <COUNT>
          Create multiple streams [default: 1]
  -h, --help
          Print help
```

## Authentication
Requires a StellarStation API key to authenticate. You can specify your key with either the `--key` argument or the `STELLARSTATION_API_KEY` environment variable which should be set to the path to that key.

## Custom Endpoints
By default, `streamcli` will connect to the production URL (`https://api.stellarstation.com`). You may override this by specifying either either the `--url` argument or the `STELLARSTATION_API_URL` environment variable.

## Examples
All examples assume that `STELLARSTATION_API_KEY` and `STELLARSTATION_API_URL` are already configured.

### Manual Reconnect
```
cargo run -- --satellite-id=1 --plan-id=2 --reconnect-stream-id=stream-123-abc --reconnect-message-index=100
```

### Auto Reconnect on Disconnect
```
cargo run -- --satellite-id=1 --plan-id=2 --reconnect
```

### Open multiple streams
```
cargo run -- --satellite-id=1 --plan-id=2 --count=2
```
