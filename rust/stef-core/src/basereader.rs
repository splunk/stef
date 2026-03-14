use crate::{
    compression::Compression,
    consts::{HDR_FLAGS_COMPRESSION_METHOD, HDR_FORMAT_VERSION, HDR_FORMAT_VERSION_MASK},
    errors::{ERR_INVALID_COMPRESSION, ERR_INVALID_FORMAT_VERSION, ERR_INVALID_HEADER, ERR_INVALID_HEADER_SIGNATURE, ERR_INVALID_VAR_HEADER, Result},
    frame::{ByteAndBlockReader, FrameDecoder},
    header::{FixedHeader, VarHeader},
    limits::{FIXED_HDR_CONTENT_SIZE_LIMIT, VAR_HDR_CONTENT_SIZE_LIMIT},
    membuffer::decode_uvarint,
    recordbuf::ReadBufs,
    schema,
    writeropts::HDR_SIGNATURE,
};

/// Common low-level reader used by generated schema readers.
pub struct BaseReader<R: ByteAndBlockReader> {
    source: Option<R>,
    /// Fixed header parsed from stream.
    pub fixed_header: FixedHeader,
    /// Variable header parsed from stream.
    pub var_header: VarHeader,
    /// Optional peer schema from variable header.
    pub schema: Option<schema::WireSchema>,
    /// Per-frame column buffers.
    pub read_bufs: ReadBufs,
    /// Frame decoder.
    pub frame_decoder: FrameDecoder<R>,
    /// Record count in currently loaded frame.
    pub frame_record_count: u64,
    /// Total read record count.
    pub record_count: u64,
}

impl<R: ByteAndBlockReader + Clone> Default for BaseReader<R> {
    fn default() -> Self {
        Self {
            source: None,
            fixed_header: FixedHeader { compression: Compression::None },
            var_header: VarHeader::default(),
            schema: None,
            read_bufs: ReadBufs::default(),
            frame_decoder: FrameDecoder::default(),
            frame_record_count: 0,
            record_count: 0,
        }
    }
}

impl<R: ByteAndBlockReader + Clone> BaseReader<R> {
    /// Initializes reader and frame decoder from source.
    pub fn init(&mut self, source: R) -> Result<()> {
        self.source = Some(source.clone());
        self.read_fixed_header()?;
        self.frame_decoder.init(source, self.fixed_header.compression)?;
        Ok(())
    }

    /// Reads and validates fixed header.
    pub fn read_fixed_header(&mut self) -> Result<()> {
        let src = self.source.as_mut().expect("source not set");
        let mut sig = vec![0u8; HDR_SIGNATURE.len()];
        read_full(src, &mut sig)?;
        if sig != HDR_SIGNATURE.as_bytes() {
            return Err(ERR_INVALID_HEADER_SIGNATURE);
        }

        let content_size = read_uvarint_from_reader(src)?;
        if content_size < 2 || content_size > FIXED_HDR_CONTENT_SIZE_LIMIT {
            return Err(ERR_INVALID_HEADER);
        }
        let mut hdr_content = vec![0u8; content_size as usize];
        read_full(src, &mut hdr_content)?;

        let version = hdr_content[0] & HDR_FORMAT_VERSION_MASK;
        if version != HDR_FORMAT_VERSION {
            return Err(ERR_INVALID_FORMAT_VERSION);
        }

        let flags = hdr_content[1];
        self.fixed_header.compression = Compression::try_from(flags & HDR_FLAGS_COMPRESSION_METHOD)
            .map_err(|_| ERR_INVALID_COMPRESSION)?;
        Ok(())
    }

    /// Reads variable header frame and validates optional schema compatibility.
    pub fn read_var_header(&mut self, own_schema: &schema::WireSchema) -> Result<()> {
        self.frame_decoder.next()?;
        let hdr_size = self.frame_decoder.remaining_size();
        if hdr_size > VAR_HDR_CONTENT_SIZE_LIMIT {
            return Err(ERR_INVALID_VAR_HEADER);
        }

        let mut hdr_bytes = vec![0u8; hdr_size as usize];
        let n = self.frame_decoder.read(&mut hdr_bytes)?;
        hdr_bytes.truncate(n);
        self.var_header.deserialize(&hdr_bytes)?;

        if !self.var_header.schema_wire_bytes.is_empty() {
            let mut ws = schema::WireSchema::default();
            ws.deserialize(&self.var_header.schema_wire_bytes)?;
            own_schema
                .compatible(&ws)
                .map_err(|e| crate::errors::StefError::message(format!("schema is not compatible with BaseReader: {e}")))?;
            self.schema = Some(ws);
        }
        Ok(())
    }

    /// Loads next frame and parses frame record count plus column sections.
    pub fn next_frame(&mut self) -> Result<crate::frameflags::FrameFlags> {
        let frame_flags = self.frame_decoder.next()?;
        self.frame_record_count = read_uvarint_from_frame_decoder(&mut self.frame_decoder)?;
        let remaining = self.frame_decoder.remaining_size();
        self.read_bufs.read_from(&mut self.frame_decoder, remaining)?;
        Ok(frame_flags)
    }
}

fn read_uvarint_from_reader(reader: &mut dyn ByteAndBlockReader) -> Result<u64> {
    let mut tmp = [0u8; 10];
    for i in 0..10 {
        tmp[i] = reader.read_byte()?;
        if tmp[i] < 0x80 {
            let (v, _) = decode_uvarint(&tmp[..=i])?;
            return Ok(v);
        }
    }
    Err(crate::errors::StefError::message("varint overflow"))
}

fn read_uvarint_from_frame_decoder<R: ByteAndBlockReader>(decoder: &mut FrameDecoder<R>) -> Result<u64> {
    let mut tmp = [0u8; 10];
    for i in 0..10 {
        tmp[i] = decoder.read_byte()?;
        if tmp[i] < 0x80 {
            let (v, _) = decode_uvarint(&tmp[..=i])?;
            return Ok(v);
        }
    }
    Err(crate::errors::StefError::message("varint overflow"))
}

fn read_full(src: &mut dyn ByteAndBlockReader, dst: &mut [u8]) -> Result<()> {
    let mut ofs = 0;
    while ofs < dst.len() {
        let n = src.read_block(&mut dst[ofs..])?;
        if n == 0 {
            return Err(std::io::Error::from(std::io::ErrorKind::UnexpectedEof).into());
        }
        ofs += n;
    }
    Ok(())
}
