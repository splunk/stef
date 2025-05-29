package net.stef;

public class Constants {
    public static final byte[] HdrSignature = new byte[]{'S', 'T', 'E', 'F'};
    public static final byte HdrFormatVersionMask = 0x0F;
    public static final int HdrFormatVersion = 0;
    public static final byte HdrFlagsCompressionMethod = 0b00000011;
}
