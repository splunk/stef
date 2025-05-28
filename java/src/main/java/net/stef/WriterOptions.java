package net.stef;

import net.stef.schema.WireSchema;

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

    // Getters and setters for all fields
    public boolean isIncludeDescriptor() { return includeDescriptor; }
    public void setIncludeDescriptor(boolean includeDescriptor) { this.includeDescriptor = includeDescriptor; }
    public Compression getCompression() { return compression; }
    public void setCompression(Compression compression) { this.compression = compression; }
    public long getMaxUncompressedFrameByteSize() { return maxUncompressedFrameByteSize; }
    public void setMaxUncompressedFrameByteSize(long maxUncompressedFrameByteSize) { this.maxUncompressedFrameByteSize = maxUncompressedFrameByteSize; }
    public int getFrameRestartFlags() { return frameRestartFlags; }
    public void setFrameRestartFlags(int frameRestartFlags) { this.frameRestartFlags = frameRestartFlags; }
    public long getMaxTotalDictSize() { return maxTotalDictSize; }
    public void setMaxTotalDictSize(long maxTotalDictSize) { this.maxTotalDictSize = maxTotalDictSize; }
    public WireSchema getSchema() { return schema; }
    public void setSchema(WireSchema schema) { this.schema = schema; }
    public java.util.Map<String, String> getUserData() { return userData; }
    public void setUserData(java.util.Map<String, String> userData) { this.userData = userData; }
}
