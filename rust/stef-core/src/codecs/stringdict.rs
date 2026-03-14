use std::collections::HashMap;

use crate::{errors::ERR_INVALID_REF_NUM, membuffer::{BytesReader, BytesWriter}, recordbuf::{ReadColumnSet, WriteColumnSet}, SizeLimiter};

/// Dictionary state for string encoder.
#[derive(Default)]
pub struct StringDictEncoderDict {
    m: HashMap<String, usize>,
    limiter: Option<*mut SizeLimiter>,
}

impl StringDictEncoderDict {
    pub fn init(&mut self, limiter: &mut SizeLimiter) {
        self.m.clear();
        self.limiter = Some(limiter);
    }

    pub fn reset(&mut self) {
        self.m.clear();
    }
}

/// Dictionary-backed string encoder.
#[derive(Default)]
pub struct StringDictEncoder {
    buf: BytesWriter,
    dict: Option<*mut StringDictEncoderDict>,
    limiter: Option<*mut SizeLimiter>,
}

impl StringDictEncoder {
    pub fn init(&mut self, dict: &mut StringDictEncoderDict, limiter: &mut SizeLimiter, _columns: &mut WriteColumnSet) {
        self.dict = Some(dict);
        self.limiter = Some(limiter);
    }

    pub fn encode(&mut self, val: &str) {
        let old_len = self.buf.bytes().len();
        unsafe {
            let dict = &mut *self.dict.expect("dict not set");
            if let Some(ref_num) = dict.m.get(val) {
                self.buf.write_varint(-(*ref_num as i64) - 1);
                let new_len = self.buf.bytes().len();
                if let Some(ptr) = dict.limiter {
                    (*ptr).add_frame_bytes(new_len - old_len);
                }
                return;
            }

            if val.len() > 1 {
                let ref_num = dict.m.len();
                dict.m.insert(val.to_string(), ref_num);
                if let Some(ptr) = dict.limiter {
                    (*ptr).add_dict_elem_size(val.len() + std::mem::size_of::<String>());
                }
            }
        }

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

/// Dictionary state for string decoder.
#[derive(Default)]
pub struct StringDictDecoderDict {
    dict: Vec<String>,
}

impl StringDictDecoderDict {
    pub fn init(&mut self) {}

    pub fn reset(&mut self) {
        self.dict.clear();
    }
}

/// Dictionary-backed string decoder.
#[derive(Default)]
pub struct StringDictDecoder {
    buf: BytesReader,
    column: Vec<u8>,
    dict: Option<*mut StringDictDecoderDict>,
}

impl StringDictDecoder {
    pub fn init(&mut self, dict: &mut StringDictDecoderDict, columns: &mut ReadColumnSet) {
        self.dict = Some(dict);
        self.column = columns.column().borrow_data();
    }

    pub fn continue_(&mut self) {
        self.buf.reset(self.column.clone());
    }

    pub fn reset(&mut self) {}

    pub fn decode(&mut self, dst: &mut String) -> crate::errors::Result<()> {
        let varint = self.buf.read_varint()?;
        if varint >= 0 {
            let str_len = varint as usize;
            *dst = self.buf.read_string_mapped(str_len)?;
            if str_len > 1 {
                unsafe { (*self.dict.expect("dict not set")).dict.push(dst.clone()) };
            }
            return Ok(());
        }

        let ref_num = (-varint - 1) as usize;
        unsafe {
            let dict = &(*self.dict.expect("dict not set")).dict;
            if ref_num >= dict.len() {
                return Err(ERR_INVALID_REF_NUM);
            }
            *dst = dict[ref_num].clone();
        }
        Ok(())
    }
}
