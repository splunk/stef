package net.stef;

import java.io.IOException;

public interface ChunkWriter {
    void writeChunk(byte[] header, byte[] content) throws IOException;
}
