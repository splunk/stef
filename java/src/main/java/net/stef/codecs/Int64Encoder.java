package net.stef.codecs;

import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

import java.io.IOException;

public class Int64Encoder {
    private BytesWriter buf = new BytesWriter(0);
    private SizeLimiter limiter;
    private long lastVal = 0;
    private long lastDelta = 0;

    public void init(SizeLimiter limiter) {
        this.limiter = limiter;
    }

    public void reset() {
        lastVal = 0;
        lastDelta = 0;
        buf.reset();
    }

    public boolean isEqual(long val) {
        return lastVal == val;
    }

    public void encode(long val) throws IOException {
        long delta = val - lastVal;
        lastVal = val;

        long deltaOfDelta = delta - lastDelta;
        lastDelta = delta;

        int oldLen = buf.size();
        buf.writeVarint(deltaOfDelta);
        int newLen = buf.size();

        limiter.addFrameBytes(newLen - oldLen);
    }

    public void collectColumns( WriteColumnSet columnSet) {
        columnSet.setBytes(buf);
    }

    public void writeTo(BytesWriter output) throws IOException {
        output.writeBytes(buf.toBytes());
    }
}