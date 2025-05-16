package net.stef;

public class SizeLimiter {
    private long dictByteSize;
    private long dictByteSizeLimit;
    private boolean dictSizeLimitReached;

    private long frameBitSize;
    private long frameBitSizeLimit;
    private boolean frameSizeLimitReached;

    // Initializes the limiter with the specified options
    public void init(WriterOptions opts) {
        this.dictByteSize = 0;
        this.frameBitSize = 0;
        this.dictByteSizeLimit = opts.getMaxTotalDictSize();
        this.frameBitSizeLimit = opts.getMaxUncompressedFrameByteSize() * 8;
    }

    // Adds the size of a dictionary element
    public void addDictElemSize(long elemByteSize) {
        if (dictByteSizeLimit != 0) {
            dictByteSize += elemByteSize;
            if (dictByteSize >= dictByteSizeLimit) {
                dictSizeLimitReached = true;
            }
        }
    }

    // Adds bits to the frame buffer
    public void addFrameBits(long bitCount) {
        if (frameBitSizeLimit != 0) {
            frameBitSize += bitCount;
            if (frameBitSize >= frameBitSizeLimit) {
                frameSizeLimitReached = true;
            }
        }
    }

    // Adds bytes to the frame buffer
    public void addFrameBytes(long byteCount) {
        addFrameBits(byteCount * 8);
    }

    // Checks if the dictionary size limit has been reached
    public boolean isDictLimitReached() {
        return dictSizeLimitReached;
    }

    // Checks if the frame size limit has been reached
    public boolean isFrameLimitReached() {
        return frameSizeLimitReached;
    }

    // Resets the dictionary size and limit indicator
    public void resetDict() {
        dictByteSize = 0;
        dictSizeLimitReached = false;
    }

    // Resets the frame size and limit indicator
    public void resetFrameSize() {
        frameBitSize = 0;
        frameSizeLimitReached = false;
    }
}
