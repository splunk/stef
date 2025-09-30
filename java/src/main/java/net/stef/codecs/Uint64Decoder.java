package net.stef.codecs;

import java.io.IOException;

public class Uint64Decoder extends Int64Decoder {
    public long decode() throws IOException {
        return buf.readUvarint();
    }
}
