package pkg

import (
	"github.com/splunk/stef/go/pkg/schema"
)

const HdrSignature = "STEF"

type WriterOptions struct {
	// IncludeDescriptor indicates that the schema descriptor must be written to the file.
	IncludeDescriptor bool

	// TimestampMultiplier is the multiplier to be used by readers to convert
	// stored timestamps to nanosecond timestamps. For example a value of 1000
	// means timestamps are stored in microsecond precision.
	//
	// The Writer will NOT perform any timestamp scaling operations. The timestamp
	// values provided via WritePoint() calls will be written as is.
	// TODO: remove from fixed header and allow adding custom fields like this.
	TimestampMultiplier uint64

	// Compression to use for frame content.
	// CompressionNone disables the compression.
	// CompressionZstd uses zstd compression for frame content.
	Compression Compression

	// The maximum size of a frame in bytes (uncompressed size if compression is used).
	// If the content exceeds this size, the frame will be closed and a new frame
	// will be started. If unspecified DefaultMaxFrameSize will be used.
	//
	// Blocks never cross frame boundaries and full blocks are written
	// before a frame is closed, so the frame size may exceed this limit by the
	// size of the largest block.
	//
	// Note that the implementation will do its best to honor this value but it may be occasionally
	// exceeded. It is not guaranteed that the frames are always smaller than this size.
	MaxUncompressedFrameByteSize uint

	// When a frame is restarted these flags define additional behavior.
	//
	// RestartDictionaries - the dictionaries will be cleared. The frames will
	//   start with new dictionaries. Can be used to limit the size of the
	//   dictionaries that the recipients must keep in memory.
	//
	// RestartCompression - the compression stream is started anew.
	//   Can be used to make the frames within a file skipable. Each new
	//   frame's compression streams starts with a new state of encoder.
	//   If this flag is unspecified the state of the compression encoder
	//   carries over through the frames, which makes impossible to skip
	//   frames and start decompressing from the next frame.
	//   This flag has effect only if Compression!=CompressionNone.
	FrameRestartFlags FrameFlags

	// MaxTotalDictSize is the maximum total byte size of all dictionaries.
	//
	// Default is DefaultMaxTotalDictSize.
	//
	// The Writer will compute the total size of all dictionaries it creates
	// during encoding process. When this limit is reached the Writer will
	// reset the dictionaries and will start a new Frame with RestartDictionaries
	// frame flag set.
	//
	// The Writer's total byte size calculation is approximate.
	// It is expected that the real memory usage by dictionaries may somewhat
	// exceed MaxTotalDictSize before the Writer detects that the limit is reached.
	MaxTotalDictSize uint

	// Schema describes the desired wire schema to write the data in.
	// The schema must be compatible with Writer's native schema otherwise
	// an error will be returned when attempting to create the Writer.
	// In nil the Writer will write in its native schema.
	Schema *schema.Schema

	// UserData is optional user-defined data that will be stored in the header.
	UserData map[string]string
}

// DefaultMaxFrameSize is the default maximum size of a frame.
// 4MiB, less 1KiB to ensure the frame fits in default gRPC message size, which is 4MiB.
const DefaultMaxFrameSize = (4 << 20) - 1024

// DefaultMaxTotalDictSize is the default maximum of MaxTotalDictSize option.
const DefaultMaxTotalDictSize = 4 << 20
