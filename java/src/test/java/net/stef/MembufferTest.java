package net.stef;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

public class MembufferTest {

    @Test
    public void testMembufReadVaruintExp() {
        BytesWriter bw = new BytesWriter(1000);
        long val = 1;
        for (int j = 0; j < 63; j++) {
            bw.writeUvarint(val);
            val *= 2;
        }

        BytesReader br = new BytesReader();
        br.reset(bw.toBytes());

        for (int i = 0; i < 1000; i++) {
            br.reset(bw.toBytes());
            long checkVal = 1;
            for (int j = 0; j < 63; j++) {
                try {
                    long readVal = br.readUvarint();
                    assertEquals(checkVal, readVal);
                    checkVal *= 2;
                } catch (Exception e) {
                    fail("Unexpected exception: " + e.getMessage());
                }
            }
        }
    }

    @Test
    public void testMembufWriteVaruintSizes() {
        for (int size = 1; size <= 9; size++) {
            long val = (1L << (size * 7)) - 1;
            BytesWriter bw = new BytesWriter(1000);

            for (int i = 0; i < 1000; i++) {
                bw.reset();
                for (int j = 0; j < 1000; j++) {
                    bw.writeUvarint(val);
                }
            }
        }
    }
}