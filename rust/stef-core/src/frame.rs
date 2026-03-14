use std::io::{ErrorKind, Read};

use crate::{
    chunkwriter::ChunkWriter,
    compression::Compression,
    errors::{ERR_FRAME_SIZE_LIMIT, Result, StefError},
    frameflags::FrameFlags,
    limits::FRAME_SIZE_LIMIT,
    membuffer::decode_uvarint,
};

/// End-of-frame sentinel error used by frame payload reader.
#[derive(Debug, thiserror::Error)]
#[error("end of frame")]
pub struct EndOfFrame;

/// Reader trait used by frame decoder for both byte-wise and block-wise access.
pub trait ByteAndBlockReader {
    /// Reads one byte.
    fn read_byte(&mut self) -> std::io::Result<u8>;
    /// Reads a block of bytes.
    fn read_block(&mut self, dst: &mut [u8]) -> std::io::Result<usize>;
}

/// Encodes data into STEF frames.
pub struct FrameEncoder<'a> {
    dest: Option<&'a mut dyn ChunkWriter>,
    uncompressed_size: usize,
    compressed_buf: Vec<u8>,
    frame_content: Vec<u8>,
    compression: Compression,
    hdr_byte: FrameFlags,
}

impl<'a> Default for FrameEncoder<'a> {
    fn default() -> Self {
        Self {
            dest: None,
            uncompressed_size: 0,
            compressed_buf: Vec::new(),
            frame_content: Vec::new(),
            compression: Compression::None,
            hdr_byte: FrameFlags(0),
        }
    }
}

impl<'a> FrameEncoder<'a> {
    /// Initializes encoder for a destination and compression mode.
    pub fn init(&mut self, dest: &'a mut dyn ChunkWriter, compr: Compression) -> Result<()> {
        self.dest = Some(dest);
        self.compression = compr;
        Ok(())
    }

    /// Opens a new frame and records restart flags.
    pub fn open_frame(&mut self, reset_flags: FrameFlags) {
        self.hdr_byte = reset_flags;
        self.frame_content.clear();
        self.compressed_buf.clear();
    }

    /// Finalizes the frame and writes it as one chunk.
    pub fn close_frame(&mut self) -> Result<()> {
        let mut frame_hdr = Vec::with_capacity(1 + 20);
        frame_hdr.push(self.hdr_byte.0);
        append_uvarint(&mut frame_hdr, self.uncompressed_size as u64);

        let content: &[u8] = match self.compression {
            Compression::None => &self.frame_content,
            Compression::Zstd => {
                self.compressed_buf = zstd::encode_all(std::io::Cursor::new(&self.frame_content), 0)?;
                append_uvarint(&mut frame_hdr, self.compressed_buf.len() as u64);
                &self.compressed_buf
            }
        };

        self.dest
            .as_deref_mut()
            .ok_or_else(|| StefError::message("frame encoder is not initialized"))?
            .write_chunk(&frame_hdr, content)?;

        self.frame_content.clear();
        self.uncompressed_size = 0;
        Ok(())
    }

    /// Writes payload bytes into current frame.
    pub fn write(&mut self, p: &[u8]) -> Result<usize> {
        self.frame_content.extend_from_slice(p);
        self.uncompressed_size += p.len();
        Ok(p.len())
    }

    /// Returns current frame uncompressed size in bytes.
    pub fn uncompressed_size(&self) -> usize {
        self.uncompressed_size
    }
}

/// Decoder for STEF frames over a byte stream.
pub struct FrameDecoder<R: ByteAndBlockReader> {
    src: Option<R>,
    compression: Compression,
    uncompressed_size: u64,
    frame_content: Vec<u8>,
    frame_ofs: usize,
    pub flags: FrameFlags,
}

impl<R: ByteAndBlockReader> Default for FrameDecoder<R> {
    fn default() -> Self {
        Self {
            src: None,
            compression: Compression::None,
            uncompressed_size: 0,
            frame_content: Vec::new(),
            frame_ofs: 0,
            flags: FrameFlags(0),
        }
    }
}

impl<R: ByteAndBlockReader> FrameDecoder<R> {
    /// Initializes decoder with source and compression mode.
    pub fn init(&mut self, src: R, compression: Compression) -> Result<()> {
        self.src = Some(src);
        self.compression = compression;
        Ok(())
    }

