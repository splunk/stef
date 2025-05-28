package net.stef;

import java.io.EOFException;
import java.nio.ByteBuffer;

public class BytesReader {
    private byte[] buf;
    private int byteIndex;
    private int capacity;

    public void reset(byte[] data) {
        reset(ByteBuffer.wrap(data));
    }

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

    public byte[] readBytes(int byteSize) throws EOFException {
        if (capacity - byteIndex < byteSize) {
            throw new EOFException();
        }
        byte[] bytes = new byte[byteSize];
        System.arraycopy(buf, byteIndex, bytes, 0, byteSize);
        byteIndex += byteSize;
        return bytes;
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