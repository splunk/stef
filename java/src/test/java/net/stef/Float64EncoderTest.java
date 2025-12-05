package net.stef;

import net.stef.codecs.Float64Decoder;
import net.stef.codecs.Float64Encoder;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class Float64EncoderTest {
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
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.404796,
        0.516459,
        0.613012,
        0.681386,
        0.759367,
        0.909626,
        0.356013,
        0.265753,
        0.891809,
        0.482783,
        0.369160,
        0.779877,
        0.286262,
        0.102260,
        0.937321,
        0.109212,
        0.606182,
        0.656072,
        0.262938,
        0.602772,
        0.820342,
        0.166441,
        0.107999,
        0.151798,
        0.034763,
        0.100905,
        0.673938,
        0.624203,
        0.494612,
        0.043941,
        0.859274,
        0.135444,
        0.363221,
        0.443968,
    };

    @Test
    void testWriteBit() throws IOException {
        Float64Encoder encoder = new Float64Encoder();
        WriteBufs writeBufs = new WriteBufs();
        encoder.init(new SizeLimiter(), writeBufs.columns);
        for (int i = 0; i < testVals.length; i++) {
            encoder.encode(testVals[i]);
        }
        //BytesWriter writer = new BytesWriter(0);
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
            assertEquals(decodedVal, testVals[i], "decoded value at index " + i + " does not match");
        }
    }
}
