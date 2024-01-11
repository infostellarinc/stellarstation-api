use std::{env, path::PathBuf, str::FromStr, time::Duration};

use api::SatelliteStreamRequest;
use clap::{Parser, Subcommand};
use google_cloud_auth::project::{create_token_source_from_credentials, Config};
use tokio_stream::wrappers::ReceiverStream;
use tonic::metadata::MetadataValue;
use tracing::info;
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

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(about, long_about = None)]
struct Args {
    /// URL to connect to for streaming
    #[arg(
        long,
        env = "STELLARSTATION_API_URL",
        default_value = "https://api.stellarstation.com",
        value_name = "URL"
    )]
    url: String,

    /// Path to a StellarStation API key
    #[arg(long, env = "STELLARSTATION_API_KEY", value_name = "FILE")]
    key: PathBuf,

    #[command(subcommand)]
    command: Option<Command>,
}

#[derive(Subcommand, Clone, Debug)]
enum Command {
    /// Opens a single stream
    Stream {
        /// Specify a satellite ID with which to Filter telemetry and commands
        #[arg(short = 's', long)]
        satellite_id: String,

        /// Specify a plan ID with which to filter telemetry and commands
        #[arg(short = 'p', long)]
        plan_id: Option<String>,

        /// Enable trying to automatically reconnect if the stream is dropped
        #[arg(short = 'r', long)]
        reconnect: bool,

        /// On the initial connection, use an existing stream ID to reconnect to that stream
        #[arg(long, value_name = "STREAM_ID")]
        reconnect_stream_id: Option<String>,

        /// On the initial connection, set the next expected message index to receive
        #[arg(long, value_name = "MESSAGE_INDEX")]
        reconnect_message_index: Option<String>,

        /// Create multiple streams
        #[arg(long, default_value_t = 1)]
        count: u16,
    },
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer())
        .with(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let args = Args::parse();
    info!(?args, "got args");

    let url = env::var("STELLARSTATION_API_URL")
        .unwrap_or_else(|_| String::from("http://api.stellarstation.com"));

    let endpoint = tonic::transport::Endpoint::new(url)?.user_agent("streamcli")?;
    let channel = endpoint.connect().await?;

    let mut client = api::stellar_station_service_client::StellarStationServiceClient::new(channel)
        .max_decoding_message_size(10 * 1024 * 1024)
        .max_encoding_message_size(10 * 1024 * 1024);

    let (tx, rx) = tokio::sync::mpsc::channel(1);

    let key_path = env::var("STELLARSTATION_API_KEY_PATH")?;

    let creds = google_cloud_auth::credentials::CredentialsFile::new_from_file(key_path).await?;
    let config = Config {
        audience: Some("https://api.stellarstation.com"),
        scopes: None,
        sub: None,
    };
    let ts = create_token_source_from_credentials(&creds, &config).await?;
    let t = ts.token().await?;

    let mut request = tonic::Request::new(ReceiverStream::new(rx));

    request.metadata_mut().insert(
        "authorization",
        MetadataValue::from_str(&format!("Bearer {}", t.access_token))?,
    );

    let res = client.open_satellite_stream(request).await?;

    // Handle received telemetry
    tokio::spawn(async move {
        let mut res = res.into_inner();
        loop {
            tokio::time::sleep(Duration::from_secs(3)).await;
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
            satellite_id: "333".into(),
            // plan_id: "1".into(),
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

    tokio::time::sleep(std::time::Duration::from_secs(60 * 60)).await;

    Ok(())
}
