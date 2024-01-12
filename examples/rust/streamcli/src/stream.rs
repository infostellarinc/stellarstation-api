use crate::Args;

use google_cloud_auth::{
    credentials::CredentialsFile,
    project::{create_token_source_from_credentials, Config},
    token_source::TokenSource,
};
use std::{str::FromStr, time::Duration};
use tokio::{select, sync::mpsc};
use tokio_stream::wrappers::ReceiverStream;
use tokio_util::sync::CancellationToken;
use tonic::{
    metadata::MetadataValue,
    transport::{Channel, Endpoint},
    Request,
};
use tracing::{error, info};

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
use api::stellar_station_service_client::StellarStationServiceClient;

async fn token_source(key: String, url: &str) -> anyhow::Result<Box<dyn TokenSource>> {
    let creds = CredentialsFile::new_from_file(key).await?;

    let config = Config {
        audience: Some(url),
        scopes: None,
        sub: None,
    };

    Ok(create_token_source_from_credentials(&creds, &config).await?)
}

struct StreamConfig {
    channel: Channel,
    bearer: String,
    satellite: String,
}

async fn one_stream(i: u16, ctx: CancellationToken, config: StreamConfig) -> anyhow::Result<()> {
    let mut client = StellarStationServiceClient::new(config.channel)
        .max_decoding_message_size(10 * 1024 * 1024)
        .max_encoding_message_size(10 * 1024 * 1024);

    let (tx, rx) = mpsc::channel(1);
    let mut request = Request::new(ReceiverStream::new(rx));

    // Use the OAuth2 bearer token generated from the StellarStation API key to authorize the
    // stream.
    request.metadata_mut().insert(
        "authorization",
        MetadataValue::from_str(&format!("Bearer {}", config.bearer))?,
    );

    let response = client.open_satellite_stream(request).await?;
    info!(i, "connected to stellarstation");

    let rx_ctx = ctx.clone();
    let rx_handle = tokio::spawn(async move {
        let mut rx = response.into_inner();

        loop {
            select! {
                msg = rx.message() => match msg {
                    Ok(Some(msg)) => info!(?msg, "stream received message from server"),
                    Ok(None) => {
                        info!("stream closed gracefully by server");
                        break;
                    }
                    Err(err) => {
                        error!("stream closed by server with error: {:?}", err);
                        break;
                    }
                },
                _ = rx_ctx.cancelled() => {
                    info!("received exit signal, client initiating stream close");
                    break;
                }
            }
        }
    });

    let tx_ctx = ctx.clone();
    let tx_handle = tokio::spawn(async move {
        let initial_stream_request = api::SatelliteStreamRequest {
            satellite_id: "333".into(),
            ..Default::default()
        };

        match tx.send(initial_stream_request).await {
            Ok(_) => info!("sent initial stream request"),
            Err(err) => error!(?err, "failed to send initial stream request"),
        };

        tx_ctx.cancelled().await;
    });

    tokio::time::sleep(Duration::from_secs(30)).await;

    tx_handle.await;
    rx_handle.await;

    Ok(())
}

pub async fn stream(args: Args) -> anyhow::Result<()> {
    // Create an OAuth2 token source to produce bearer tokens for authentication
    let tokens = token_source(args.key, &args.url).await?;

    // Create a reusable channel for gRPC connections
    let channel = Endpoint::new(args.url.clone())?
        .user_agent("streamcli")?
        .connect()
        .await?;

    let ctx = CancellationToken::new();

    let mut tasks = Vec::new();
    for i in 0..args.count {
        let task = tokio::spawn(one_stream(
            i,
            ctx.child_token(),
            StreamConfig {
                channel: channel.clone(),
                bearer: tokens.token().await?.access_token,
                satellite: args.satellite_id.clone(),
            },
        ));
        tasks.push(task);
    }

    for task in tasks {
        if let Err(err) = task.await? {
            tracing::warn!(?err, "failed to wait for task?");
        }
    }

    Ok(())
}
