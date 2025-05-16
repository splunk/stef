package net.stef;

public class FrameFlags {
    public static final int RESTART_DICTIONARIES = (1 << 0);
    public static final int RESTART_COMPRESSION = (1 << 1);
    public static final int RESTART_CODECS = (1 << 2);

    public static final int FRAME_FLAGS_MASK = RESTART_DICTIONARIES | RESTART_COMPRESSION | RESTART_CODECS;

    public static boolean isValid(int flags) {
        return (flags & ~FRAME_FLAGS_MASK) == 0;
    }
}
