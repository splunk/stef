package net.stef.codecs;

import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

/**
 * Dictionary-based encoder for byte arrays.
 * It encodes byte arrays by referencing previously seen values to save space.
 */
public class BytesDictEncoder extends BytesEncoder {
    private BytesDictEncoderDict dict;

    public void init(BytesDictEncoderDict dict, SizeLimiter limiter, WriteColumnSet columns) {
        this.dict = dict;
        this.limiter = limiter;
    }

    public void encode(byte[] val) {
        int oldLen = buf.size();
        Integer refNum = dict.get(val);
        if (refNum != null) {
            buf.writeVarint(-refNum - 1);
            int newLen = buf.size();
            limiter.addFrameBytes(newLen - oldLen);
            return;
        }
        int bytesLen = val==null ? 0 : val.length;
        if (bytesLen > 1) {
            int refN = dict.size();
            dict.put(val, refN);
            limiter.addDictElemSize((long) bytesLen + 24);
        }
        buf.writeVarint(bytesLen);
        if (val!=null) {
            buf.writeBytes(val, 0, bytesLen);
        }
        int newLen = buf.size();
        limiter.addFrameBytes(newLen - oldLen);
    }
}
