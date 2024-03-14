use crate::{stream::api::stream_event::Event, Args};

use google_cloud_auth::{
    credentials::CredentialsFile,
    project::{create_token_source_from_credentials, Config},
    token_source::TokenSource,
};
use std::str::FromStr;
use tokio::{select, sync::mpsc};
use tokio_stream::wrappers::ReceiverStream;
use tokio_util::sync::CancellationToken;
use tonic::{
    metadata::MetadataValue,
    transport::{Channel, Endpoint},
    Request, Streaming,
};
use tracing::{debug, error, info, warn};

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
use api::{
    satellite_stream_response::Response,
    stellar_station_service_client::StellarStationServiceClient, SatelliteStreamRequest,
    SatelliteStreamResponse,
};

/// The partial/complete results of a stream.
struct StreamResult {
    /// Was all the expected telemetry received.
    complete: bool,
    /// How many total bytes have been received.
    bytes: usize,
    /// How many total frames have been received.
    frames: usize,
    /// The ID of the stream. Used to reconnect in the event of a disconnect.
    stream_id: Option<String>,
    /// The ID of the last message received on the stream. Used to correctly resume the stream from
    /// the last message received by the client. This ensures exactly-once deliver of messages even
    /// through reconnections.
    stream_resume_id: Option<String>,
}

/// Handle a `stream` command.
pub async fn stream(args: Args) -> anyhow::Result<()> {
    // Create an OAuth2 token source to produce bearer tokens for authentication
    let tokens = token_source(args.key, &args.url).await?;

    let channel = Endpoint::new(args.url.clone())?
        .user_agent("streamcli")?
        .connect()
        .await?;

    let client = StellarStationServiceClient::new(channel)
        .max_decoding_message_size(5 * 1024 * 1024)
        .max_encoding_message_size(5 * 1024 * 1024);

    let ctx = CancellationToken::new();

    let mut tasks = Vec::new();
    for _ in 0..args.count {
        let task = match args.reconnect {
            true => tokio::spawn(stream_with_reconnect(
                ctx.child_token(),
                tokens.token().await?.access_token,
                client.clone(),
                args.satellite_id.clone(),
                args.plan_id.clone(),
                args.reconnect_stream_id.clone(),
                args.reconnect_message_index.clone(),
            )),
            false => tokio::spawn(stream_attempt(
                ctx.child_token(),
                tokens.token().await?.access_token,
                client.clone(),
                args.satellite_id.clone(),
                args.plan_id.clone(),
                args.reconnect_stream_id.clone(),
                args.reconnect_message_index.clone(),
            )),
        };
        tasks.push(task);
    }

    // Wait for all stream tasks to complete
    for task in tasks {
        if let Err(err) = task.await? {
            tracing::warn!(?err, "failed to wait for task?");
        }
    }

    Ok(())
}

/// Generate a new OAuth2 token source from a StellarStation API key file
async fn token_source(key: String, url: &str) -> anyhow::Result<Box<dyn TokenSource>> {
    let creds = CredentialsFile::new_from_file(key).await?;

    let config = Config {
        audience: Some(url),
        scopes: None,
        sub: None,
    };

    Ok(create_token_source_from_credentials(&creds, &config).await?)
}

/// Open a new stream with reconnection attempts.
/// TODO: Implement a exponential backoff on reconnections.
/// TODO: Generate a new token on each reconnection since the tokens are only valid for a short
/// period of time. Without this, a stream may fail to reconnect as the token is no longer valid.
async fn stream_with_reconnect(
    ctx: CancellationToken,
    token: String,
    client: StellarStationServiceClient<Channel>,
    satellite_id: String,
    plan_id: Option<String>,
    stream_id: Option<String>,
    stream_resume_id: Option<String>,
) -> anyhow::Result<StreamResult> {
    let mut stream_results = StreamResult {
        complete: false,
        bytes: 0,
        frames: 0,
        stream_id: None,
        stream_resume_id: None,
    };

    while !stream_results.complete {
        // Either use the original stream reconnection details or use the latest ones from the
        // stream results.
        let stream_id = stream_results.stream_id.or_else(|| stream_id.clone());
        let stream_resume_id = stream_results
            .stream_resume_id
            .or_else(|| stream_resume_id.clone());

        let attempt_results = stream_attempt(
            ctx.clone(),
            token.clone(),
            client.clone(),
            satellite_id.clone(),
            plan_id.clone(),
            stream_id,
            stream_resume_id,
        )
        .await?;

        stream_results.complete = attempt_results.complete;
        stream_results.stream_id = attempt_results.stream_id;
        stream_results.stream_resume_id = attempt_results.stream_resume_id;
        stream_results.bytes += attempt_results.bytes;
        stream_results.frames += attempt_results.frames;
    }

    Ok(stream_results)
}

