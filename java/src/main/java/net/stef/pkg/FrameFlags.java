package net.stef.pkg;

public enum FrameFlags {
    RESTART_DICTIONARIES(1 << 0),
    RESTART_COMPRESSION(1 << 1),
    RESTART_CODECS(1 << 2);

    public static final int FRAME_FLAGS_MASK = RESTART_DICTIONARIES.value | RESTART_COMPRESSION.value | RESTART_CODECS.value;

    private final int value;

    FrameFlags(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }

    public static boolean isValid(int flags) {
        return (flags & ~FRAME_FLAGS_MASK) == 0;
    }
}
