use crate::{bitstream::{BitsReader, BitsWriter}, recordbuf::{ReadColumnSet, ReadableColumn, WriteColumnSet}, SizeLimiter};

/// Boolean bit-packed encoder.
#[derive(Default)]
pub struct BoolEncoder {
    buf: BitsWriter,
    limiter: Option<*mut SizeLimiter>,
}

impl BoolEncoder {
    pub fn init(&mut self, limiter: &mut SizeLimiter, _columns: &mut WriteColumnSet) {
        self.limiter = Some(limiter);
    }

    pub fn reset(&mut self) {}

    pub fn encode(&mut self, val: bool) {
        self.buf.write_bit(if val { 1 } else { 0 });
        if let Some(ptr) = self.limiter {
            // Safe because encoder uses shared lifecycle with limiter in generated code.
            unsafe { (*ptr).add_frame_bits(1) };
        }
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        column_set.set_bits(&mut self.buf);
    }
}

/// Boolean decoder.
#[derive(Default)]
pub struct BoolDecoder {
    buf: BitsReader,
    column: Option<Vec<u8>>,
}

impl BoolDecoder {
    pub fn init(&mut self, columns: &mut ReadColumnSet) {
        self.column = Some(columns.column().data().to_vec());
    }

    pub fn continue_(&mut self) {
        if let Some(col) = &self.column {
            self.buf.reset(col);
        }
    }

    pub fn decode(&mut self, dst: &mut bool) {
        *dst = self.buf.read_bit() != 0;
    }

    pub fn reset(&mut self) {}
}
