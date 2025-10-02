package net.stef.codecs;

import net.stef.BytesValue;
import net.stef.SizeLimiter;

import java.util.HashMap;
import java.util.Map;

/**
 * A dictionary for encoding byte arrays to integer IDs.
 * Uses a HashMap to store the mapping from byte arrays to integer IDs.
 * TODO: must supply custom hashCode/equals for byte[] key of HashMap.
 */
public class BytesDictEncoderDict {
    private Map<BytesValue, Integer> m;

    public void init(SizeLimiter limiter) {
        this.m = new HashMap<>();
    }

    public void reset() {
        this.m = new HashMap<>();
    }

    public Integer get(BytesValue key) {
        return m.get(key);
    }

    public void put(BytesValue key, int value) {
        m.put(key, value);
    }

    public int size() {
        return m.size();
    }
}

