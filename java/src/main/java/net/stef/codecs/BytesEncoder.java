package net.stef.codecs;

import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class BytesEncoder {
    private final BytesWriter buf = new BytesWriter(0);
    private BytesEncoderDict dict;
    private SizeLimiter limiter;

    public void init(BytesEncoderDict dict, SizeLimiter limiter, WriteColumnSet columns) {
        this.dict = dict;
        this.limiter = limiter;
    }

    public void encode(byte[] val) {
        int oldLen = buf.size();
        if (dict != null) {
            Integer refNum = dict.get(val);
            if (refNum != null) {
                buf.writeVarint(-refNum - 1);
                int newLen = buf.size();
                limiter.addFrameBytes(newLen - oldLen);
                return;
            }
        }
        int bytesLen = val==null ? 0 : val.length;
        if (dict != null && bytesLen > 1) {
            int refNum = dict.size();
            dict.put(val, refNum);
            limiter.addDictElemSize((long) bytesLen + 24);
        }
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
