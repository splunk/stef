package net.stef.codecs;

import net.stef.StringValue;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;

public class StringValueEncoder {
    private BytesEncoder encoder = new BytesEncoder();

    public void init(BytesEncoderDict dict, SizeLimiter limiter, WriteColumnSet columns) {
        encoder.init(dict, limiter, columns);
    }

    public void encode(StringValue value) {
        encoder.encode(value.getBytes());
    }

    public void collectColumns(WriteColumnSet columnSet) {
        encoder.collectColumns(columnSet);
    }
}

