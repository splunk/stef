package net.stef.codecs;

import net.stef.BytesReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;

import java.io.IOException;
import java.nio.ByteBuffer;

public class Int64DeltaDeltaDecoder {
    private BytesReader buf = new BytesReader();
    private ReadableColumn column;
    private long lastVal = 0;
    private long lastDelta = 0;

    public void init(ReadColumnSet columns) {
        column = columns.getColumn();
    }

    public void reset() {
        lastVal = 0;
        lastDelta = 0;
    }

    public void continueDecoding() {
        buf.reset(ByteBuffer.wrap(column.getData()));
    }

    public long decode() throws IOException {
        long deltaOfDelta = buf.readVarint();
        long delta = lastDelta + deltaOfDelta;
        lastDelta = delta;

        lastVal += delta;
        return lastVal;
    }
}