package net.stef;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

// TODO: need to convert fast reading methods from Go.
public class BitsReader {
    private ByteBuffer buf;
    // Position to read next from the buf.
    private int byteIndex;
    private long bitBuf = 0;
    private int availBitCount = 0;
    private boolean isEOF;

    public void reset(ByteBuffer data) {
        if (data.order()!= ByteOrder.BIG_ENDIAN) {
            throw new RuntimeException("Invalid order in ByteBuffer");
        }

        buf = data;
        bitBuf = 0;
        availBitCount = 0;
    }

    public void reset(byte[] data) {
        reset(ByteBuffer.wrap(data));
    }

    // PeekBits must ensure at least nbits bits become available, i.e. b.availBitCount >= nbits
    // on return. If this means going past EOF then zero bits are appended at the end.
    // Maximum allowed value for nbits is 56.
    public long peekBits(int nbits)  {
        if (nbits <= availBitCount) {
            // Fast path. Have enough available bits.
            if (nbits == 0) {
                return 0;
            }
            return bitBuf >>> (64 - nbits);
        }
        // Slow path. Not enough available bits. Refill, then peek.
        return refillAndPeekBits(nbits);
    }

    public long refillAndPeekBits(int nbits)  {
        if (nbits > 56) {
            throw new RuntimeException("at most 56 bits can be peeked");
        }

        // bitBuf has availBitCount filled. Fill bitBuf at least to 56 bits, which is more than nbits.

        if (byteIndex+8 < buf.limit()) {
            // Plenty of room till end of buffer. Read 8 bytes at once.
            bitBuf |= buf.getLong(byteIndex) >>> availBitCount;
            // Advance by full bytes.
            byteIndex += (63 - availBitCount) >> 3;
            // Update number of available bits. [56..63] of available bits are supported.
            availBitCount |= 56;
        } else {
            // Close to end of buffer. Read slowly more carefully.
            refillSlow();
        }

        // Now peek from bitBuf.
        return bitBuf >>> (64 - nbits);
    }

    private long refillSlow() {
        if (byteIndex >= buf.limit()) {
            isEOF = true;
            return 0;
        }

        while (byteIndex < buf.limit() && availBitCount < 56) {
            byte byt = buf.get(byteIndex);
            bitBuf |= Byte.toUnsignedLong(byt) << (64 - availBitCount - 8);
            byteIndex++;
            availBitCount += 8;
        }

        if (byteIndex >= buf.limit()) {
            // Ensure essentially unlimited zero bits are available for consumption
            // past EOF.
            availBitCount = Integer.MAX_VALUE;
        }

        return 0;
    }

    // Consume advances the bit pointer by nbits bits. Consume must
    // be preceded by PeekBits() call with at least the same value of nbits.
    // Maximum allowed value for count is 56.
    public void consume(int nbits ) {
        bitBuf <<= nbits;
        availBitCount -= nbits;
    }


    public long readBits(int nbits) {
        if (nbits <= 56) {
            long val = peekBits(nbits);
            consume(nbits);
            return val;
        }
        return readBitsMoreThan56(nbits);
    }

    private long readBitsMoreThan56(int nbits) {
        long val = peekBits(56);
        int toConsume = availBitCount;
        if (toConsume > 56) {
            toConsume = 56;
        }
        consume(toConsume);
        nbits -= toConsume;
        val = (val << nbits) | peekBits(nbits);
        consume(nbits);
        return val;
    }

    public int readBit() {
        return (int) readBits(1);
    }

    public long readVarintCompact() {
        long ux = readUvarintCompact();
        return (ux >>> 1) ^ -(ux & 1);
    }

    public long readUvarintCompact() {
        long val = peekBits(56);
        int zeros = Long.numberOfLeadingZeros(val);
        long ret = (val >>> BitstreamLookupTables.readShiftByZeros[zeros]) & BitstreamLookupTables.readMaskByZeros[zeros];
        consume(BitstreamLookupTables.readConsumeCountByZeros[zeros]);
        return ret;
    }
}
