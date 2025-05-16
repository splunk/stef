package net.stef;

import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.util.Random;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;

class BitsWriterTest {

    @Test
    void testWriteBit() {
        BitsWriter bw = new BitsWriter();

        for (int i = 0; i < 11; i++) {
            bw.writeBits(1, 1);
        }
        bw.close();
        byte[] expected = {(byte) 0b11111111, (byte) 0b11100000};

        assertArrayEquals(expected, bw.toBytesCopy());
    }

    @Test
    void testIncreasingWriteReadBits() {
        BitsWriter bw = new BitsWriter();

        final long count = 0x1000000;
        for (long i = 1; i <= count; i += 111) {
            long v = i;
            int bitCount = (int) (Math.floor(Math.log(v) / Math.log(2)) + 1);
            bw.writeBits(v, bitCount);
        }
        bw.close();

        BitsReader br = new BitsReader();
        br.reset(bw.toBytes());

        for (long i = 1; i <= count; i += 111) {
            long v = i;
            int bitCount = (int) (Math.floor(Math.log(v) / Math.log(2)) + 1);
            long val = br.readBits(bitCount);
            assertEquals(v, val, "Mismatch at index " + i);
        }
    }

    @Test
    void testRandWriteReadBits() throws IOException {
        BitsWriter bw = new BitsWriter();

        final long count = 0x10000;
        Random random = new Random(0);

        for (long i = 1; i <= count; i++) {
            int shift = random.nextInt(64);
            long v = (random.nextLong()& Long.MAX_VALUE) >>> shift;
            int bitCount = (v == 0) ? 0 : (int) (Math.floor(Math.log(  v) / Math.log(2)) + 1);
            bw.writeBits(v, bitCount);
        }
        bw.close();

        BitsReader br = new BitsReader();
        br.reset(bw.toBytes());

        random = new Random(0);

        for (long i = 1; i <= count; i++) {
            int shift = random.nextInt(64);
            long v = (random.nextLong()& Long.MAX_VALUE) >>> shift;
            int bitCount = (v == 0) ? 0 : (int) (Math.floor(Math.log(v) / Math.log(2)) + 1);
            long val = br.readBits(bitCount);
            assertEquals(v, val, "Mismatch at index " + i);
        }
    }

    @Test
    void testReadUvarintCompact() {
        BitsWriter bw = new BitsWriter();
        for (int i = 0; i < 48; i++) {
            bw.writeUvarintCompact(1L << i);
        }
        bw.close();

        BitsReader br = new BitsReader();
        br.reset(bw.toBytes());

        for (int i = 0; i < 48; i++) {
            long expected = 1L << i;
            long actual = br.readUvarintCompact();
            assertEquals(expected, actual, "Mismatch at index " + i);
        }
    }

    @Test
    void testReadVarintCompact() {
        BitsWriter bw = new BitsWriter();
        for (int i = 0; i < 47; i++) {
            bw.writeVarintCompact(1L << i);
        }
        for (int i = 0; i < 47; i++) {
            bw.writeVarintCompact(-(1L << i));
        }
        bw.close();

        BitsReader br = new BitsReader();
        br.reset(bw.toBytes());

        for (int i = 0; i < 47; i++) {
            long expected = 1L << i;
            long actual = br.readVarintCompact();
            assertEquals(expected, actual, "Mismatch at index " + i);
        }
        for (int i = 0; i < 47; i++) {
            long expected = -(1L << i);
            long actual = br.readVarintCompact();
            assertEquals(expected, actual, "Mismatch at index -" + i);
        }
    }
}
