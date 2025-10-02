package net.stef.codecs;

import net.stef.BytesValue;
import net.stef.StringValue;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class StringEncoder {
    private BytesEncoder encoder = new BytesEncoder();

    public StringEncoder() {}

    public void init(SizeLimiter limiter, WriteColumnSet columns) {
        encoder.init(limiter, columns);
    }

    public void encode(StringValue value) {
        if (value == null) {
            encoder.encode(null);
            return;
        }
        encoder.encode(new BytesValue(value.getBytes()));
    }

    public void collectColumns(WriteColumnSet columnSet) {
        encoder.collectColumns(columnSet);
    }

    public void reset() {}
}

