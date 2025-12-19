package net.stef;

import org.junit.jupiter.api.Test;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import static org.junit.jupiter.api.Assertions.*;

class LimitedReaderTest {
    @Test
    void testReadSingleBytesWithinLimit() throws IOException {
        byte[] data = {1, 2, 3, 4, 5};
        LimitedReader reader = new LimitedReader();
        reader.init(new ByteArrayInputStream(data));
        reader.setLimit(3);
        assertEquals(1, reader.read());
        assertEquals(2, reader.read());
        assertEquals(3, reader.read());
        assertEquals(-1, reader.read());
    }

    @Test
    void testReadBufferWithinLimit() throws IOException {
        byte[] data = {10, 20, 30, 40, 50};
        LimitedReader reader = new LimitedReader();
        reader.init(new ByteArrayInputStream(data));
        reader.setLimit(4);
        byte[] buf = new byte[10];
        int n = reader.read(buf, 0, buf.length);
        assertEquals(4, n);
        assertArrayEquals(new byte[]{10, 20, 30, 40, 0, 0, 0, 0, 0, 0}, buf);
        assertEquals(-1, reader.read());
    }

    @Test
    void testReadPastLimit() throws IOException {
        byte[] data = {1, 2, 3};
        LimitedReader reader = new LimitedReader();
        reader.init(new ByteArrayInputStream(data));
        reader.setLimit(2);
        assertEquals(1, reader.read());
        assertEquals(2, reader.read());
        assertEquals(-1, reader.read());
    }

    @Test
    void testZeroLimit() throws IOException {
        byte[] data = {1, 2, 3};
        LimitedReader reader = new LimitedReader();
        reader.init(new ByteArrayInputStream(data));
        reader.setLimit(0);
        assertEquals(-1, reader.read());
    }

    @Test
    void testSetLimitMultipleTimes() throws IOException {
        byte[] data = {1, 2, 3, 4};
        LimitedReader reader = new LimitedReader();
        reader.init(new ByteArrayInputStream(data));
        reader.setLimit(2);
        assertEquals(1, reader.read());
        assertEquals(2, reader.read());
        assertEquals(-1, reader.read());
        reader.setLimit(2);
        assertEquals(3, reader.read());
        assertEquals(4, reader.read());
        assertEquals(-1, reader.read());
    }
}

