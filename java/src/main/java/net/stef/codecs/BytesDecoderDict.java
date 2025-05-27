package net.stef.codecs;

import net.stef.Bytes;

import java.util.ArrayList;
import java.util.List;

public class BytesDecoderDict {
    private List<Bytes> dict;

    public BytesDecoderDict() {
        this.dict = new ArrayList<>();
    }

    public void init() {}

    public void reset() {
        this.dict.clear();
    }

    public void add(Bytes value) {
        this.dict.add(value);
    }

    public Bytes get(int index) {
        return this.dict.get(index);
    }

    public int size() {
        return this.dict.size();
    }
}

