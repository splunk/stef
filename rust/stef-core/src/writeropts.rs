use std::collections::HashMap;

use crate::{compression::Compression, frameflags::FrameFlags, schema::WireSchema};

/// Fixed signature written at the beginning of every STEF stream.
pub const HDR_SIGNATURE: &str = "STEF";

/// Default maximum frame size (4MiB - 1KiB).
pub const DEFAULT_MAX_FRAME_SIZE: usize = (4 << 20) - 1024;

/// Default maximum total dictionary size.
pub const DEFAULT_MAX_TOTAL_DICT_SIZE: usize = 4 << 20;

/// Controls writer framing, compression, dictionaries, and optional schema descriptor behavior.
#[derive(Debug, Clone)]
pub struct WriterOptions {
    /// If true, include descriptor in header.
    pub include_descriptor: bool,
    /// Frame compression method.
    pub compression: Compression,
    /// Soft maximum uncompressed frame size.
    pub max_uncompressed_frame_byte_size: usize,
    /// Extra behaviors enabled when opening new frame.
    pub frame_restart_flags: FrameFlags,
    /// Soft maximum total dictionary size.
    pub max_total_dict_size: usize,
    /// Optional output wire schema override.
    pub schema: Option<WireSchema>,
    /// Optional user metadata stored in variable header.
    pub user_data: HashMap<String, String>,
}

impl Default for WriterOptions {
    fn default() -> Self {
        Self {
            include_descriptor: false,
            compression: Compression::None,
            max_uncompressed_frame_byte_size: DEFAULT_MAX_FRAME_SIZE,
            frame_restart_flags: FrameFlags(0),
            max_total_dict_size: DEFAULT_MAX_TOTAL_DICT_SIZE,
            schema: None,
            user_data: HashMap::new(),
        }
    }
}
