/// Frame compression mode.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Compression {
    /// Uncompressed frame payload.
    None = 0,
    /// Zstandard-compressed frame payload.
    Zstd = 1,
}

impl Default for Compression {
    fn default() -> Self {
        Self::None
    }
}

impl TryFrom<u8> for Compression {
    type Error = crate::errors::StefError;

    fn try_from(value: u8) -> Result<Self, Self::Error> {
        match value {
            0 => Ok(Self::None),
            1 => Ok(Self::Zstd),
            _ => Err(crate::errors::StefError::decode("invalid compression method")),
        }
    }
}

/// Bit-mask for compression flags in fixed header.
pub const COMPRESSION_MASK: u8 = 0b11;
