/// Maximum number of elements in a multimap field during decoding.
pub const MULTIMAP_ELEM_COUNT_LIMIT: u64 = 1024;

/// Maximum fixed-header content size.
pub const FIXED_HDR_CONTENT_SIZE_LIMIT: u64 = 1 << 20;

/// Maximum variable-header content size.
pub const VAR_HDR_CONTENT_SIZE_LIMIT: u64 = 1 << 20;

/// Maximum bytes allocated while decoding one record.
pub const RECORD_ALLOC_LIMIT: usize = 1 << 25;

/// Maximum compressed or uncompressed frame size.
pub const FRAME_SIZE_LIMIT: u64 = 1 << 26;
