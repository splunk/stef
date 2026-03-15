use crate::{membuffer::{BytesReader, BytesWriter}, recordbuf::{ReadColumnSet, ReadableColumn, WriteColumnSet}, SizeLimiter};

/// Delta-of-delta encoder for unsigned integers.
#[derive(Default)]
pub struct Uint64Encoder {
    buf: BytesWriter,
    limiter: Option<*mut SizeLimiter>,
    last_val: u64,
    last_delta: u64,
}

impl Uint64Encoder {
    pub fn init(&mut self, limiter: &mut SizeLimiter, _columns: &mut WriteColumnSet) {
        self.limiter = Some(limiter);
    }

    pub fn reset(&mut self) {
        self.last_val = 0;
        self.last_delta = 0;
    }

    pub fn is_equal(&self, val: u64) -> bool {
        self.last_val == val
    }

    pub fn encode(&mut self, val: u64) {
        let delta = val.wrapping_sub(self.last_val);
        self.last_val = val;
        let delta_of_delta = delta.wrapping_sub(self.last_delta) as i64;
        self.last_delta = delta;

        let old_len = self.buf.bytes().len();
        self.buf.write_varint(delta_of_delta);
        let new_len = self.buf.bytes().len();
        if let Some(ptr) = self.limiter {
            unsafe { (*ptr).add_frame_bytes(new_len - old_len) };
        }
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        column_set.set_bytes(&mut self.buf);
    }
}

/// Delta-of-delta decoder for unsigned integers.
#[derive(Default)]
pub struct Uint64Decoder {
    buf: BytesReader,
    column: Option<*mut ReadableColumn>,
    last_val: u64,
    last_delta: u64,
}

impl Uint64Decoder {
    pub fn init(&mut self, columns: &mut ReadColumnSet) {
        self.column = Some(columns.column() as *mut ReadableColumn);
    }

    pub fn continue_(&mut self) {
        let column_ptr = self.column.expect("decoder not initialized");
        // Safe because generated code keeps read column tree alive for decoder lifetime.
        let data = unsafe { (&*column_ptr).data().to_vec() };
        self.buf.reset(data);
    }

    pub fn decode(&mut self, dst: &mut u64) -> crate::errors::Result<()> {
        let delta_of_delta = self.buf.read_varint()?;
        let delta = self.last_delta.wrapping_add(delta_of_delta as u64);
        self.last_delta = delta;
        self.last_val = self.last_val.wrapping_add(delta);
        *dst = self.last_val;
        Ok(())
    }

    pub fn reset(&mut self) {
        self.last_val = 0;
        self.last_delta = 0;
    }
}
