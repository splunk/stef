package net.stef;

import net.stef.codecs.Float64Decoder;
import net.stef.codecs.Float64Encoder;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class Float64EncoderTest {
    // This sequence was failing before bug fix https://github.com/splunk/stef/issues/268
    private double testVals[] = {
        0.516459,
        0.516459,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.026014,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
    };

    @Test
    void testEncodeDecode() throws IOException {
        Float64Encoder encoder = new Float64Encoder();
        WriteBufs writeBufs = new WriteBufs();
        encoder.init(new SizeLimiter(), writeBufs.columns);
        for (int i = 0; i < testVals.length; i++) {
            encoder.encode(testVals[i]);
        }

        encoder.collectColumns(writeBufs.columns);
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        writeBufs.writeTo(out);

        ReadBufs readBufs = new ReadBufs();
        ByteArrayInputStream in = new ByteArrayInputStream(out.toByteArray());
        readBufs.readFrom(in);

        Float64Decoder decoder = new Float64Decoder();
        decoder.init(readBufs.columns);
        decoder.continueDecoding();
        for (int i = 0; i < testVals.length; i++) {
            double decodedVal = decoder.decode();
            assertEquals(testVals[i], decodedVal, "decoded value at index " + i + " does not match");
        }
    }

    @Test
    void testEncodeDecodeRandomSequence() throws IOException {
        final int N = 100000;
        final long seed = System.nanoTime();;
        java.util.Random rand = new java.util.Random(seed);
        Float64Encoder encoder = new Float64Encoder();
        WriteBufs writeBufs = new WriteBufs();
        encoder.init(new SizeLimiter(), writeBufs.columns);
        for (int i = 0; i < N; i++) {
            encoder.encode(rand.nextDouble()*rand.nextInt());
        }
        encoder.collectColumns(writeBufs.columns);
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        writeBufs.writeTo(out);

        ReadBufs readBufs = new ReadBufs();
        ByteArrayInputStream in = new ByteArrayInputStream(out.toByteArray());
        readBufs.readFrom(in);

        Float64Decoder decoder = new Float64Decoder();
        decoder.init(readBufs.columns);
        decoder.continueDecoding();
        java.util.Random rand2 = new java.util.Random(seed);
        for (int i = 0; i < N; i++) {
            double expected = rand2.nextDouble()*rand2.nextInt();
            double decoded = decoder.decode();
            assertEquals(expected, decoded, "seed "+seed+", index " + i + " does not match");
        }
    }
}
