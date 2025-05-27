package net.stef.codecs;

import net.stef.StringValue;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class StringEncoder {
    private BytesEncoder encoder = new BytesEncoder();

    public StringEncoder() {}

    public void init(BytesEncoderDict dict, SizeLimiter limiter, WriteColumnSet columns) {
        encoder.init(dict, limiter, columns);
    }

    public void encode(StringValue value) {
        if (value == null) {
            encoder.encode(null);
            return;
        }
        encoder.encode(value.getBytes());
    }

    public void collectColumns(WriteColumnSet columnSet) {
        encoder.collectColumns(columnSet);
    }

    public void reset() {}
}

