# C# example

## Requirements

- .NET Core > 6.0 SDK installed

## First Setup

- Compile to generate stubs: in run `dotnet build ./StellarStationService`
- Create a configuration file `config.json` as follows:

```
{
    "api_address": "https://api.stellarstation.com",
    "api_key_path": "./api-key.json",
    "ground_stations": [
        {
            "id": 45,
            "name": "GS_001"
        },
        {
            "id": 46,
            "name": "GS_002"
        }
    ],
    "satellites": [
        {
            "id": 297,
            "name": "SAT_001"
        },
        {
            "id": 140,
            "name": "SAT_002"
        }
    ]
}
```

- `api_address`: should always be `"https://api.stellarstation.com"`
- `api_key_path`: path to the API key you generated through `https://www.stellarstation.com/console`.
- `ground_stations`: List of the ground stations your organisation is managing.
- `satellites`: List of the satellites your organisation is managing.

You can now run `.\bin\Debug\net6.0\StellarStationService.exe` after compilation and get the list of available passes for all the satellites in `config.json`.