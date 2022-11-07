# Go API examples

## Requirements

- Go > v1.19

## Installation

- Create a configuration file `config.json` as follows:

```
{
    "api_address": "api.stellarstation.com:443",
    "api_key_path": "./api-key.json"
}
```

- `api_address`: should always be `"api.stellarstation.com:443"`
- `api_key_path`: path to the API key you generated through `https://www.stellarstation.com/console`.

## Available commands

- List available upcoming passes in JSON format for a given satellite (replace `123` by your satellite ID):

> go run .\cmd\list-upcoming-available-passes\main.go -c .\config.json -id 123