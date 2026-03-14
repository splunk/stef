use std::sync::Arc;

use stef_core::{ChunkWriter, WriterOptions, schema::{Compatibility, WireSchema}};
use tokio::sync::{Mutex, oneshot};

use crate::{
    proto::{self, stef_server_message::Message, stef_destination_client::StefDestinationClient},
    types::{Logger, NopLogger},
};

/// Client callbacks for stream lifecycle and acks.
pub struct ClientCallbacks {
    /// Called when stream disconnects.
    pub on_disconnect: Arc<dyn Fn(Option<String>) + Send + Sync>,
    /// Called on each ack id.
    pub on_ack: Arc<dyn Fn(u64) -> Result<(), String> + Send + Sync>,
}

impl Default for ClientCallbacks {
    fn default() -> Self {
        Self {
            on_disconnect: Arc::new(|_| {}),
            on_ack: Arc::new(|_| Ok(())),
        }
    }
}

/// Client-side schema settings for handshake.
#[derive(Clone)]
pub struct ClientSchema {
    /// Root struct name.
    pub root_struct_name: String,
    /// Client wire schema.
    pub wire_schema: WireSchema,
}

/// Client initialization settings.
pub struct ClientSettings {
    /// Logger.
    pub logger: Option<Arc<dyn Logger>>,
    /// gRPC destination endpoint.
    pub grpc_endpoint: String,
    /// Client schema.
    pub client_schema: ClientSchema,
    /// Event callbacks.
    pub callbacks: ClientCallbacks,
}

/// Send-side wrapper error to distinguish transport failures.
#[derive(Debug, thiserror::Error)]
#[error("stefgrpc write error: {0}")]
pub struct SendError(pub String);

/// STEF gRPC client with handshake and ack receiver loop.
pub struct Client {
    grpc_client: Option<StefDestinationClient<tonic::transport::Channel>>,
    callbacks: ClientCallbacks,
    client_schema: ClientSchema,
    logger: Arc<dyn Logger>,
    cancel_tx: Option<oneshot::Sender<()>>,
    wait_rx: Option<oneshot::Receiver<()>>,
}

impl Client {
    /// Creates client from settings.
    pub async fn new(settings: ClientSettings) -> Result<Self, String> {
        if settings.client_schema.root_struct_name.is_empty() {
            return Err("client schema root struct name is empty".into());
        }

        let channel = tonic::transport::Endpoint::new(settings.grpc_endpoint.clone())
            .map_err(|e| format!("invalid endpoint: {e}"))?
            .connect()
            .await
            .map_err(|e| format!("connect failed: {e}"))?;

        Ok(Self {
            grpc_client: Some(StefDestinationClient::new(channel)),
            callbacks: settings.callbacks,
            client_schema: settings.client_schema,
            logger: settings.logger.unwrap_or_else(|| Arc::new(NopLogger)),
            cancel_tx: None,
            wait_rx: None,
        })
    }

