package net.stef;

import net.stef.schema.WireSchema;
import java.util.Map;

public class WriterOptions {
    // IncludeDescriptor indicates that the schema descriptor must be written to the file.
    private boolean includeDescriptor;

    // Compression to use for frame content.
    // CompressionNone disables the compression.
    // CompressionZstd uses zstd compression for frame content.
    private Compression compression;

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
    private long maxUncompressedFrameByteSize;

    // When a frame is restarted these flags define additional behavior.
    //
    // RestartDictionaries - the dictionaries will be cleared. All new frames will
    //   start with new dictionaries. Can be used to limit the size of the
    //   dictionaries that the recipients must keep in memory. Note that the
    //   frames always restart dictionaries regardless of MaxTotalDictSize setting.
    //
    // RestartCompression - the compression stream is started anew.
    //   All new frames will start with a new compressor state.
    //   Can be used to make the frames within a file skipable. Each new
    //   frame's compression streams starts with a new state of encoder.
    //   If this flag is unspecified the state of the compression encoder
    //   carries over through the frames, which makes impossible to skip
    //   frames and start decompressing from the next frame.
    //   This flag has effect only if Compression!=CompressionNone.
    //
    // RestartEncoders - the encoder's state will be cleared. All new frames will
    //   start with initial state of encoders.
    //
    // A combination of RestartDictionaries|RestartCompression|RestartEncoders flags
    // ensures that a frame is readable and decodable on its own, without the need
    // to read any preceding frames.
    private int frameRestartFlags;

    // MaxTotalDictSize is the maximum total byte size of all dictionaries.
    //
    // Default is DefaultMaxTotalDictSize.
    //
    // The Writer will compute the total size of all dictionaries it creates
    // during encoding process. When this limit is reached the Writer will
    // reset the dictionaries and will start a new Frame with RestartDictionaries
    // frame flag set.
    // Note that when the limit is reached the dictionaries will be reset regardless
    // of the value in RestartDictionaries bit in FrameRestartFlags.
    //
    // The Writer's total byte size calculation is approximate.
    // It is expected that the real memory usage by dictionaries may somewhat
    // exceed MaxTotalDictSize before the Writer detects that the limit is reached.
    private long maxTotalDictSize;

    // Schema describes the desired wire schema to write the data in.
    // The schema must be compatible with Writer's native schema otherwise
    // an error will be returned when attempting to create the Writer.
    // In null the Writer will write in its native schema.
    private WireSchema schema;

    // UserData is optional user-defined data that will be stored in the header.
    private java.util.Map<String, String> userData;

    private WriterOptions(Builder builder) {
        includeDescriptor = builder.includeDescriptor;
        compression = builder.compression;
        maxUncompressedFrameByteSize = builder.maxUncompressedFrameByteSize;
        frameRestartFlags = builder.frameRestartFlags;
        maxTotalDictSize = builder.maxTotalDictSize;
        userData = builder.userData;
        schema = builder.schema;
        if (schema != null) {
            // If the schema is overridden must include the descriptor so that the readers
            // can decode the data correctly.
            includeDescriptor = true;
        }
    }

    public static Builder builder() {
        return new Builder();
    }

    // Getters for all fields
    public boolean isIncludeDescriptor() { return includeDescriptor; }
    public Compression getCompression() { return compression; }
    public long getMaxUncompressedFrameByteSize() { return maxUncompressedFrameByteSize; }
    public int getFrameRestartFlags() { return frameRestartFlags; }
    public long getMaxTotalDictSize() { return maxTotalDictSize; }
    public WireSchema getSchema() { return schema; }
    public java.util.Map<String, String> getUserData() { return userData; }

    // DefaultMaxFrameSize is the default maximum size of a frame.
    // 4MiB, less 1KiB to ensure the frame fits in default gRPC message size, which is 4MiB.
    public static final int DefaultMaxFrameSize = (4 << 20) - 1024;

    // DefaultMaxTotalDictSize is the default maximum of MaxTotalDictSize option.
    public static final int DefaultMaxTotalDictSize = 4 << 20;

    public static class Builder {
        private boolean includeDescriptor;
        private Compression compression = Compression.None;
        private long maxUncompressedFrameByteSize = DefaultMaxFrameSize;
        private int frameRestartFlags;
        private long maxTotalDictSize = DefaultMaxTotalDictSize;
        private WireSchema schema;
        private Map<String, String> userData;

        public Builder includeDescriptor(boolean includeDescriptor) {
            this.includeDescriptor = includeDescriptor;
            return this;
        }
        public Builder compression(Compression compression) {
            this.compression = compression;
            return this;
        }
        public Builder maxUncompressedFrameByteSize(long size) {
            this.maxUncompressedFrameByteSize = size;
            return this;
        }
        public Builder frameRestartFlags(int flags) {
            this.frameRestartFlags = flags;
            return this;
        }
        public Builder maxTotalDictSize(long size) {
            this.maxTotalDictSize = size;
            return this;
        }
        public Builder schema(WireSchema schema) {
            this.schema = schema;
            return this;
        }
        public Builder userData(Map<String, String> userData) {
            this.userData = userData;
            return this;
        }
        public WriterOptions build() {
            return new WriterOptions(this);
        }
    }
}
