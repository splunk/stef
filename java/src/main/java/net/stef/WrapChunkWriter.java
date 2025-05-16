package net.stef;

import java.io.IOException;
import java.io.OutputStream;

public class WrapChunkWriter implements ChunkWriter {
    private OutputStream out;

    public WrapChunkWriter(OutputStream out) {
        this.out = out;
    }

    @Override
    public void writeChunk(byte[] header, byte[] content) throws IOException {
        out.write(header);
        out.write(content);
    }
}
