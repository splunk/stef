package net.stef.codecs;

import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class BytesEncoder {
    protected final BytesWriter buf = new BytesWriter(0);
    protected SizeLimiter limiter;

    public void init(SizeLimiter limiter, WriteColumnSet columns) {
        this.limiter = limiter;
    }

    public void encode(byte[] val) {
        int oldLen = buf.size();
        int bytesLen = val==null ? 0 : val.length;
        buf.writeVarint(bytesLen);
        if (val!=null) {
            buf.writeBytes(val, 0, bytesLen);
        }
        int newLen = buf.size();
        limiter.addFrameBytes(newLen - oldLen);
    }

    public void collectColumns(WriteColumnSet columnSet) {
        columnSet.setBytes(buf);
    }

    public void writeTo(BytesWriter outBuf) {
        outBuf.writeBytes(buf.toBytes());
    }

    public void reset() {}
}
