package net.stef.codecs;

import net.stef.BytesValue;
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

    public void encode(BytesValue val) {
        int oldLen = buf.size();
        Integer refNum = dict.get(val);
        if (refNum != null) {
            buf.writeVarint(-refNum - 1);
            int newLen = buf.size();
            limiter.addFrameBytes(newLen - oldLen);
            return;
        }
        byte[] bytes = val.getBytes();
        int bytesLen = bytes==null ? 0 : bytes.length;
        if (bytesLen > 1) {
            int refN = dict.size();
            dict.put(val, refN);
            limiter.addDictElemSize((long) bytesLen + 24);
        }
        buf.writeVarint(bytesLen);
        if (bytes!=null) {
            buf.writeBytes(bytes, 0, bytesLen);
        }
        int newLen = buf.size();
        limiter.addFrameBytes(newLen - oldLen);
    }
}
