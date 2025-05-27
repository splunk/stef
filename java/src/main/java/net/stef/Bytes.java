package net.stef;

import java.util.Arrays;

// Bytes is a sequence of immutable bytes.
public class Bytes {
    private final byte[] value;

    public Bytes(byte[] value) {
        this.value = value;
    }

    public byte[] getValue() {
        return value;
    }

    public int compareTo(Bytes right) {
        return Arrays.compare(this.value, right.value);
    }
}
