mod stream;

use clap::Parser;
use stream::stream;
use tracing::info;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

/// Basic example to stream telemetry from StellarStation
#[derive(Parser, Debug, Clone)]
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
    key: String,

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
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer())
        .with(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let args = Args::parse();
    info!(?args, "got args");

    stream(args).await?;

    Ok(())
}