/// Open a new stream without reconnection attempts.
async fn stream_attempt(
    ctx: CancellationToken,
    token: String,
    mut client: StellarStationServiceClient<Channel>,
    satellite_id: String,
    plan_id: Option<String>,
    stream_id: Option<String>,
    stream_resume_id: Option<String>,
) -> anyhow::Result<StreamResult> {
    let (tx, rx) = mpsc::channel(1);
    let mut request = Request::new(ReceiverStream::new(rx));

    request.metadata_mut().insert(
        "authorization",
        MetadataValue::from_str(&format!("Bearer {token}"))?,
    );

    let response = client.open_satellite_stream(request).await?;

    stream_setup(&tx, satellite_id, plan_id, stream_id, stream_resume_id).await;

    let results = stream_receiver(ctx.clone(), response.into_inner()).await;

    results
}

/// Receive messages from the stream. When the stream is complete or a disconnect occurs the
/// results of the stream are returned to allow for reconnections.
///
/// If you command sending is required, this can be run inside a task with `tokio::spawn` alongside
/// another task for command sending.
async fn stream_receiver(
    ctx: CancellationToken,
    mut rx: Streaming<SatelliteStreamResponse>,
) -> anyhow::Result<StreamResult> {
    let mut results = StreamResult {
        complete: false,
        bytes: 0,
        frames: 0,
        stream_id: None,
        stream_resume_id: None,
    };

    loop {
        select! {
            msg = rx.message() => match msg {
                Ok(Some(msg)) => {
                    on_message(msg, &mut results);
                    if results.complete {
                        info!("received end-of-stream message, stream completed");
                        break
                    }
                },
                Ok(None) => {
                    info!("stream closed gracefully by server");
                    break;
                }
                Err(err) => {
                    error!("stream closed by server with error: {:?}", err);
                    break;
                }
            },
            _ = ctx.cancelled() => {
                info!("received exit signal, client initiating stream close");
                break;
            }
        }
    }

    Ok(results)
}

/// Setup the stream by sending the initial configuration request.
async fn stream_setup(
    tx: &mpsc::Sender<SatelliteStreamRequest>,
    satellite_id: String,
    plan_id: Option<String>,
    stream_id: Option<String>,
    stream_resume_id: Option<String>,
) {
    let initial = SatelliteStreamRequest {
        satellite_id,
        plan_id: plan_id.unwrap_or_default(),
        stream_id: stream_id.unwrap_or_default(),
        resume_stream_message_ack_id: stream_resume_id.unwrap_or_default(),
        ..Default::default()
    };

    match tx.send(initial).await {
        Ok(_) => info!("sent initial stream request"),
        Err(err) => error!(?err, "failed to send initial stream request"),
    };
}

/// Handle a single message received on the stream by updating the current streams results
fn on_message(msg: SatelliteStreamResponse, results: &mut StreamResult) {
    results.stream_id = Some(msg.stream_id);

    match msg.response {
        Some(msg) => match msg {
            Response::ReceiveTelemetryResponse(msg) => {
                let frames = msg.telemetry.len();
                let bytes = msg
                    .telemetry
                    .iter()
                    .map(|frame| frame.data.len())
                    .sum::<usize>();

                debug!(
                    frames,
                    bytes,
                    plan = msg.plan_id,
                    satellite = msg.satellite_id,
                    groundstation = msg.ground_station_id,
                    message = msg.message_ack_id,
                    "received telemetry message"
                );

                results.frames += frames;
                results.bytes += bytes;
                results.complete = frames == 1 && bytes == 0;
                results.stream_resume_id = Some(msg.message_ack_id);
            }
            Response::StreamEvent(event) => {
                debug!("received event message");
                match event.event {
                    Some(Event::CommandSent(event)) => {
                        debug!(?event, "received CommandSent event (currently unsupported)")
                    }
                    Some(Event::PlanMonitoringEvent(event)) => {
                        debug!(?event, "received PlanMonitoringEvent event")
                    }
                    None => warn!("received empty stream event message"),
                }
            }
        },
        None => {}
    };
}
