package net.stef.codecs;

import java.util.ArrayList;
import java.util.List;

public class BytesDecoderDict {
    private List<byte[]> dict;

    public BytesDecoderDict() {
        this.dict = new ArrayList<>();
    }

    public void init() {}

    public void reset() {
        this.dict.clear();
    }

    public void add(byte[] value) {
        this.dict.add(value);
    }

    public byte[] get(int index) {
        return this.dict.get(index);
    }

    public int size() {
        return this.dict.size();
    }
}

