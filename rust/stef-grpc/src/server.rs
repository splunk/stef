use std::sync::Arc;

use async_trait::async_trait;
use parking_lot::RwLock;
use stef_core::schema::WireSchema;
use tokio::sync::mpsc;
use tokio_stream::{StreamExt, wrappers::ReceiverStream};
use tonic::{Request, Response, Status};

use crate::{
    proto::{self, stef_server_message::Message, stef_destination_server::StefDestination, stef_destination_server::StefDestinationServer},
    types::{Logger, NopLogger},
};

/// Aggregated stream stats.
#[derive(Debug, Clone, Copy, Default)]
pub struct GrpcReaderStats {
    pub messages_received: u64,
    pub bytes_received: u64,
}

/// Reader abstraction for assembled chunks.
pub trait GrpcReader: Send {
    fn read(&mut self, p: &mut [u8]) -> Result<usize, Status>;
    fn stats(&self) -> GrpcReaderStats;
}

#[async_trait]
pub trait STEFStream: Send + Sync {
    async fn send_data_response(&self, response: proto::StefDataResponse) -> Result<(), Status>;
}

/// Server callbacks.
#[derive(Default)]
pub struct Callbacks {
    pub on_stream: Option<Arc<dyn Fn(Box<dyn GrpcReader>, Arc<dyn STEFStream>) -> Result<(), Status> + Send + Sync>>,
}

/// Server configuration.
pub struct ServerSettings {
    pub logger: Option<Arc<dyn Logger>>,
    pub server_schema: WireSchema,
    pub max_dict_bytes: u64,
    pub callbacks: Callbacks,
}

/// Tonic service implementation.
pub struct StreamServer {
    logger: Arc<dyn Logger>,
    server_schema: WireSchema,
    max_dict_bytes: u64,
    callbacks: Callbacks,
}

impl StreamServer {
    pub fn new(settings: ServerSettings) -> Self {
        Self {
            logger: settings.logger.unwrap_or_else(|| Arc::new(NopLogger)),
            server_schema: settings.server_schema,
            max_dict_bytes: settings.max_dict_bytes,
            callbacks: settings.callbacks,
        }
    }

    pub fn service(self) -> StefDestinationServer<Self> {
        StefDestinationServer::new(self)
    }
}

struct ChunkAssembler {
    rx: mpsc::Receiver<Vec<u8>>,
    buf: Vec<u8>,
    read_index: usize,
    stats: RwLock<GrpcReaderStats>,
}

impl GrpcReader for ChunkAssembler {
    fn read(&mut self, p: &mut [u8]) -> Result<usize, Status> {
        if self.read_index >= self.buf.len() {
            self.buf = self.rx.blocking_recv().ok_or_else(|| Status::internal("stream closed"))?;
            self.read_index = 0;
            let mut stats = self.stats.write();
            stats.messages_received += 1;
            stats.bytes_received += self.buf.len() as u64;
        }
        let n = p.len().min(self.buf.len() - self.read_index);
        p[..n].copy_from_slice(&self.buf[self.read_index..self.read_index + n]);
        self.read_index += n;
        Ok(n)
    }

    fn stats(&self) -> GrpcReaderStats {
        *self.stats.read()
    }
}

struct GrpcStreamResponder {
    tx: mpsc::Sender<Result<proto::StefServerMessage, Status>>,
}

#[async_trait]
impl STEFStream for GrpcStreamResponder {
    async fn send_data_response(&self, response: proto::StefDataResponse) -> Result<(), Status> {
        self.tx
            .send(Ok(proto::StefServerMessage { message: Some(Message::Response(response)) }))
            .await
            .map_err(|e| Status::internal(format!("send failed: {e}")))
    }
}

#[tonic::async_trait]
impl StefDestination for StreamServer {
    type StreamStream = ReceiverStream<Result<proto::StefServerMessage, Status>>;

    async fn stream(
        &self,
        request: Request<tonic::Streaming<proto::StefClientMessage>>,
    ) -> Result<Response<Self::StreamStream>, Status> {
        let mut inbound = request.into_inner();

        let first = inbound
            .next()
            .await
            .ok_or_else(|| Status::invalid_argument("missing first message"))??;

        let first_msg = first
            .first_message
            .ok_or_else(|| Status::invalid_argument("FirstMessage is nil"))?;
        if first_msg.root_struct_name.is_empty() {
            return Err(Status::invalid_argument("RootStructName is unspecified"));
        }

        let mut schema_bytes = Vec::new();
        self.server_schema.serialize(&mut schema_bytes);

        let (server_tx, server_rx) = mpsc::channel::<Result<proto::StefServerMessage, Status>>(32);
        let (chunk_tx, chunk_rx) = mpsc::channel::<Vec<u8>>(32);

        server_tx
            .send(Ok(proto::StefServerMessage {
                message: Some(Message::Capabilities(proto::StefDestinationCapabilities {
                    dictionary_limits: Some(proto::StefDictionaryLimits { max_dict_bytes: self.max_dict_bytes }),
                    schema: schema_bytes,
                })),
            }))
            .await
            .map_err(|e| Status::internal(format!("cannot send message to the client: {e}")))?;

        tokio::spawn(async move {
            let mut chunk_buf = Vec::new();
            while let Some(msg) = inbound.next().await {
                match msg {
                    Ok(m) => {
                        chunk_buf.extend_from_slice(&m.stef_bytes);
                        if m.is_end_of_chunk {
                            if chunk_tx.send(std::mem::take(&mut chunk_buf)).await.is_err() {
                                break;
                            }
                        }
                    }
                    Err(_) => break,
                }
            }
        });

        let reader: Box<dyn GrpcReader> = Box::new(ChunkAssembler {
            rx: chunk_rx,
            buf: Vec::new(),
            read_index: 0,
            stats: RwLock::new(GrpcReaderStats::default()),
        });

        let stream: Arc<dyn STEFStream> = Arc::new(GrpcStreamResponder { tx: server_tx.clone() });
        if let Some(on_stream) = &self.callbacks.on_stream {
            let on_stream = on_stream.clone();
            let callback_tx = server_tx.clone();
            tokio::task::spawn_blocking(move || {
                if let Err(status) = on_stream(reader, stream) {
                    let _ = callback_tx.blocking_send(Err(status));
                }
            });
        }

        Ok(Response::new(ReceiverStream::new(server_rx)))
    }
}
