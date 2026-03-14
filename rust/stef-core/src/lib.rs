#![cfg_attr(bench_nightly, feature(test))]

//! Core STEF encoding/decoding primitives.
//!
//! This crate is a Rust port of `go/pkg` from the STEF repository, excluding the
//! IDL parser package.

pub mod allocsizechecker;
pub mod basereader;
pub mod benches;
pub mod bitstream;
pub mod bitstream_lookuptables;
pub mod chunkwriter;
pub mod codecs;
pub mod compression;
pub mod consts;
pub mod dictlimiter;
pub mod errors;
pub mod frame;
pub mod frameflags;
pub mod header;
pub mod helpers;
pub mod limits;
pub mod membuffer;
pub mod memreaderwriter;
pub mod readopts;
pub mod recordbuf;
pub mod schema;
pub mod slices;
pub mod types;
pub mod writeropts;

pub use allocsizechecker::AllocSizeChecker;
pub use bitstream::{BitsReader, BitsWriter};
pub use chunkwriter::{ChunkWriter, MemChunkWriter, WrapChunkWriter};
pub use compression::Compression;
pub use dictlimiter::SizeLimiter;
pub use frame::{ByteAndBlockReader, EndOfFrame, FrameDecoder, FrameEncoder};
pub use frameflags::FrameFlags;
pub use header::{FixedHeader, VarHeader};
pub use membuffer::{BytesReader, BytesWriter};
pub use memreaderwriter::MemReaderWriter;
pub use readopts::{ErrEndOfFrame, ReadOptions};
pub use recordbuf::{ReadBufs, ReadColumnSet, ReadableColumn, WriteBufs, WriteColumnSet};
pub use writeropts::{WriterOptions, DEFAULT_MAX_FRAME_SIZE, DEFAULT_MAX_TOTAL_DICT_SIZE, HDR_SIGNATURE};
