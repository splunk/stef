//! gRPC transport for STEF streams.

pub mod client;
pub mod server;
pub mod types;

/// Generated protobuf bindings for `destination.proto`.
pub mod proto {
    tonic::include_proto!("stef");
}

pub use client::{Client, ClientCallbacks, ClientSchema, ClientSettings, SendError};
pub use server::{Callbacks, GrpcReader, GrpcReaderStats, STEFStream, ServerSettings, StreamServer};
