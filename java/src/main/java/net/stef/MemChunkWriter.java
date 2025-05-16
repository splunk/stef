package net.stef;

import java.io.ByteArrayOutputStream;
import java.io.IOException;

public class MemChunkWriter implements ChunkWriter {
    private ByteArrayOutputStream buf = new ByteArrayOutputStream();

    @Override
    public void writeChunk(byte[] header, byte[] content) throws IOException {
        buf.write(header);
        buf.write(content);
    }

    public byte[] getBytes() {
        return buf.toByteArray();
    }
}