/// Mask for fixed-header format version.
pub const HDR_FORMAT_VERSION_MASK: u8 = 0x0F;

/// Current STEF fixed-header format version.
pub const HDR_FORMAT_VERSION: u8 = 0;

/// Mask selecting compression method bits in fixed-header flags.
pub const HDR_FLAGS_COMPRESSION_METHOD: u8 = 0b0000_0011;
