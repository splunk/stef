package net.stef.codecs;

import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

import java.io.IOException;

public class Int64DeltaEncoder {
    private BytesWriter buf = new BytesWriter(0);
    private SizeLimiter limiter;
    private long lastVal = 0;

    public void init(SizeLimiter limiter, WriteColumnSet columns) {
        this.limiter = limiter;
    }

    public void reset() {
        lastVal = 0;
    }

    public void encode(long val) throws IOException {
        long delta = val - lastVal;
        lastVal = val;

        int oldLen = buf.size();
        buf.writeVarint(delta);
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