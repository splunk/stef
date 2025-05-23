package net.stef.codecs;

import net.stef.BitsReader;

public class Float64Decoder {
    private BitsReader buf;
    private double lastVal = 0.0;
    private long leadingBits = 0;
    private long trailingBits = 0;

    public void init(BitsReader buf) {
        this.buf = buf;
    }

    public double decode() {
        long hdrBits = buf.peekBits(13);
        if ((hdrBits & Float64Constants.FLOAT64_NON_IDENTICAL_BIT) == 0) {
            buf.consume(1);
            return lastVal;
        } else {
            long leading, trailing, sigbits;
            if ((hdrBits & Float64Constants.FLOAT64_NEW_LEADING_TRAILING_BIT) == 0) {
                buf.consume(2);
                leading = leadingBits;
                trailing = trailingBits;
                sigbits = 64 - leading - trailing;
            } else {
                buf.consume(13);
                leading = (hdrBits & Float64Constants.FLOAT64_LEADING_BIT_MASK) >> Float64Constants.FLOAT64_SIG_BITS_COUNT;
                sigbits = (hdrBits & Float64Constants.FLOAT64_SIG_BIT_MASK) + 1;
                trailing = 64 - leading - sigbits;
                leadingBits = leading;
                trailingBits = trailing;
            }

            long xorVal = buf.readBits((int) sigbits);
            xorVal <<= trailing;
            lastVal = Double.longBitsToDouble(xorVal ^ Double.doubleToRawLongBits(lastVal));
        }
        return lastVal;
    }

    public void reset() {
        lastVal = 0.0;
        leadingBits = 0;
        trailingBits = 0;
    }

    class Float64Constants {
        // Float64NonIdenticalBit indicates that the value is not identical to the previous value.
        public static final int FLOAT64_NON_IDENTICAL_BIT = 0b1000000000000;

        // Float64NewLeadingTrailingBit indicates that encoding uses new leading/trailing bit counts.
        public static final int FLOAT64_NEW_LEADING_TRAILING_BIT = 0b0100000000000;

        // Float64LeadingBitMask contains bits that store the leading bit count.
        public static final int FLOAT64_LEADING_BIT_MASK = 0b0011111000000;

        // Float64SigBitMask contains bits that store the trailing bit count.
        public static final int FLOAT64_SIG_BIT_MASK = 0b0000000111111;

        // Float64SigBitsCount is the number of bits used by Float64SigBitMask, same as the
        // number of bits to shift Float64LeadingBitMask to get its value.
        public static final int FLOAT64_SIG_BITS_COUNT = 6;
    }
}