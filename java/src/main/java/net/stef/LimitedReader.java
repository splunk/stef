package net.stef;

import java.io.IOException;
import java.io.InputStream;

// LimitedReader is an InputStream wrapper that limits
// the number of bytes that can be read from the underlying
// InputStream.
public class LimitedReader extends ByteAndBlockReader {
    private InputStream src;
    private long limit;

    public void init(InputStream src) {
        this.src = src;
    }

    public void setLimit(long limit) {
        this.limit = limit;
    }

    @Override
    public int read() throws IOException {
        if (limit <= 0) {
            return -1;
        }
        limit--;
        return src.read();
    }

    @Override
    public int read(byte[] b, int off, int len) throws IOException {
        if (limit <= 0) {
            return -1;
        }
        int toRead = (int) Math.min(limit, b.length);
        int n = src.read(b, off, toRead);
        limit -= n;
        return n;
    }
}
