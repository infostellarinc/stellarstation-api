# Go API examples

## Requirements

- Go > v1.23

## Arguments

- `addr`: defaults to `api.stellarstation.com:443`
- `key`: path to the API key you generated through `https://www.stellarstation.com/console`.

## Available commands

View the available args with `-h`
> go run .\cmd\list-upcoming-available-passes\main.go -h

Example execution:
> go run .\cmd\list-upcoming-available-passes\main.go -key=/path/to/apikey.json -id <satelliteId>
