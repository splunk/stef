package net.stef.pkg;

import java.util.ArrayList;
import java.util.List;

public class BitsWriter {
    private List<Byte> stream = new ArrayList<>();
    private long bitsBuf = 0;
    private int bitsBufUsed = 0;

    public void reset() {
        bitsBuf = 0;
        bitsBufUsed = 0;
        stream.clear();
    }

    public void close() {
        while (bitsBufUsed > 0) {
            writeByte((byte) (bitsBuf >>> 56));
            bitsBuf <<= 8;
            bitsBufUsed -= 8;
        }
    }

    public byte[] toByteArray() {
        byte[] result = new byte[stream.size()];
        for (int i = 0; i < stream.size(); i++) {
            result[i] = stream.get(i);
        }
        return result;
    }

    public void writeBit(int bit) {
        bitsBuf = (bitsBuf << 1) | (bit & 1);
        bitsBufUsed++;
        if (bitsBufUsed == 8) {
            writeByte((byte) (bitsBuf >>> 56));
            bitsBufUsed = 0;
        }
    }

    public void writeBits(long value, int nbits) {
        for (int i = nbits - 1; i >= 0; i--) {
            writeBit((int) ((value >>> i) & 1));
        }
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

    private void writeByte(byte b) {
        stream.add(b);
    }
}
