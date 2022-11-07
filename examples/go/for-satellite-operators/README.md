# Go API examples

## Requirements

- Go > v1.19

## Installation

- Create a configuration file `config.json` as follows:

```
{
    "api_address": "api.stellarstation.com:443",
    "api_key_path": "./api-key.json",
    "ground_stations": [
        {
            "id": 45
        },
        {
            "id": 46
        }
    ],
    "satellites": [
        {
            "id": 297
        },
        {
            "id": 140
        }
    ]
}
```

- `api_address`: should always be `"api.stellarstation.com:443"`
- `api_key_path`: path to the API key you generated through `https://www.stellarstation.com/console`.
- `ground_stations`: List of the ground stations your organisation is managing.
- `satellites`: List of the satellites your organisation is managing.

## Available commands

- List available upcoming passes in JSON format for a given satellite (replace `123` by your satellite ID):

> go run .\cmd\list-upcoming-available-passes\main.go -c .\config.json -id 123