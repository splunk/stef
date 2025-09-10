package net.stef.codecs;

import net.stef.ReadColumnSet;

import java.io.IOException;

/**
 * Decoder for byte arrays with dictionary encoding.
 */
public class BytesDictDecoder extends BytesDecoder {
    private BytesDictDecoderDict dict;

    public void init(BytesDictDecoderDict dict, ReadColumnSet columns) {
        this.dict = dict;
        this.column = columns.getColumn();
    }

    public byte[] decode() throws IOException {
        long varint = buf.readVarint();
        if (varint >= 0) {
            int strLen = (int) varint;
            byte[] value = buf.readBytes(strLen);
            if (strLen > 1) {
                dict.add(value);
            }
            return value;
        } else {
            long refNum = -varint - 1;
            if (refNum >= dict.size()) {
                throw new IOException("Invalid RefNum, out of dictionary range");
            }
            return dict.get((int) refNum);
        }
    }
}
