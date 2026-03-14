use std::fmt::{Display, Formatter};

/// Error type for STEF encoding/decoding operations.
#[derive(Debug)]
pub enum StefError {
    /// I/O failure from the underlying reader/writer.
    Io(std::io::Error),
    /// Structured decode/validation failure.
    Decode(&'static str),
    /// Dynamic validation failure.
    Message(String),
}

impl StefError {
    /// Creates a static decode error.
    pub const fn decode(msg: &'static str) -> Self {
        Self::Decode(msg)
    }

    /// Creates a dynamic message error.
    pub fn message(msg: impl Into<String>) -> Self {
        Self::Message(msg.into())
    }
}

impl Display for StefError {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Io(e) => write!(f, "{e}"),
            Self::Decode(m) => write!(f, "{m}"),
            Self::Message(m) => write!(f, "{m}"),
        }
    }
}

impl std::error::Error for StefError {}

impl From<std::io::Error> for StefError {
    fn from(value: std::io::Error) -> Self {
        Self::Io(value)
    }
}

/// Common crate result type.
pub type Result<T> = std::result::Result<T, StefError>;

pub const ERR_MULTIMAP: StefError = StefError::Decode("invalid multimap");
pub const ERR_MULTIMAP_COUNT_LIMIT: StefError = StefError::Decode("too many elements in the multimap");
pub const ERR_INVALID_REF_NUM: StefError = StefError::Decode("invalid refNum");
pub const ERR_INVALID_ONEOF_TYPE: StefError = StefError::Decode("invalid oneof type");
pub const ERR_INVALID_HEADER: StefError = StefError::Decode("invalid FixedHeader");
pub const ERR_INVALID_HEADER_SIGNATURE: StefError = StefError::Decode("invalid FixedHeader signature");
pub const ERR_INVALID_FORMAT_VERSION: StefError = StefError::Decode("invalid format version in the FixedHeader");
pub const ERR_INVALID_COMPRESSION: StefError = StefError::Decode("invalid compression method");
pub const ERR_INVALID_VAR_HEADER: StefError = StefError::Decode("invalid VarHeader");
pub const ERR_FRAME_SIZE_LIMIT: StefError = StefError::Decode("frame is too large");
pub const ERR_COLUMN_SIZE_LIMIT_EXCEEDED: StefError = StefError::Decode("column size limit exceeded");
pub const ERR_TOTAL_COLUMN_SIZE_LIMIT_EXCEEDED: StefError = StefError::Decode("total column size limit exceeded");
pub const ERR_RECORD_ALLOC_LIMIT_EXCEEDED: StefError = StefError::Decode("record allocation limit exceeded");
pub const ERR_TOO_MANY_FIELDS_TO_DECODE: StefError = StefError::Decode("too many fields to decode");
pub const ERR_EMPTY_ROOT_STRUCT_DISALLOWED: StefError = StefError::Decode("cannot decode empty root struct");
pub const ERR_DECODE_ERROR: StefError = StefError::Decode("decoding error");
