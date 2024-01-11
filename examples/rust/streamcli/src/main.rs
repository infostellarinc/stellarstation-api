use std::{env, str::FromStr};

use api::SatelliteStreamRequest;
use tokio_stream::wrappers::ReceiverStream;
use tonic::metadata::{AsciiMetadataValue, MetadataValue};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

/// Include the rust code generated from the stellarstation-api protos by `prost`.
/// In the future we will replace this with a crate hosted on cargo.
pub mod api {
    tonic::include_proto!("stellarstation.api.v1");
    pub mod radio {
        tonic::include_proto!("stellarstation.api.v1.radio");
    }
    pub mod antenna {
        tonic::include_proto!("stellarstation.api.v1.antenna");
    }
    pub mod monitoring {
        tonic::include_proto!("stellarstation.api.v1.monitoring");
    }
    pub mod orbit {
        tonic::include_proto!("stellarstation.api.v1.orbit");
    }
    pub mod common {
        tonic::include_proto!("stellarstation.api.v1.common");
    }
    pub mod groundstation {
        tonic::include_proto!("stellarstation.api.v1.groundstation");
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer())
        .with(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let url = env::var("STELLARSTATION_API_URL")
        .unwrap_or_else(|_| String::from("http://api.stellarstation.com"));

    let endpoint = tonic::transport::Endpoint::new(url)?.user_agent("streamcli")?;
    let channel = endpoint.connect().await?;

    let mut client = api::stellar_station_service_client::StellarStationServiceClient::new(channel)
        .max_decoding_message_size(10 * 1024 * 1024)
        .max_encoding_message_size(10 * 1024 * 1024);

    let (tx, rx) = tokio::sync::mpsc::channel(1);

    let mut request = tonic::Request::new(ReceiverStream::new(rx));

    request
        .metadata_mut()
        .insert("authorization", MetadataValue::from_str("Bearer 123")?);

    let res = client.open_satellite_stream(request).await?;

    // Handle received telemetry
    tokio::spawn(async move {
        let mut res = res.into_inner();
        loop {
            match res.message().await {
                Ok(Some(res)) => println!("stream received message: {:?}", res),
                Ok(None) => {
                    println!("stream closed gracefully by server");
                    break;
                }
                Err(err) => {
                    println!("stream closed by server with error: {:?}", err);
                    break;
                }
            }
        }
    });

    tokio::time::sleep(std::time::Duration::from_secs(5)).await;
    let req = tx
        .send(SatelliteStreamRequest {
            satellite_id: "1".into(),
            plan_id: "1".into(),
            // enable_events: true,
            // stream_id: todo!(),
            // ground_station_id: todo!(),
            // request_id: todo!(),
            // accepted_framing: todo!(),
            // resume_stream_message_ack_id: todo!(),
            // enable_flow_control: todo!(),
            // request: todo!(),
            ..Default::default()
        })
        .await;

    println!("send request with result: {:?}", req);

    tokio::time::sleep(std::time::Duration::from_secs(5)).await;

    Ok(())
}
