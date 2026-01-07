package pkg

const MultimapElemCountLimit = 1024

const FixedHdrContentSizeLimit = 1 << 20
const VarHdrContentSizeLimit = 1 << 20

// FrameSizeLimit is the maximum allowed uncompressed or compressed size for a frame.
const FrameSizeLimit = 1 << 26
