use crate::{codecs::string::{StringDecoder, StringEncoder}, types::Bytes, recordbuf::{ReadColumnSet, WriteColumnSet}, SizeLimiter};

/// Bytes encoder using same layout as string codec.
#[derive(Default)]
pub struct BytesEncoder {
    inner: StringEncoder,
}

impl BytesEncoder {
    pub fn init(&mut self, limiter: &mut SizeLimiter, columns: &mut WriteColumnSet) {
        self.inner.init(limiter, columns);
    }

    pub fn encode(&mut self, val: &[u8]) {
        let s = String::from_utf8_lossy(val).to_string();
        self.inner.encode(&s);
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        self.inner.collect_columns(column_set);
    }

    pub fn reset(&mut self) {
        self.inner.reset();
    }
}

/// Bytes decoder.
#[derive(Default)]
pub struct BytesDecoder {
    inner: StringDecoder,
}

impl BytesDecoder {
    pub fn init(&mut self, columns: &mut ReadColumnSet) {
        self.inner.init(columns);
    }

    pub fn continue_(&mut self) {
        self.inner.continue_();
    }

    pub fn decode(&mut self, dst: &mut Bytes) -> crate::errors::Result<()> {
        let mut s = String::new();
        self.inner.decode(&mut s)?;
        *dst = s.into_bytes();
        Ok(())
    }

    pub fn reset(&mut self) {
        self.inner.reset();
    }
}
