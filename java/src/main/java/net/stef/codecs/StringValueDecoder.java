package net.stef.codecs;

import net.stef.ReadColumnSet;
import net.stef.StringValue;

public class StringValueDecoder {
    private BytesDecoder decoder = new BytesDecoder();

    public void init(BytesDecoderDict dict, ReadColumnSet columns) throws Exception {
        decoder.init(dict,  columns);
    }

    public void continueDecoding() {
        decoder.continueDecoding();
    }

    public void reset() {
        decoder.reset();
    }

    public StringValue decode() throws Exception {
        byte[] bytes = decoder.decode();
        return new StringValue(bytes);
    }
}

