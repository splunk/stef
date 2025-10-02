package net.stef.codecs;

import net.stef.BytesValue;
import net.stef.ReadColumnSet;
import net.stef.StringValue;

import java.io.IOException;

public class StringDecoder {
    private BytesDecoder decoder = new BytesDecoder();

    public void init(ReadColumnSet columns) {
        decoder.init(columns);
    }

    public void continueDecoding() {
        decoder.continueDecoding();
    }

    public void reset() {
        decoder.reset();
    }

    public StringValue decode() throws IOException {
        BytesValue bytes = decoder.decode();
        return new StringValue(bytes.getBytes());
    }
}

