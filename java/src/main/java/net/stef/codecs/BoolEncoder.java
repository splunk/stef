package net.stef.codecs;

import net.stef.BitsWriter;
import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class BoolEncoder {
    private BitsWriter buf = new BitsWriter();
    private SizeLimiter limiter;

    public void init(SizeLimiter limiter, WriteColumnSet columns) {
        this.limiter = limiter;
    }

    public void reset() {
        // No-op
    }

    public void encode(boolean val) {
        int v = val ? 1 : 0;
        buf.writeBit(v);
        limiter.addFrameBits(1);
    }

    public void collectColumns(WriteColumnSet columnSet) {
        columnSet.setBits(buf);
    }

    public void writeTo(BytesWriter buf) {
        buf.writeBytes(this.buf.toBytes());
    }
}