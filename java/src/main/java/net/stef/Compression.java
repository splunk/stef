package net.stef;

public enum Compression {
    NONE(0),
    ZSTD(1);

    public static final int COMPRESSION_MASK = 0b11;

    private final int value;

    Compression(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }

    public static Compression fromValue(int value) {
        for (Compression compression : values()) {
            if (compression.value == value) {
                return compression;
            }
        }
        throw new IllegalArgumentException("Unknown compression value: " + value);
    }
}
