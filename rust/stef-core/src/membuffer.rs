use std::io::{ErrorKind, Read};

/// Lightweight byte-slice reader used by codecs and schema decoding.
#[derive(Debug, Default, Clone)]
pub struct BytesReader {
    buf: Vec<u8>,
    byte_index: usize,
}

impl BytesReader {
    /// Replaces the reader view with `buf`.
    pub fn reset(&mut self, buf: Vec<u8>) {
        self.buf = buf;
        self.byte_index = 0;
    }

    /// Replaces the reader view with `buf` borrowed copy.
    pub fn reset_slice(&mut self, buf: &[u8]) {
        self.buf.clear();
        self.buf.extend_from_slice(buf);
        self.byte_index = 0;
    }

    /// Reads one byte.
    pub fn read_byte(&mut self) -> std::io::Result<u8> {
        if self.byte_index >= self.buf.len() {
            return Err(std::io::Error::new(ErrorKind::UnexpectedEof, "EOF"));
        }
        let b = self.buf[self.byte_index];
        self.byte_index += 1;
        Ok(b)
    }

    /// Reads unsigned varint.
    pub fn read_uvarint(&mut self) -> std::io::Result<u64> {
        decode_uvarint(&self.buf[self.byte_index..]).map(|(v, n)| {
            self.byte_index += n;
            v
        })
    }

    /// Reads zig-zag encoded signed varint.
    pub fn read_varint(&mut self) -> std::io::Result<i64> {
        let x = self.read_uvarint()?;
        Ok(((x >> 1) as i64) ^ (-((x & 1) as i64)))
    }

    /// Reads UTF-8 bytes as string (copying).
    pub fn read_string_bytes(&mut self, byte_size: usize) -> std::io::Result<String> {
        if self.buf.len().saturating_sub(self.byte_index) < byte_size {
            return Err(std::io::Error::new(ErrorKind::UnexpectedEof, "EOF"));
        }
        let s = String::from_utf8_lossy(&self.buf[self.byte_index..self.byte_index + byte_size]).to_string();
        self.byte_index += byte_size;
        Ok(s)
    }

    /// Reads bytes by copying.
    pub fn read_bytes_mapped(&mut self, byte_size: usize) -> std::io::Result<Vec<u8>> {
        if self.buf.len().saturating_sub(self.byte_index) < byte_size {
            return Err(std::io::Error::new(ErrorKind::UnexpectedEof, "EOF"));
        }
        let out = self.buf[self.byte_index..self.byte_index + byte_size].to_vec();
        self.byte_index += byte_size;
        Ok(out)
    }

    /// Reads UTF-8 bytes as string.
    pub fn read_string_mapped(&mut self, byte_size: usize) -> std::io::Result<String> {
        self.read_string_bytes(byte_size)
    }

    /// Maps bytes from another reader.
    pub fn map_bytes_from_mem_buf(&mut self, src: &mut BytesReader, byte_size: usize) -> std::io::Result<()> {
        let mapped = src.read_bytes_mapped(byte_size)?;
        self.buf = mapped;
        self.byte_index = 0;
        Ok(())
    }

    pub(crate) fn remaining(&self) -> &[u8] {
        &self.buf[self.byte_index..]
    }

    /// Resets read cursor to the start of the current buffer.
    pub fn rewind(&mut self) {
        self.byte_index = 0;
    }
}

/// Growable byte writer with varint helpers.
#[derive(Debug, Default, Clone)]
pub struct BytesWriter {
    buf: Vec<u8>,
    byte_index: usize,
}

impl BytesWriter {
    /// Creates writer with reserved capacity.
    pub fn new(cap: usize) -> Self {
        Self { buf: Vec::with_capacity(cap), byte_index: 0 }
    }

    pub fn write_byte(&mut self, b: u8) {
        self.buf.push(b);
    }

    pub fn write_bytes(&mut self, bytes: &[u8]) {
        self.buf.extend_from_slice(bytes);
    }

    pub fn write_string_bytes(&mut self, val: &str) {
        self.buf.extend_from_slice(val.as_bytes());
    }

    pub fn write_uvarint(&mut self, mut value: u64) {
        // Fast path: write directly into spare capacity to avoid per-byte push checks.
        self.buf.reserve(10);
        let start = self.buf.len();
        let mut n = 0usize;

        // SAFETY:
        // - reserve(10) guarantees at least 10 writable bytes in spare capacity.
        // - n is incremented at most 10 times for u64 varint encoding.
        // - set_len(start + n) is called exactly once after fully initializing those bytes.
        unsafe {
            let dst = self.buf.as_mut_ptr().add(start);
            loop {
                let mut b = (value & 0x7f) as u8;
                value >>= 7;
                if value != 0 {
                    b |= 0x80;
                }
                std::ptr::write(dst.add(n), b);
                n += 1;
                if value == 0 {
                    break;
                }
            }
            self.buf.set_len(start + n);
        }
    }

    pub fn write_varint(&mut self, x: i64) {
        let ux = ((x >> 63) as u64) ^ ((x as u64) << 1);
        self.write_uvarint(ux);
    }

    pub fn reset(&mut self) {
        self.buf.clear();
        self.byte_index = 0;
    }

    pub fn reset_and_reserve(&mut self, len: usize) {
        let need_cap = len + 8;
        if self.buf.capacity() < need_cap {
            self.buf.reserve(need_cap - self.buf.capacity());
        }
        self.buf.resize(len, 0);
        self.byte_index = 0;
    }

    pub fn bytes(&self) -> &[u8] {
        &self.buf
    }

    pub fn as_vec(&self) -> Vec<u8> {
        self.buf.clone()
    }

    pub(crate) fn map_bytes_to_bits_reader(
        &mut self,
        dest: &mut crate::bitstream::BitsReader,
        byte_size: usize,
    ) -> std::io::Result<()> {
        if self.buf.len().saturating_sub(self.byte_index) < byte_size {
            return Err(std::io::Error::new(ErrorKind::UnexpectedEof, "EOF"));
        }
        dest.reset(&self.buf[self.byte_index..self.byte_index + byte_size]);
        self.byte_index += byte_size;
        Ok(())
    }
}

pub(crate) fn decode_uvarint(buf: &[u8]) -> std::io::Result<(u64, usize)> {
    let mut x = 0u64;
    let mut s = 0u32;
    for (i, &b) in buf.iter().enumerate() {
        if b < 0x80 {
            if i > 9 || (i == 9 && b > 1) {
                return Err(std::io::Error::new(ErrorKind::InvalidData, "varint overflow"));
            }
            return Ok((x | ((b as u64) << s), i + 1));
        }
        x |= ((b & 0x7f) as u64) << s;
        s += 7;
    }
    Err(std::io::Error::new(ErrorKind::UnexpectedEof, "EOF"))
}

impl Read for BytesReader {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        let n = buf.len().min(self.remaining().len());
        if n == 0 {
            return Ok(0);
        }
        buf[..n].copy_from_slice(&self.remaining()[..n]);
        self.byte_index += n;
        Ok(n)
    }
}
