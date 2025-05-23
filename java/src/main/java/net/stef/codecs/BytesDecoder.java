package net.stef.codecs;

import net.stef.*;

import java.nio.ByteBuffer;

public class BytesDecoder {
    private BytesReader buf;
    private BytesDecoderDict dict;
    private ReadableColumn column;

    public void init(BytesDecoderDict dict, ReadColumnSet columns) throws Exception {
        this.dict = dict;
        this.column = columns.getColumn();
    }

    public void continueDecoding() {
        buf.reset(ByteBuffer.wrap(column.getData()));
    }

    public void reset() {}

    public byte[] decode() throws Exception {
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
                throw new Exception("Invalid RefNum, out of dictionary range");
            }
            int refNum = (int) (-varint - 1);
            if (refNum >= dict.size()) {
                throw new Exception("Invalid RefNum, out of dictionary range");
            }
            return dict.get(refNum);
        }
    }
}
