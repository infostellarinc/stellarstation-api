# Streaming API Benchmarking Tool

Requires Go version 1.17

## How to set up

```sh
go mod download
go mod tidy
```

## To run the tool with [fake server](../../fakeserver):
- you need to copy this [API key](../../fakeserver/src/main/jib/var/keys/api-key.json) to the current directory.
```bash
$ cp ../../fakeserver/src/main/jib/var/keys/api-key.json ./
```
- uncomment ```InsecureSkipVerify: true,``` in benchmark/conn.go file inside ```tlsConfig := &tls.Config{}```
- and then run the following command:

```bash
$ go run . -E=localhost:8080
```

## To run the tool with (FAKE-DEMO2-SAT2):
You need to create your own API key for the test satellite’s organization:

- Go to this [link](https://internal.stellarstation.com/console)
- Impersonate the org you want the key for (FAKE-DEMO2-SAT2) ID 101 is a common one to use. It’s under the organization StellarStation Demo2
- Go to Settings -> API -> Generate a new API key

And place it in the same directory of this project and name it "api-key.json".
if you want to name it any other name you will need to pass it as argument with -k flag.

And then run the following command:

```bash
$ go run . -s=101
```

## To build from source:
```bash
$ go build
```

Usage:
```
Usage of ./benchmark:
  -E string
    	API endpoint (default "api.stellarstation.com:443")
  -P	Do not print a pass summary after each pass
  -S	Do not print an overall summary when the program exits
  -e duration
    	Assume a pass has ended after this much time has passed without receiving any additional data (default 10s)
  -i duration
    	Reporting interval.  (10s, 1m, etc.)  During a pass, an output line will be generated for each reporting interval. (default 10s)
  -k string
    	StellarStation API Key file (default "api-key.json")
  -o string
    	Write report output to a file instead of standard out
  -s string
    	Satellite ID as provided by StellarStation (default "5")
  -x	Exit the program after a pass ends
```

* CTRL-C will stop the session at any time
