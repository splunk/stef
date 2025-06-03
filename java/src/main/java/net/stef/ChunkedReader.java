package net.stef;

import java.io.IOException;
import java.io.InputStream;

public class ChunkedReader extends ByteAndBlockReader {
    private InputStream src;
    private long limit;
    private NextChunkCallback nextChunk;

    public void init(InputStream src) {
        this.src = src;
    }

    public void setLimit(long limit) {
        this.limit = limit;
    }

    public void setNextChunk(NextChunkCallback nextChunk) {
        this.nextChunk = nextChunk;
    }

    @Override
    public int read() throws IOException {
        if (limit <= 0) {
            nextChunk.next();
        }
        limit--;
        return src.read();
    }

    @Override
    public int read(byte[] b, int off, int len) throws IOException {
        if (limit <= 0) {
            nextChunk.next();
        }
        int toRead = (int) Math.min(limit, b.length);
        int n = src.read(b, off, toRead);
        limit -= n;
        return n;
    }

    @FunctionalInterface
    public interface NextChunkCallback {
        void next() throws IOException;
    }
}
