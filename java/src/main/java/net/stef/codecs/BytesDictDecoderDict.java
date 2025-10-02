package net.stef.codecs;

import net.stef.BytesValue;

import java.util.ArrayList;
import java.util.List;

/**
 * A dictionary for decoding byte arrays using dictionary encoding.
 * This class maintains a list of byte arrays that represent the dictionary entries.
 */
public class BytesDictDecoderDict {
    private List<BytesValue> dict;

    public BytesDictDecoderDict() {
        this.dict = new ArrayList<>();
    }

    public void init() {}

    public void reset() {
        this.dict.clear();
    }

    public void add(BytesValue value) {
        this.dict.add(value);
    }

    public BytesValue get(int index) {
        return this.dict.get(index);
    }

    public int size() {
        return this.dict.size();
    }
}

