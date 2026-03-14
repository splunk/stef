use crate::codecs::uint64::{Uint64Decoder, Uint64Encoder};

/// Signed delta-of-delta encoder.
#[derive(Default)]
pub struct Int64Encoder {
    inner: Uint64Encoder,
}

impl Int64Encoder {
    pub fn is_equal(&self, val: i64) -> bool {
        self.inner.is_equal(val as u64)
    }

    pub fn encode(&mut self, val: i64) {
        self.inner.encode(val as u64)
    }
}

/// Signed delta-of-delta decoder.
#[derive(Default)]
pub struct Int64Decoder {
    inner: Uint64Decoder,
}

impl Int64Decoder {
    pub fn decode(&mut self, dst: &mut i64) -> crate::errors::Result<()> {
        let mut u = 0u64;
        self.inner.decode(&mut u)?;
        *dst = u as i64;
        Ok(())
    }
}