    /// Connects stream, performs schema handshake, and returns chunk writer + writer options.
    pub async fn connect(&mut self) -> Result<(GrpcWriter, WriterOptions), String> {
        self.logger.debugf("Begin connecting");

        let mut opts = WriterOptions::default();

        let client = self.grpc_client.as_mut().ok_or_else(|| "grpc client not initialized".to_string())?;
        let (tx, rx) = tokio::sync::mpsc::channel::<proto::StefClientMessage>(16);

        let first_message = proto::StefClientMessage {
            first_message: Some(proto::StefClientFirstMessage { root_struct_name: self.client_schema.root_struct_name.clone() }),
            stef_bytes: vec![],
            is_end_of_chunk: false,
        };
        tx.send(first_message).await.map_err(|e| format!("failed to send to server: {e}"))?;

        let outbound = tokio_stream::wrappers::ReceiverStream::new(rx);
        let response = client.stream(outbound).await.map_err(|e| format!("failed to gRPC stream: {e}"))?;
        let mut inbound = response.into_inner();

        let message = inbound.message().await.map_err(|e| format!("error received from server: {e}"))?.ok_or_else(|| "invalid server response".to_string())?;
        let capabilities = match message.message {
            Some(Message::Capabilities(c)) => c,
            _ => return Err("invalid server response".into()),
        };

        if let Some(limits) = capabilities.dictionary_limits {
            opts.max_total_dict_size = limits.max_dict_bytes as usize;
        }

        let mut server_schema = WireSchema::default();
        server_schema
            .deserialize(&capabilities.schema)
            .map_err(|e| format!("failed to unmarshal capabilities schema: {e}"))?;

        match server_schema.compatible(&self.client_schema.wire_schema) {
            Ok(Compatibility::Exact) => {}
            Ok(Compatibility::Superset) => {
                opts.include_descriptor = true;
                opts.schema = Some(self.client_schema.wire_schema.clone());
            }
            Err(_) | Ok(Compatibility::Incompatible) => {
                match self.client_schema.wire_schema.compatible(&server_schema) {
                    Ok(Compatibility::Superset) => {
                        opts.include_descriptor = true;
                        opts.schema = Some(self.client_schema.wire_schema.clone());
                    }
                    _ => return Err("client and server schemas are incompatble".into()),
                }
            }
        }

        let sender = Arc::new(Mutex::new(tx));
        let (cancel_tx, mut cancel_rx) = oneshot::channel::<()>();
        let (wait_tx, wait_rx) = oneshot::channel::<()>();
        self.cancel_tx = Some(cancel_tx);
        self.wait_rx = Some(wait_rx);

        let callbacks = ClientCallbacks {
            on_disconnect: self.callbacks.on_disconnect.clone(),
            on_ack: self.callbacks.on_ack.clone(),
        };
        let logger = self.logger.clone();

        tokio::spawn(async move {
            loop {
                tokio::select! {
                    _ = &mut cancel_rx => {
                        break;
                    }
                    msg = inbound.message() => {
                        match msg {
                            Ok(Some(resp)) => {
                                let ack_id = match resp.message {
                                    Some(Message::Response(r)) => r.ack_record_id,
                                    _ => {
                                        (callbacks.on_disconnect)(Some("invalid server response".into()));
                                        break;
                                    }
                                };
                                if let Err(e) = (callbacks.on_ack)(ack_id) {
                                    (callbacks.on_disconnect)(Some(e));
                                    break;
                                }
                            }
                            Ok(None) => {
                                (callbacks.on_disconnect)(None);
                                break;
                            }
                            Err(e) => {
                                logger.errorf(&format!("Error receiving acks: {e}"));
                                (callbacks.on_disconnect)(Some(e.to_string()));
                                break;
                            }
                        }
                    }
                }
            }
            let _ = wait_tx.send(());
        });

        Ok((GrpcWriter { sender }, opts))
    }

    /// Disconnects stream and waits for receive task completion.
    pub async fn disconnect(&mut self) -> Result<(), String> {
        if let Some(tx) = self.cancel_tx.take() {
            let _ = tx.send(());
        }
        if let Some(wait_rx) = self.wait_rx.take() {
            let _ = wait_rx.await;
        }
        Ok(())
    }
}

/// Chunk writer backed by gRPC client stream.
pub struct GrpcWriter {
    sender: Arc<Mutex<tokio::sync::mpsc::Sender<proto::StefClientMessage>>>,
}

impl ChunkWriter for GrpcWriter {
    fn write_chunk(&mut self, header: &[u8], content: &[u8]) -> stef_core::errors::Result<()> {
        let mut bytes = Vec::with_capacity(header.len() + content.len());
        bytes.extend_from_slice(header);
        bytes.extend_from_slice(content);

        let msg = proto::StefClientMessage { first_message: None, stef_bytes: bytes, is_end_of_chunk: true };
        let sender = self.sender.clone();
        tokio::runtime::Handle::current().block_on(async move {
            sender.lock().await.send(msg).await.map_err(|e| stef_core::errors::StefError::message(SendError(e.to_string()).to_string()))
        })
    }
}
