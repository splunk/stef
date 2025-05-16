package net.stef.pkg;

import java.io.EOFException;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;

public class BytesReader {
    private byte[] buf;
    private int byteIndex;
    private int capacity;

    public void reset(ByteBuffer buf) {
        this.buf = buf.array();
        this.byteIndex = buf.arrayOffset();
        this.capacity = buf.capacity();
    }

    public byte readByte() throws EOFException {
        if (byteIndex >= capacity) {
            throw new EOFException();
        }
        return buf[byteIndex++];
    }

    public long readUvarint() throws EOFException {
        long value = 0;
        int shift = 0;
        while (true) {
            if (byteIndex >= capacity) {
                throw new EOFException();
            }
            byte b = buf[byteIndex++];
            value |= (long) (b & 0x7F) << shift;
            if ((b & 0x80) == 0) {
                break;
            }
            shift += 7;
        }
        return value;
    }

    public long readVarint() throws EOFException {
        long x = readUvarint();
        return (x >>> 1) ^ -(x & 1);
    }

    public String readStringBytes(int byteSize) throws EOFException {
        if (capacity - byteIndex < byteSize) {
            throw new EOFException();
        }
        String str = new String(buf, byteIndex, byteSize, StandardCharsets.UTF_8);
        byteIndex += byteSize;
        return str;
    }

    public ByteBuffer readBytesMapped(int byteSize) throws EOFException {
        if (capacity - byteIndex < byteSize) {
            throw new EOFException();
        }
        return ByteBuffer.wrap(buf, byteIndex, byteSize);
    }

    public void mapBytesFromMemBuf(BytesReader src, int byteSize) throws EOFException {
        if (src.byteIndex + byteSize > src.capacity) {
            throw new EOFException();
        }
        this.buf = src.buf;
        this.byteIndex = src.byteIndex;
        this.capacity = byteSize;
    }
}