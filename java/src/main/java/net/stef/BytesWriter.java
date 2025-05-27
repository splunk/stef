package net.stef;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Arrays;

public class BytesWriter {
    private byte[] buf;
    private int byteIndex;

    public BytesWriter(int capacity) {
        this.buf = new byte[capacity];
        this.byteIndex = 0;
    }

    public void writeByte(byte b) {
        ensureCapacity(1);
        buf[byteIndex++] = b;
    }

    public void writeBytes(byte[] bytes, int offset, int length) {
        ensureCapacity(length);
        System.arraycopy(bytes, offset, buf, byteIndex, length);
        byteIndex += length;
    }

    public void writeBytes(ByteBuffer bytes) {
        writeBytes(bytes.array(), bytes.arrayOffset(), bytes.capacity());
    }

    public void writeUvarint(long value) {
        while ((value & ~0x7FL) != 0) {
            writeByte((byte) ((value & 0x7F) | 0x80));
            value >>>= 7;
        }
        writeByte((byte) value);
    }

    public void writeVarint(long x) {
        writeUvarint((x << 1) ^ (x >> 63));
    }

    public void reset() {
        byteIndex = 0;
    }

    public void resetAndReserve(int len) {
        if (buf.length < len) {
            buf = Arrays.copyOf(buf, len + 8);
        }
        byteIndex = len;
    }

    public ByteBuffer toBytes() {
        return ByteBuffer.wrap(buf, 0, byteIndex).order(ByteOrder.BIG_ENDIAN);
    }

    public byte[] toBytesCopy() {
        byte[] copy = new byte[byteIndex];
        System.arraycopy(buf, 0, copy, 0, byteIndex);
        return copy;
    }

    private void ensureCapacity(int additionalBytes) {
        if (byteIndex + additionalBytes > buf.length) {
            buf = Arrays.copyOf(buf, Math.max(buf.length * 2, byteIndex + additionalBytes));
        }
    }

    public int size() {
        return byteIndex;
    }
}