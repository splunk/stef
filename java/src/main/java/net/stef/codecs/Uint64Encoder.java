package net.stef.codecs;

import java.io.IOException;

public class Uint64Encoder extends Int64Encoder {
    public void encode(long val) throws IOException {
        int oldLen = buf.size();
        buf.writeUvarint(val);
        int newLen = buf.size();
        limiter.addFrameBytes(newLen - oldLen);
    }
}
