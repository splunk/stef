package net.stef;

public class FrameFlags {
    public static final int RestartDictionaries = (1 << 0);
    public static final int RestartCompression = (1 << 1);
    public static final int RestartCodecs = (1 << 2);

    public static final int FRAME_FLAGS_MASK = RestartDictionaries | RestartCompression | RestartCodecs;

    public static boolean isValid(int flags) {
        return (flags & ~FRAME_FLAGS_MASK) == 0;
    }
}
