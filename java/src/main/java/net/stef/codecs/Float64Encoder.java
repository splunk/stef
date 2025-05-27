package net.stef.codecs;

import net.stef.BitsWriter;
import net.stef.BytesWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class Float64Encoder {
    private BitsWriter buf = new BitsWriter();
    private SizeLimiter limiter;
    private double lastVal = 0.0;
    private int leadingBits = 0;
    private int trailingBits = 0;

    public void init(SizeLimiter limiter, WriteColumnSet columns) {
        this.limiter = limiter;
    }

    public boolean isEqual(double val) {
        return lastVal == val;
    }

    public void encode(double val) {
        long xorVal = Double.doubleToRawLongBits(val) ^ Double.doubleToRawLongBits(lastVal);
        lastVal = val;

        if (xorVal == 0) {
            // Same value
            buf.writeBit(0);
            limiter.addFrameBits(1);
            return;
        }

        int oldBitLen = buf.bitCount();
        int leading = Long.numberOfLeadingZeros(xorVal);
        if (leading >= 32) {
            leading = 31;
        }
        int trailing = Long.numberOfTrailingZeros(xorVal);

        int prevLeading = leadingBits;
        int prevTrailing = trailingBits;
        int sigbits = 64 - leading - trailing;

        if (leadingBits != -1 && leading >= leadingBits && trailing >= trailingBits) {
            // Fits in previous [leading..trailing] range.
            if (53-prevLeading-prevTrailing < sigbits) {
                // Current scheme is smaller than trying reset the range. Use the current scheme.
                buf.writeBits(0b10, 2);
                buf.writeBits(xorVal>>>prevTrailing, 64-prevLeading-prevTrailing);
                limiter.addFrameBits(buf.bitCount() - oldBitLen);
                return;
            }
        }

        leadingBits = leading;
        trailingBits = trailing;
        if (sigbits == 0) {
            throw new RuntimeException("unexpected");
        }

        long bitsVal = 0b11;
        bitsVal = (bitsVal << 5) | leading;
        bitsVal = (bitsVal << 6) | (sigbits - 1);

        buf.writeBits(bitsVal, 13);

        buf.writeBits(xorVal>>>trailing, sigbits);

        limiter.addFrameBits(buf.bitCount() - oldBitLen);
    }

    public void collectColumns(WriteColumnSet columnSet) {
        columnSet.setBits(buf);
    }

    public void  writeTo(BytesWriter dest) {
        buf.close();
        dest.writeBytes(buf.toBytes());
    }

    public void reset() {
        lastVal = 0.0;
        leadingBits = 0;
        trailingBits = 0;
    }
}