use std::io::Write;

use crate::errors::Result;

/// Writes header+content chunks into an output stream.
pub trait ChunkWriter {
    /// Writes chunk header followed by chunk content.
    fn write_chunk(&mut self, header: &[u8], content: &[u8]) -> Result<()>;
}

/// In-memory chunk sink.
#[derive(Debug, Default, Clone)]
pub struct MemChunkWriter {
    buf: Vec<u8>,
}

impl MemChunkWriter {
    /// Returns accumulated bytes.
    pub fn bytes(&self) -> &[u8] {
        &self.buf
    }
}

impl ChunkWriter for MemChunkWriter {
    fn write_chunk(&mut self, header: &[u8], content: &[u8]) -> Result<()> {
        self.buf.extend_from_slice(header);
        self.buf.extend_from_slice(content);
        Ok(())
    }
}

/// Adapter from `std::io::Write` into `ChunkWriter`.
pub struct WrapChunkWriter<W: Write> {
    w: W,
}

impl<W: Write> WrapChunkWriter<W> {
    /// Creates a chunk writer around an `io::Write` implementation.
    pub fn new(w: W) -> Self {
        Self { w }
    }
}

impl<W: Write> ChunkWriter for WrapChunkWriter<W> {
    fn write_chunk(&mut self, header: &[u8], content: &[u8]) -> Result<()> {
        self.w.write_all(header)?;
        self.w.write_all(content)?;
        Ok(())
    }
}
