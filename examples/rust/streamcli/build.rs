const PROTO_PATH: &str = "../../../api/src/main/proto";

// stellarstation
// └── api
//     └── v1
//         ├── antenna
//         │   └── antenna.proto
//         ├── common
//         │   └── common.proto
//         ├── groundstation
//         │   └── groundstation.proto
//         ├── monitoring
//         │   └── monitoring.proto
//         ├── orbit
//         │   └── orbit.proto
//         ├── radio
//         │   └── radio.proto
//         ├── stellarstation.proto
//         └── transport.proto
fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .compile(
            &[
                format!("{PROTO_PATH}/stellarstation/api/v1/stellarstation.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/transport.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/radio/radio.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/orbit/orbit.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/monitoring/monitoring.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/groundstation/groundstation.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/common/common.proto"),
                format!("{PROTO_PATH}/stellarstation/api/v1/antenna/antenna.proto"),
            ],
            &[PROTO_PATH],
        )?;
    Ok(())
}
