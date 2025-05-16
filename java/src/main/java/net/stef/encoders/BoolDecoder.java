package net.stef.encoders;

import net.stef.BitsReader;
import net.stef.ReadableColumn;
import net.stef.ReadColumnSet;

public class BoolDecoder {
    private BitsReader buf = new BitsReader();
    private ReadableColumn column;

    public void init(ReadColumnSet columns) {
        this.column = columns.getColumn();
    }

    public void reset() {
        // No-op
    }

    public void continueDecoding() {
        buf.reset(column.getData());
    }

    public void decode(Boolean[] dst) {
        int bit = buf.readBit();
        dst[0] = bit != 0;
    }
}