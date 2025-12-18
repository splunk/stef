package pkg

const MultimapElemCountLimit = 1024

const FixedHdrContentSizeLimit = 1 << 20
const VarHdrContentSizeLimit = 1 << 20

// RecordAllocLimit defines the maximum number of bytes that can be allocated
// during the reading of a single record.
const RecordAllocLimit = 1 << 25 // 32 MiB

// FrameSizeLimit is the maximum allowed uncompressed or compressed size for a frame.
const FrameSizeLimit = 1 << 26
