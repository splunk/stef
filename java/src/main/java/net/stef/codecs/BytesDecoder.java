package net.stef.codecs;

import net.stef.BytesReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;

import java.io.IOException;
import java.nio.ByteBuffer;

public class BytesDecoder {
    private BytesReader buf = new BytesReader();
    private BytesDecoderDict dict;
    private ReadableColumn column;

    public void init(BytesDecoderDict dict, ReadColumnSet columns) {
        this.dict = dict;
        this.column = columns.getColumn();
    }

    public void continueDecoding() {
        buf.reset(ByteBuffer.wrap(column.getData()));
    }

    public void reset() {}

    public byte[] decode() throws IOException {
        long varint = buf.readVarint();
        if (varint >= 0) {
            int strLen = (int) varint;
            byte[] value = buf.readBytes(strLen);
            if (strLen > 1 && dict != null) {
                dict.add(value);
            }
            return value;
        } else {
            if (dict == null) {
                throw new IOException("Invalid RefNum, out of dictionary range");
            }
            long refNum = -varint - 1;
            if (refNum >= dict.size()) {
                throw new IOException("Invalid RefNum, out of dictionary range");
            }
            return dict.get((int) refNum);
        }
    }
}
