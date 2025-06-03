package net.stef;

import org.junit.jupiter.api.Test;

import java.io.EOFException;
import java.io.IOException;

import static org.junit.jupiter.api.Assertions.*;

public class FrameTest {

    @Test
    public void testLastFrameAndContinue() throws Exception {
        // This test verifies that it is possible to decode until the end of available
        // data, get a correct indication that it is the end of the frame and end
        // of all available data, then once new data becomes available the decoding
        // can continue successfully from the newly added data.
        // The continuation is only possible at the frame boundary.

        // Encode one frame with some data.
        FrameEncoder encoder = new FrameEncoder();
        MemChunkReaderWriter buf = new MemChunkReaderWriter();
        encoder.init(buf, Compression.None);

        byte[] writeStr = "hellohellohellohellohellohellohellohellohellohello".getBytes();
        encoder.write(writeStr);
        encoder.closeFrame();

        // Now decode that frame.
        FrameDecoder decoder = new FrameDecoder();
        decoder.init(buf, Compression.None);
        decoder.next();

        byte[] readStr = new byte[writeStr.length];
        int n = decoder.read(readStr);
        assertEquals(writeStr.length, n);
        assertArrayEquals(writeStr, readStr);

        // Try decoding more, past the end of frame.
        byte[] finalReadStr = readStr;
        Exception exception = assertThrows(IOException.class, () -> decoder.read(finalReadStr));
        assertEquals("End of frame", exception.getMessage());

        // Try decoding the next frame and make sure we get EOF.
        assertThrows(EOFException.class, ()-> decoder.next());

        // Continue adding to the same source byte buffer using encoder.
        encoder.openFrame(0);
        writeStr = "foofoofoofoofoofoofoofoofoofoo".getBytes();
        encoder.write(writeStr);
        encoder.closeFrame();

        // Try reading again. We should get an EndOfFrame error.
        byte[] finalReadStr1 = readStr;
        exception = assertThrows(IOException.class, () -> decoder.read(finalReadStr1));
        assertEquals("End of frame", exception.getMessage());

        // Now try decoding a new frame. This time it should succeed since we added a new frame.
        decoder.next();

        // Read the encoded data.
        readStr = new byte[writeStr.length];
        n = decoder.read(readStr);
        assertEquals(writeStr.length, n);
        assertArrayEquals(writeStr, readStr);

        // Try decoding more, past the end of the second frame.
        byte[] finalReadStr2 = readStr;
        exception = assertThrows(IOException.class, () -> decoder.read(finalReadStr2));
        assertEquals("End of frame", exception.getMessage());
    }
}