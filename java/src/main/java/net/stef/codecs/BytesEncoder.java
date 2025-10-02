package net.stef.codecs;

import net.stef.BytesValue;
import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class BytesEncoder {
    protected final BytesWriter buf = new BytesWriter(0);
    protected SizeLimiter limiter;

    public void init(SizeLimiter limiter, WriteColumnSet columns) {
        this.limiter = limiter;
    }

    public void encode(BytesValue val) {
        int oldLen = buf.size();
        byte[] bytes = val.getBytes();
        int bytesLen = bytes==null ? 0 : bytes.length;
        buf.writeVarint(bytesLen);
        if (bytes!=null) {
            buf.writeBytes(bytes, 0, bytesLen);
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
