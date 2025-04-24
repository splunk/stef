package net.stef.codecs;

import net.stef.BitsReader;
import net.stef.ReadableColumn;
import net.stef.ReadColumnSet;

public class BoolDecoder {
    private BitsReader buf = new BitsReader();
    private ReadableColumn column;

    public void init(ReadColumnSet columns) {
        this.column = columns.getColumn();
    }

    public void reset() {}

    public void continueDecoding() {
        buf.reset(column.getData());
    }

    public boolean decode() {
        int bit = buf.readBit();
        return bit != 0;
    }
}