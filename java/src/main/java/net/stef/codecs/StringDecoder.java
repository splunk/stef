package net.stef.codecs;

import net.stef.ReadColumnSet;
import net.stef.StringValue;

import java.io.IOException;

public class StringDecoder {
    private BytesDecoder decoder = new BytesDecoder();

    public void init(BytesDecoderDict dict, ReadColumnSet columns) {
        decoder.init(dict,  columns);
    }

    public void continueDecoding() {
        decoder.continueDecoding();
    }

    public void reset() {
        decoder.reset();
    }

    public StringValue decode() throws IOException {
        byte[] bytes = decoder.decode();
        return new StringValue(bytes);
    }
}

