# Fake ground station server for testing Ground Station API clients

Currently, this fake server only implements the following API calls:
* `ListPlans`
* `OpenGroundStationStream`

Currently, all `ListPlans` calls return a plan with the following properties:
* Plan ID: 3
* Starts 10 seconds after the `ListPlans` call
* 10 minute duration


## Install StellarStation API library
To run the fake server, you need stubs generated from .proto file. To install precompiled client stubs for Python, run:

```bash
$  pip install --upgrade stellarstation
```

## Try it out!
To start the server, run the following command:
```bash
$  python ground_station_service.py
```
