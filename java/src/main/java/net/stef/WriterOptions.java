package net.stef;

public class WriterOptions {
    private long maxTotalDictSize;
    private long maxUncompressedFrameByteSize;

    public long getMaxTotalDictSize() {
        return maxTotalDictSize;
    }

    public void setMaxTotalDictSize(long maxTotalDictSize) {
        this.maxTotalDictSize = maxTotalDictSize;
    }

    public long getMaxUncompressedFrameByteSize() {
        return maxUncompressedFrameByteSize;
    }

    public void setMaxUncompressedFrameByteSize(long maxUncompressedFrameByteSize) {
        this.maxUncompressedFrameByteSize = maxUncompressedFrameByteSize;
    }
}
