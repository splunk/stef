package net.stef.pkg;

import java.io.ByteArrayOutputStream;

public class MemChunkWriter implements ChunkWriter {
    private ByteArrayOutputStream buf = new ByteArrayOutputStream();

    @Override
    public void writeChunk(byte[] header, byte[] content) throws Exception {
        buf.write(header);
        buf.write(content);
    }

    public byte[] getBytes() {
        return buf.toByteArray();
    }
}