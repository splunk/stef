package net.stef;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

public class BitsWriter {
    private byte[] buf;
    private int bufSize;
    private long bitsBuf = 0;
    private int bitsBufUsed = 0;

    public BitsWriter() {
        reset();
    }

    public void reset() {
        bitsBuf = 0;
        bitsBufUsed = 0;
        buf = new byte[8];
        bufSize = 0;
    }

    public void close() {
        int targetLen = bufSize + (bitsBufUsed+7)/8;
        writeUint64(bitsBuf);
        bufSize = targetLen;
    }

    public ByteBuffer toBytes() {
        return ByteBuffer.wrap(buf, 0, bufSize).order(ByteOrder.BIG_ENDIAN);
    }

    public byte[] toBytesCopy() {
        byte[] copy = new byte[bufSize];
        System.arraycopy(buf, 0, copy, 0, bufSize);
        return copy;
    }

    public void writeBit(int bit) {
        if (bitsBufUsed <= 63) {
            bitsBuf |= Integer.toUnsignedLong(bit) << (63 - bitsBufUsed);
            bitsBufUsed++;
            return;
        }
        writeBitsSlow(bit, 1);
    }

    public void writeBits(long val, int nbits) {
        int nbitsComplement = 64 - nbits;
        if (bitsBufUsed <= nbitsComplement) {
            bitsBuf |= val << (nbitsComplement - bitsBufUsed);
            bitsBufUsed += nbits;
            return;
        }
        writeBitsSlow(val, nbits);
    }

    private void writeBitsSlow(long val, int nbits) {
        // Complete bitsBuf to 64 bits.
        int bitsBufFree = 64 - bitsBufUsed;
        if (bitsBufFree>0) {
            bitsBuf |= val >>> (nbits - bitsBufFree);
        }

        // And append 64 bits to stream.
        writeUint64(bitsBuf);

        // Write the rest of bits
        nbits -= bitsBufFree;
        bitsBuf = val << (64 - nbits);
        bitsBufUsed = nbits;
    }

    public void writeVarintCompact(long value) {
        long ux = (value >> 63) ^ (value << 1);
        writeUvarintCompact(ux);
    }

    public void writeUvarintCompact(long value) {
        // The format is the following:
        // Prefix Bits:   Followed by big endian bits:
        // 1              Nothing. Encodes value of 0.
        // 01             2 bit value
        // 001            5 bit value
        // 0001           12 bit value
        // 00001          19 bit value
        // 000001         26 bit value
        // 0000001        33 bit value
        // 00000001       48 bit value

        int leadingZeros = Long.numberOfLeadingZeros(value);
        value |= BitstreamLookupTables.writeMaskByZeros[leadingZeros];
        int bitCount = BitstreamLookupTables.writeBitsCountByZeros[leadingZeros];
        writeBits(value, bitCount);
    }

    private void ensureSpace(int len) {
        if (bufSize + len <= buf.length) {
            return;
        }

        int newBufSize = bufSize * 2;
        if (bufSize+len > newBufSize) {
            newBufSize = bufSize+len;
        }
        byte[] newBuf = new byte[newBufSize];
        System.arraycopy(buf, 0, newBuf, 0, buf.length);
        buf = newBuf;
    }

    private void writeByte(byte b) {
        ensureSpace(1);
        buf[bufSize++] = b;
    }

    private void writeUint64(long l) {
        ensureSpace(8);
        buf[bufSize] = (byte)(l>>>56);
        buf[bufSize+1] = (byte)(l>>>48);
        buf[bufSize+2] = (byte)(l>>>40);
        buf[bufSize+3] = (byte)(l>>>32);
        buf[bufSize+4] = (byte)(l>>>24);
        buf[bufSize+5] = (byte)(l>>>16);
        buf[bufSize+6] = (byte)(l>>>8);
        buf[bufSize+7] = (byte)(l);
        bufSize += 8;
    }

    public int bitCount()  {
        return bufSize*8 + bitsBufUsed;
    }
}
