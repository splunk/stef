use crate::{codecs::stringdict::{StringDictDecoder, StringDictDecoderDict, StringDictEncoder, StringDictEncoderDict}, recordbuf::{ReadColumnSet, WriteColumnSet}, SizeLimiter, types::Bytes};

/// Dictionary bytes encoder.
#[derive(Default)]
pub struct BytesDictEncoder {
    inner: StringDictEncoder,
}

pub type BytesDictEncoderDict = StringDictEncoderDict;

impl BytesDictEncoder {
    pub fn init(&mut self, dict: &mut BytesDictEncoderDict, limiter: &mut SizeLimiter, columns: &mut WriteColumnSet) {
        self.inner.init(dict, limiter, columns);
    }

    pub fn encode(&mut self, val: &[u8]) {
        self.inner.encode(&String::from_utf8_lossy(val));
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        self.inner.collect_columns(column_set);
    }
}

/// Dictionary bytes decoder.
#[derive(Default)]
pub struct BytesDictDecoder {
    inner: StringDictDecoder,
}

pub type BytesDictDecoderDict = StringDictDecoderDict;

impl BytesDictDecoder {
    pub fn init(&mut self, dict: &mut BytesDictDecoderDict, columns: &mut ReadColumnSet) {
        self.inner.init(dict, columns);
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
}
