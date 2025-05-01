package net.stef.pkg;

import java.nio.ByteBuffer;

// TODO: need to convert fast reading methods from Go.
public class BitsReader {
    private ByteBuffer buffer;
    private long bitBuf = 0;
    private int availBitCount = 0;

    public void reset(byte[] data) {
        buffer = ByteBuffer.wrap(data);
        bitBuf = 0;
        availBitCount = 0;
    }

    public long readBits(int nbits) {
        while (availBitCount < nbits) {
            if (!buffer.hasRemaining()) {
                throw new IllegalStateException("EOF reached");
            }
            bitBuf = (bitBuf << 8) | (buffer.get() & 0xFF);
            availBitCount += 8;
        }
        long result = bitBuf >>> (availBitCount - nbits);
        availBitCount -= nbits;
        return result;
    }

    public int readBit() {
        return (int) readBits(1);
    }

    public long readVarintCompact() {
        long ux = readUvarintCompact();
        return (ux >>> 1) ^ -(ux & 1);
    }

    public long readUvarintCompact() {
        long value = 0;
        int shift = 0;
        while (true) {
            int b = readBit();
            value |= (long) b << shift;
            shift++;
            if (b == 0) break;
        }
        return value;
    }
}
