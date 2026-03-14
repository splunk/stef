use crate::{errors::ERR_INVALID_REF_NUM, membuffer::{BytesReader, BytesWriter}, recordbuf::{ReadColumnSet, WriteColumnSet}, SizeLimiter};

/// Plain string encoder using length+bytes records.
#[derive(Default)]
pub struct StringEncoder {
    buf: BytesWriter,
    limiter: Option<*mut SizeLimiter>,
}

impl StringEncoder {
    pub fn init(&mut self, limiter: &mut SizeLimiter, _columns: &mut WriteColumnSet) {
        self.limiter = Some(limiter);
    }

    pub fn encode(&mut self, val: &str) {
        let old_len = self.buf.bytes().len();
        self.buf.write_varint(val.len() as i64);
        self.buf.write_string_bytes(val);
        let new_len = self.buf.bytes().len();
        if let Some(ptr) = self.limiter {
            unsafe { (*ptr).add_frame_bytes(new_len - old_len) };
        }
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        column_set.set_bytes(&mut self.buf);
    }

    pub fn reset(&mut self) {}
}

/// Plain string decoder.
#[derive(Default)]
pub struct StringDecoder {
    buf: BytesReader,
    column: Vec<u8>,
}

impl StringDecoder {
    pub fn init(&mut self, columns: &mut ReadColumnSet) {
        self.column = columns.column().borrow_data();
    }

    pub fn continue_(&mut self) {
        self.buf.reset(self.column.clone());
    }

    pub fn reset(&mut self) {}

    pub fn decode(&mut self, dst: &mut String) -> crate::errors::Result<()> {
        let varint = self.buf.read_varint()?;
        if varint >= 0 {
            *dst = self.buf.read_string_mapped(varint as usize)?;
            Ok(())
        } else {
            Err(ERR_INVALID_REF_NUM)
        }
    }
}
