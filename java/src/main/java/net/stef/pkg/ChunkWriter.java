package net.stef.pkg;

public interface ChunkWriter {
    void writeChunk(byte[] header, byte[] content) throws Exception;
}
