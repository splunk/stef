use std::io::Write;

use crate::{
    bitstream::{BitsReader, BitsWriter},
    errors::{ERR_COLUMN_SIZE_LIMIT_EXCEEDED, ERR_TOTAL_COLUMN_SIZE_LIMIT_EXCEEDED, Result},
    frame::ByteAndBlockReader,
    membuffer::{decode_uvarint, BytesWriter},
    slices::ensure_len,
};

/// Column node used during encoding.
#[derive(Debug, Default, Clone)]
pub struct WriteColumnSet {
    data: Vec<u8>,
    sub_columns: Vec<WriteColumnSet>,
}

impl WriteColumnSet {
    /// Returns total number of columns including descendants.
    pub fn total_count(&self) -> u32 {
        1 + self.sub_columns.iter().map(WriteColumnSet::total_count).sum::<u32>()
    }

    /// Adds and returns a new sub-column.
    pub fn add_sub_column(&mut self) -> &mut WriteColumnSet {
        self.sub_columns.push(WriteColumnSet::default());
        self.sub_columns.last_mut().expect("pushed")
    }

    /// Moves encoded bits into this column.
    pub fn set_bits(&mut self, b: &mut BitsWriter) {
        b.close();
        self.data = b.bytes().to_vec();
        b.reset();
    }

    /// Moves encoded bytes into this column.
    pub fn set_bytes(&mut self, b: &mut BytesWriter) {
        self.data = b.as_vec();
        b.reset();
    }

    fn write_sizes_to(&self, buf: &mut BitsWriter) {
        buf.write_uvarint_compact(self.data.len() as u64);
        if self.data.is_empty() {
            return;
        }
        for sub in &self.sub_columns {
            sub.write_sizes_to(buf);
        }
    }

    fn write_data_to(&self, buf: &mut dyn Write) -> Result<()> {
        buf.write_all(&self.data)?;
        if self.data.is_empty() {
            return Ok(());
        }
        for sub in &self.sub_columns {
            sub.write_data_to(buf)?;
        }
        Ok(())
    }

    /// Returns sub-column by index.
    pub fn at(&mut self, i: usize) -> &mut WriteColumnSet {
        &mut self.sub_columns[i]
    }

    /// Returns current column payload bytes.
    pub fn data(&self) -> &[u8] {
        &self.data
    }

    /// Replaces current column payload bytes.
    pub fn set_data(&mut self, data: Vec<u8>) {
        self.data = data;
    }
}

/// Top-level write buffers for one record/frame payload.
#[derive(Debug, Default, Clone)]
pub struct WriteBufs {
    /// Root column tree.
    pub columns: WriteColumnSet,
    temp_buf: BitsWriter,
    bytes: Vec<u8>,
}

impl WriteBufs {
    /// Serializes size section and column payloads.
    pub fn write_to(&mut self, buf: &mut dyn Write) -> Result<()> {
        self.temp_buf.reset();
        self.columns.write_sizes_to(&mut self.temp_buf);
        self.temp_buf.close();

        self.bytes.clear();
        append_uvarint(&mut self.bytes, self.temp_buf.bytes().len() as u64);
        buf.write_all(&self.bytes)?;
        buf.write_all(self.temp_buf.bytes())?;
        self.columns.write_data_to(buf)
    }
}

/// Read-only column payload access.
#[derive(Debug, Default, Clone)]
pub struct ReadableColumn {
    data: Vec<u8>,
}

impl ReadableColumn {
    /// Returns a shared view of column bytes.
    pub fn data(&self) -> &[u8] {
        &self.data
    }

    /// Moves column bytes out, leaving column empty.
    pub fn borrow_data(&mut self) -> Vec<u8> {
        std::mem::take(&mut self.data)
    }
}

/// Column node used during decoding.
#[derive(Debug, Default, Clone)]
pub struct ReadColumnSet {
    column: ReadableColumn,
    sub_columns: Vec<ReadColumnSet>,
}

impl ReadColumnSet {
    /// Returns this column payload object.
    pub fn column(&mut self) -> &mut ReadableColumn {
        &mut self.column
    }

    /// Adds and returns new sub-column.
    pub fn add_sub_column(&mut self) -> &mut ReadColumnSet {
        self.sub_columns.push(ReadColumnSet::default());
        self.sub_columns.last_mut().expect("pushed")
    }

    /// Number of sub-columns.
    pub fn sub_column_len(&self) -> usize {
        self.sub_columns.len()
    }

    /// Replaces column payload bytes.
    pub fn set_column_data(&mut self, data: Vec<u8>) {
        self.column.data = data;
    }

    fn read_sizes_from(&mut self, buf: &mut BitsReader, read_limit: &mut u64) -> Result<()> {
        let data_size = buf.read_uvarint_compact();
        if data_size > *read_limit {
            return Err(ERR_COLUMN_SIZE_LIMIT_EXCEEDED);
        }
        *read_limit -= data_size;
        ensure_len(&mut self.column.data, data_size as usize);

        if data_size == 0 {
            for sub in &mut self.sub_columns {
                sub.reset_data();
            }
            return Ok(());
        }
        for sub in &mut self.sub_columns {
            sub.read_sizes_from(buf, read_limit)?;
        }
        Ok(())
    }

    fn read_data_from(&mut self, buf: &mut dyn ByteAndBlockReader) -> Result<()> {
        read_full(buf, &mut self.column.data)?;
        for sub in &mut self.sub_columns {
            sub.read_data_from(buf)?;
        }
        Ok(())
    }

    fn reset_data(&mut self) {
        self.column.data.clear();
        for sub in &mut self.sub_columns {
            sub.reset_data();
        }
    }
}

/// Top-level read buffers for one record/frame payload.
#[derive(Debug, Default, Clone)]
pub struct ReadBufs {
    /// Root column tree.
    pub columns: ReadColumnSet,
    temp_buf: BitsReader,
    temp_buf_bytes: Vec<u8>,
    read_limit: u64,
}

impl ReadBufs {
    /// Loads and parses column sections from frame payload.
    pub fn read_from(&mut self, buf: &mut dyn ByteAndBlockReader, read_limit: u64) -> Result<()> {
        let buf_size = read_uvarint_from_reader(buf)?;
        if buf_size > read_limit {
            return Err(ERR_TOTAL_COLUMN_SIZE_LIMIT_EXCEEDED);
        }

        ensure_len(&mut self.temp_buf_bytes, buf_size as usize);
        read_full(buf, &mut self.temp_buf_bytes)?;
        self.temp_buf.reset(&self.temp_buf_bytes);

        self.read_limit = read_limit - buf_size;
        self.columns.read_sizes_from(&mut self.temp_buf, &mut self.read_limit)?;
        self.columns.read_data_from(buf)
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

fn read_full(reader: &mut dyn ByteAndBlockReader, dst: &mut [u8]) -> Result<()> {
    let mut ofs = 0;
    while ofs < dst.len() {
        let n = reader.read_block(&mut dst[ofs..])?;
        if n == 0 {
            return Err(std::io::Error::from(std::io::ErrorKind::UnexpectedEof).into());
        }
        ofs += n;
    }
    Ok(())
}
