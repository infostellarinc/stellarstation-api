# Streaming API Benchmarking Tool

Requires Go version 1.3

To build from source:
```
go build -o benchmark benchmark/*.go
```

To build binaries for all operating systems:
```
build.sh
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
    	StellarStation API Key file (default "stellarstation-private-key.json")
  -o string
    	Write report output to a file instead of standard out
  -s string
    	Satellite ID as provided by StellarStation (default "5")
  -x	Exit the program after a pass ends
```

* CTRL-C will stop the session at any time