    fn next_frame(&mut self) -> Result<()> {
        let src = self
            .src
            .as_mut()
            .ok_or_else(|| StefError::message("frame decoder not initialized"))?;
        let hdr_byte = src.read_byte()?;
        self.flags = FrameFlags(hdr_byte);
        if !self.flags.is_valid() {
            return Err(StefError::message("invalid frame flags"));
        }

        let uncompressed_size = read_uvarint_from(src)?;
        if uncompressed_size > FRAME_SIZE_LIMIT {
            return Err(ERR_FRAME_SIZE_LIMIT);
        }

        let payload = if self.compression == Compression::None {
            read_exact_from(src, uncompressed_size as usize)?
        } else {
            let compressed_size = read_uvarint_from(src)?;
            if compressed_size > FRAME_SIZE_LIMIT {
                return Err(ERR_FRAME_SIZE_LIMIT);
            }
            let compressed = read_exact_from(src, compressed_size as usize)?;
            zstd::decode_all(std::io::Cursor::new(compressed))?
        };

        self.frame_content = payload;
        self.uncompressed_size = uncompressed_size;
        self.frame_ofs = 0;
        Ok(())
    }

    /// Moves decoder to the next frame.
    pub fn next(&mut self) -> Result<FrameFlags> {
        self.next_frame()?;
        Ok(self.flags)
    }

    /// Returns remaining bytes in current frame.
    pub fn remaining_size(&self) -> u64 {
        self.uncompressed_size.saturating_sub(self.frame_ofs as u64)
    }

    /// Reads frame bytes until frame end.
    pub fn read(&mut self, p: &mut [u8]) -> std::result::Result<usize, StefError> {
        if self.remaining_size() == 0 {
            return Err(StefError::message("end of frame"));
        }
        let to_read = p.len().min(self.remaining_size() as usize);
        p[..to_read].copy_from_slice(&self.frame_content[self.frame_ofs..self.frame_ofs + to_read]);
        self.frame_ofs += to_read;
        Ok(to_read)
    }

    /// Reads one frame byte.
    pub fn read_byte(&mut self) -> std::result::Result<u8, StefError> {
        if self.remaining_size() == 0 {
            return Err(StefError::message("end of frame"));
        }
        let b = self.frame_content[self.frame_ofs];
        self.frame_ofs += 1;
        Ok(b)
    }
}

impl<R: ByteAndBlockReader> ByteAndBlockReader for FrameDecoder<R> {
    fn read_byte(&mut self) -> std::io::Result<u8> {
        FrameDecoder::read_byte(self).map_err(|e| std::io::Error::new(ErrorKind::Other, e.to_string()))
    }

    fn read_block(&mut self, dst: &mut [u8]) -> std::io::Result<usize> {
        FrameDecoder::read(self, dst).map_err(|e| std::io::Error::new(ErrorKind::Other, e.to_string()))
    }
}

fn append_uvarint(dst: &mut Vec<u8>, mut v: u64) {
    loop {
        let mut b = (v & 0x7f) as u8;
        v >>= 7;
        if v != 0 {
            b |= 0x80;
        }
        dst.push(b);
        if v == 0 {
            break;
        }
    }
}

fn read_uvarint_from<R: ByteAndBlockReader>(src: &mut R) -> Result<u64> {
    let mut buf = [0u8; 10];
    for i in 0..10 {
        buf[i] = src.read_byte()?;
        if buf[i] < 0x80 {
            let (v, _) = decode_uvarint(&buf[..=i])?;
            return Ok(v);
        }
    }
    Err(StefError::message("varint overflow"))
}

fn read_exact_from<R: ByteAndBlockReader>(src: &mut R, len: usize) -> Result<Vec<u8>> {
    let mut out = vec![0u8; len];
    let mut ofs = 0;
    while ofs < len {
        let n = src.read_block(&mut out[ofs..])?;
        if n == 0 {
            return Err(std::io::Error::new(ErrorKind::UnexpectedEof, "EOF").into());
        }
        ofs += n;
    }
    Ok(out)
}

impl ByteAndBlockReader for crate::memreaderwriter::MemReaderWriter {
    fn read_byte(&mut self) -> std::io::Result<u8> {
        self.read_byte()
    }

    fn read_block(&mut self, dst: &mut [u8]) -> std::io::Result<usize> {
        self.read(dst)
    }
}
