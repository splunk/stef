package net.stef.pkg;

import java.io.OutputStream;

public class WrapChunkWriter implements ChunkWriter {
    private OutputStream out;

    public WrapChunkWriter(OutputStream out) {
        this.out = out;
    }

    @Override
    public void writeChunk(byte[] header, byte[] content) throws Exception {
        out.write(header);
        out.write(content);
    }
}
