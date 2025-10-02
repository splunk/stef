package net.stef.codecs;

import net.stef.BytesValue;
import net.stef.ReadColumnSet;
import net.stef.StringValue;

import java.io.IOException;

public class StringDictDecoder {
    private BytesDictDecoder decoder = new BytesDictDecoder();

    public void init(BytesDictDecoderDict dict, ReadColumnSet columns) {
        decoder.init(dict,  columns);
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

