package net.stef;

import java.io.*;
import java.nio.charset.StandardCharsets;

public class Serde {
    private static final int MAX_STRING_LEN = 256;

    public static void writeString(String str, ByteArrayOutputStream dst) throws IOException {
        if (str.length() > MAX_STRING_LEN) {
            throw new IOException("String too long");
        }
        writeUvarint(str.length(), dst);
        dst.write(str.getBytes(StandardCharsets.UTF_8));
    }

    public static String readString(ByteArrayInputStream src) throws IOException {
        long length = readUvarint(src);
        if (length > MAX_STRING_LEN) {
            throw new IOException("String too long");
        }

        byte[] buffer = new byte[(int) length];
        if (src.read(buffer) != length) {
            throw new EOFException("Failed to read string");
        }
        return new String(buffer, StandardCharsets.UTF_8);
    }

    public static void writeUvarint(long value, OutputStream dst) throws IOException {
        while ((value & ~0x7FL) != 0) {
            dst.write((int) (value & 0x7F) | 0x80);
            value >>>= 7;
        }
        dst.write((int) value);
    }

    public static long readUvarint(InputStream src) throws IOException {
        long value = 0;
        int shift = 0;
        int b;
        do {
            b = src.read();
            if (b == -1) {
                throw new EOFException("Unexpected end of stream");
            }
            value |= (long) (b & 0x7F) << shift;
            shift += 7;
        } while ((b & 0x80) != 0);
        return value;
    }
}
