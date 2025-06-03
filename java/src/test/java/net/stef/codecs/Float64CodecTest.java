package net.stef.codecs;

import net.stef.BytesWriter;
import net.stef.ReadColumnSet;
import net.stef.SizeLimiter;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class Float64CodecTest {
    @Test
    public void testEncodeDecode() {
        double[] values = {0.0, 1.0, -1.0, 123.456, -789.012, Double.MAX_VALUE, Double.MIN_VALUE,
                42.0, 42.0, 42.0, 42.0,
                Double.NaN, Double.POSITIVE_INFINITY, Double.NEGATIVE_INFINITY, 0.0, -0.0};
        Float64Encoder encoder = new Float64Encoder();
        SizeLimiter sizeLimiter = new SizeLimiter();
        encoder.init(sizeLimiter, null); // SizeLimiter not needed for this test
        for (double v : values) {
            encoder.encode(v);
        }

        BytesWriter buf = new BytesWriter(0);
        encoder.writeTo(buf);

        ReadColumnSet columns = new ReadColumnSet();
        columns.getColumn().setData(buf.toBytesCopy());

        Float64Decoder decoder = new Float64Decoder();
        decoder.init(columns);
        decoder.continueDecoding();

        for (double expected : values) {
            double actual = decoder.decode();
            assertEquals(expected, actual, 0.0);
        }
    }
}

