package net.stef.codecs;

import net.stef.BytesReader;
import net.stef.BytesValue;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;

import java.io.IOException;
import java.nio.ByteBuffer;

public class BytesDecoder {
    protected BytesReader buf = new BytesReader();
    protected ReadableColumn column;

    public void init(ReadColumnSet columns) {
        this.column = columns.getColumn();
    }

    public void continueDecoding() {
        buf.reset(ByteBuffer.wrap(column.getData()));
    }

    public void reset() {}

    public BytesValue decode() throws IOException {
        long varint = buf.readVarint();
        if (varint >= 0) {
            int strLen = (int) varint;
            byte[] value = buf.readBytes(strLen);
            return new BytesValue(value);
        } else {
            throw new IOException("Invalid RefNum, out of dictionary range");
        }
    }
}
