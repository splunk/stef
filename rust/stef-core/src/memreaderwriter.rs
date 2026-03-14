use std::io::{Cursor, Read};

use crate::{chunkwriter::ChunkWriter, errors::Result};

/// In-memory duplex object for tests: write chunks, then read bytes.
#[derive(Debug, Default, Clone)]
pub struct MemReaderWriter {
    cursor: Cursor<Vec<u8>>,
}

impl MemReaderWriter {
    /// Returns full buffer bytes.
    pub fn bytes(&self) -> &[u8] {
        self.cursor.get_ref()
    }

    /// Appends raw bytes to backing buffer.
    pub fn append_bytes(&mut self, data: &[u8]) {
        self.cursor.get_mut().extend_from_slice(data);
    }

    /// Sets read position.
    pub fn set_position(&mut self, pos: u64) {
        self.cursor.set_position(pos);
    }

    /// Reads one byte.
    pub fn read_byte(&mut self) -> std::io::Result<u8> {
        let mut b = [0u8; 1];
        let n = self.cursor.read(&mut b)?;
        if n == 0 {
            return Err(std::io::Error::new(std::io::ErrorKind::UnexpectedEof, "EOF"));
        }
        Ok(b[0])
    }
}

impl Read for MemReaderWriter {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        self.cursor.read(buf)
    }
}

impl ChunkWriter for MemReaderWriter {
    fn write_chunk(&mut self, header: &[u8], content: &[u8]) -> Result<()> {
        self.cursor.get_mut().extend_from_slice(header);
        self.cursor.get_mut().extend_from_slice(content);
        Ok(())
    }
}
