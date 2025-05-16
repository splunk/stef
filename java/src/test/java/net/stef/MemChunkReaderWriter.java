package net.stef;

import java.io.EOFException;
import java.io.IOException;

public class MemChunkReaderWriter extends ByteAndBlockReader implements ChunkWriter {
    private byte[] buf = new byte[0];
    private int readOfs;

    @Override
    public int read() throws IOException {
        if (readOfs >= buf.length) {
            throw new EOFException();
        }
        return buf[readOfs++];
    }

    private static byte[] concat(byte[] b1, byte[] b2) {
        byte[] c = new byte[b1.length + b2.length];
        System.arraycopy(b1, 0, c, 0, b1.length);
        System.arraycopy(b2, 0, c, b1.length, b2.length);
        return c;
    }

    public void writeChunk(byte[] header, byte[] content) throws IOException {
        buf = concat(buf, header);
        buf = concat(buf, content);
    }
}