package testutils

import "github.com/klauspost/compress/zstd"

// Create a writer that caches compressors.
// For this operation type we supply a nil Reader.
var zstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))

var zstdDecoder, _ = zstd.NewReader(nil)

func CompressZstd(input []byte) []byte {
	return zstdEncoder.EncodeAll(input, make([]byte, 0, len(input)))
}

func DecompressZstd(input []byte) ([]byte, error) {
	return zstdDecoder.DecodeAll(input, nil)
}
