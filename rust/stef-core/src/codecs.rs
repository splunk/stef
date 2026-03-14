//! Primitive and dictionary codecs used by generated readers/writers.

#[path = "codecs/bool.rs"]
pub mod bool;
#[path = "codecs/bytes.rs"]
pub mod bytes;
#[path = "codecs/bytesdict.rs"]
pub mod bytesdict;
#[path = "codecs/float64.rs"]
pub mod float64;
#[path = "codecs/int64.rs"]
pub mod int64;
#[path = "codecs/string.rs"]
pub mod string;
#[path = "codecs/stringdict.rs"]
pub mod stringdict;
#[path = "codecs/uint64.rs"]
pub mod uint64;
