use std::collections::HashMap;

use crate::{errors::Result, membuffer::decode_uvarint};

const MAX_SCHEMA_WIRE_BYTES: usize = 1024 * 1024;
const MAX_USER_DATA_VALUES: usize = 1024;
const MAX_STRING_LEN: usize = 256;

/// Fixed stream header.
#[derive(Debug, Clone, Copy, Default)]
pub struct FixedHeader {
    /// Compression used for frame payloads.
    pub compression: crate::compression::Compression,
}

/// Variable stream header.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct VarHeader {
    /// Optional serialized wire schema.
    pub schema_wire_bytes: Vec<u8>,
    /// Optional user key/value pairs.
    pub user_data: HashMap<String, String>,
}

impl VarHeader {
    /// Serializes var header in the same binary layout as Go implementation.
    pub fn serialize(&self, dst: &mut Vec<u8>) -> Result<()> {
        write_uvarint(dst, self.schema_wire_bytes.len() as u64);
        dst.extend_from_slice(&self.schema_wire_bytes);
        write_uvarint(dst, self.user_data.len() as u64);
        for (k, v) in &self.user_data {
            write_string(dst, k);
            write_string(dst, v);
        }
        Ok(())
    }

    /// Deserializes var header from bytes.
    pub fn deserialize(&mut self, src: &[u8]) -> Result<()> {
        let mut ofs = 0usize;
        let (slen_u64, n) = decode_uvarint(&src[ofs..])?;
        ofs += n;
        let slen = slen_u64 as usize;
        if slen > MAX_SCHEMA_WIRE_BYTES {
            return Err(crate::errors::StefError::message(format!(
                "schema too large: {slen} > {MAX_SCHEMA_WIRE_BYTES}"
            )));
        }
        if src.len().saturating_sub(ofs) < slen {
            return Err(std::io::Error::from(std::io::ErrorKind::UnexpectedEof).into());
        }
        self.schema_wire_bytes = src[ofs..ofs + slen].to_vec();
        ofs += slen;

        let (count_u64, m) = decode_uvarint(&src[ofs..])?;
        ofs += m;
        let count = count_u64 as usize;
        if count > MAX_USER_DATA_VALUES {
            return Err(crate::errors::StefError::message(format!(
                "too many user data values: {count} > {MAX_USER_DATA_VALUES}"
            )));
        }

        self.user_data.clear();
        for _ in 0..count {
            let (k, used1) = read_string(&src[ofs..])?;
            ofs += used1;
            let (v, used2) = read_string(&src[ofs..])?;
            ofs += used2;
            self.user_data.insert(k, v);
        }
        Ok(())
    }
}

fn write_uvarint(dst: &mut Vec<u8>, mut v: u64) {
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

fn write_string(dst: &mut Vec<u8>, s: &str) {
    write_uvarint(dst, s.len() as u64);
    dst.extend_from_slice(s.as_bytes());
}

fn read_string(src: &[u8]) -> Result<(String, usize)> {
    let (l_u64, n) = decode_uvarint(src)?;
    let l = l_u64 as usize;
    if l > MAX_STRING_LEN {
        return Err(crate::errors::StefError::message("string too long"));
    }
    if src.len().saturating_sub(n) < l {
        return Err(std::io::Error::from(std::io::ErrorKind::UnexpectedEof).into());
    }
    Ok((String::from_utf8_lossy(&src[n..n + l]).to_string(), n + l))
}
