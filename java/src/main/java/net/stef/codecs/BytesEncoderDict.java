package net.stef.codecs;

import java.util.HashMap;
import java.util.Map;

public class BytesEncoderDict {
    private Map<byte[], Integer> m;

    public void init() {
        this.m = new HashMap<>();
    }

    public void reset() {
        this.m = new HashMap<>();
    }

    public Integer get(byte[] key) {
        return m.get(key);
    }

    public void put(byte[] key, int value) {
        m.put(key, value);
    }

    public int size() {
        return m.size();
    }
}

