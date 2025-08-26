use std::{
    collections::{HashMap, HashSet},
    fs,
    str::FromStr,
    time::{Duration, SystemTime, UNIX_EPOCH},
};

use tokio_util::sync::CancellationToken;
use tonic::{
    metadata::MetadataValue,
    transport::{Certificate, ClientTlsConfig, Endpoint, Identity},
    IntoRequest, Request,
};
use tracing::{error, info};

use crate::{
    stream::api::{
        reservations::{
            reservation_service_client::ReservationServiceClient, CancelReservationRequest,
            ChannelSetToken, CreateReservationRequest, GetReservationRequest, ListPassesRequest,
            ListReservationsRequest, Reservation, TimeRange, UpdateReservationRequest,
        },
        stellar_station_service_client::StellarStationServiceClient,
    },
    token_source, Args,
};

pub async fn list_reservations_demo(args: Args) -> anyhow::Result<()> {
    // Create an OAuth2 token source to produce bearer tokens for authentication
    let tokens = token_source(args.key, &args.url).await?;
    let token = tokens.token().await?.access_token;

    // const BASE: &str = "/home/dezyh/code/infostellar/stellarstation-core/dev/tls";
    let base = match std::env::var("STREAMCLI_TLS_DIR") {
        Ok(value) => value,
        Err(e) => panic!("requires STREAMCLI_TLS_DIR to be set and point to a directory containing ca.crt, cert.crt, cert.pem"),
    };

    // Load your root CA (used to verify the server)
    let ca_cert = fs::read(format!("{base}/ca.crt"))?;
    let ca_cert = Certificate::from_pem(ca_cert);

    // Optionally, load your own client cert + key for mTLS
    let client_cert = fs::read(format!("{base}/cert.crt"))?;
    let client_key = fs::read(format!("{base}/cert.pem"))?;
    let identity = Identity::from_pem(client_cert, client_key);

    // Build TLS config
    let tls = ClientTlsConfig::new()
        .ca_certificate(ca_cert)
        .identity(identity) // omit this if you don’t need client auth
        .domain_name("localhost"); // must match server cert

    // Build channel with TLS
    let channel = Endpoint::new(args.url.clone())?
        .user_agent("streamcli")?
        .tls_config(tls)?
        .connect()
        .await?;

    // By default, GRPC sets the max message size to 4MB, but StellarStation can support up to 10MB.
    // If GRPC message would be which exceeds this GRPC limit, a RESOURCE_EXHAUSTED error will be returned.
    let mut client = ReservationServiceClient::new(channel)
        .max_decoding_message_size(10 * 1024 * 1024)
        .max_encoding_message_size(10 * 1024 * 1024);

    let start = system_time_to_timestamp(SystemTime::now());
    let stop = system_time_to_timestamp(SystemTime::now() + Duration::from_secs(24 * 60 * 60));

    // =========  1. List no reservations =========
    let req = ListReservationsRequest {
        satellite_id: "9".into(),
        visibility: Some(TimeRange {
            start: Some(start.clone()),
            end: Some(stop.clone()),
        }),
    };
    println!("{:#?}", req);
    let req = request(req, &token);
    let res = client
        .list_reservations(req)
        .await
        .expect("can list reservations")
        .into_inner();
    println!("{:#?}", res);

    // ========= 2. List passes =========
    let req = ListPassesRequest {
        satellite_id: "9".into(),
    };
    println!("{:#?}", req);
    let req = request(req, &token);
    let res = client.list_passes(req).await;
    let passes = res.expect("can list passes").into_inner().passes;
    let pass = passes.first().expect("can list at least 1 pass");
    println!("ListPassesResponse[0] -> {:#?}", pass);

    // ========= 3. Schedule as many channel sets from the 1st pass as possible... =========

    // 3a. group the channel sets into slots
    let mut slots: HashMap<&str, Vec<&ChannelSetToken>> = HashMap::new();
    for channel_set in pass.channel_sets.iter() {
        slots
            .entry(&channel_set.slot)
            .or_default()
            .push(channel_set);
    }

    // 3b. pick the first channel set from each slot
    let channel_sets: Vec<ChannelSetToken> = slots
        .into_iter()
        .filter_map(|(_slot, tokens)| tokens.first().copied())
        .map(ChannelSetToken::to_owned)
        .collect();

    // 3c. send the request
    let req = CreateReservationRequest { channel_sets };
    println!("{:#?}", req);
    let req = request(req, &token);
    let res = client
        .create_reservation(req)
        .await
        .expect("can create reservation")
        .into_inner();
    println!("{:#?}", res);

    // ========= 4. Try getting our request =========
    let reservation = res
        .reservation
        .expect("create reservation returns a reservation");

    let req = GetReservationRequest {
        reservation_id: reservation.id.clone(),
    };
    println!("{:#?}", req);
    let req = request(req, &token);
    let res = client
        .get_reservation(req)
        .await
        .expect("can get reservation")
        .into_inner();
    println!("{:#?}", res);

    // ========= 5. Update our reservation, by dropping the first channel-set =========
    let updated_channel_sets: Vec<_> = reservation
        .channel_sets
        .iter()
        .filter(|channel_set| channel_set.scheduled)
        .skip(1)
        .map(ChannelSetToken::to_owned)
        .collect();

    let req = UpdateReservationRequest {
        reservation_id: reservation.id.clone(),
        channel_sets: updated_channel_sets,
    };
    println!("{:#?}", req);
    let req = request(req, &token);
    let res = client
        .update_reservation(req)
        .await
        .expect("can update rservation")
        .into_inner();
    println!("{:#?}", res);

    // ========= 6. Delete the reservation =========
    let req = CancelReservationRequest {
        reservation_id: reservation.id.clone(),
    };
    println!("{:#?}", req);
    let req = request(req, &token);
    let res = client
        .cancel_reservation(req)
        .await
        .expect("can cancel reservation")
        .into_inner();
    println!("{:#?}", res);

    Ok(())
}

fn system_time_to_timestamp(time: SystemTime) -> prost_types::Timestamp {
    let duration = time.duration_since(UNIX_EPOCH).unwrap();
    prost_types::Timestamp {
        seconds: duration.as_secs() as i64,
        nanos: duration.subsec_nanos() as i32,
    }
}

fn request<T>(t: T, token: &str) -> Request<T> {
    let mut req = Request::new(t);
    req.metadata_mut().insert(
        "authorization",
        MetadataValue::from_str(&format!("Bearer {token}")).expect("valid bearer token"),
    );
    req
}
