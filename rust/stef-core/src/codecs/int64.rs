use crate::codecs::uint64::{Uint64Decoder, Uint64Encoder};
use crate::recordbuf::{ReadColumnSet, WriteColumnSet};
use crate::SizeLimiter;

/// Signed delta-of-delta encoder.
#[derive(Default)]
pub struct Int64Encoder {
    inner: Uint64Encoder,
}

impl Int64Encoder {
    pub fn init(&mut self, limiter: &mut SizeLimiter, columns: &mut WriteColumnSet) {
        self.inner.init(limiter, columns);
    }

    pub fn reset(&mut self) {
        self.inner.reset();
    }

    pub fn is_equal(&self, val: i64) -> bool {
        self.inner.is_equal(val as u64)
    }

    pub fn encode(&mut self, val: i64) {
        self.inner.encode(val as u64)
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        self.inner.collect_columns(column_set);
    }
}

/// Signed delta-of-delta decoder.
#[derive(Default)]
pub struct Int64Decoder {
    inner: Uint64Decoder,
}

impl Int64Decoder {
    pub fn init(&mut self, columns: &mut ReadColumnSet) {
        self.inner.init(columns);
    }

    pub fn continue_(&mut self) {
        self.inner.continue_();
    }

    pub fn decode(&mut self, dst: &mut i64) -> crate::errors::Result<()> {
        let mut u = 0u64;
        self.inner.decode(&mut u)?;
        *dst = u as i64;
        Ok(())
    }

    pub fn reset(&mut self) {
        self.inner.reset();
    }
}
