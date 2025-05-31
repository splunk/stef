package net.stef;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

/**
 * StringValue is a wrapper for UTF-8 encoded strings.
 * <p>
 * It stores the string as a UTF-8 byte array and provides methods to access
 * and modify the value as either a Java String or as raw bytes.
 * </p>
 * <ul>
 *   <li>Use {@link #asString()} to get the value as a Java String.</li>
 *   <li>Use {@link #getBytes()} to get the raw UTF-8 bytes.</li>
 *   <li>Use {@link #byteSize()} to get the length in bytes.</li>
 *   <li>Use {@link #set(byte[])} to update the value from a byte array.</li>
 * </ul>
 */
public class StringValue {
    private byte[] utf8Bytes;

    public StringValue(byte[] utf8Bytes) {
        this.utf8Bytes = utf8Bytes;
    }

    public StringValue(String str) {
        this.utf8Bytes = str.getBytes(StandardCharsets.UTF_8);
    }

    public final static StringValue empty = new StringValue(new byte[0]);

    public String asString() {
        return new String(utf8Bytes, StandardCharsets.UTF_8);
    }

    public byte[] getBytes() {
        return utf8Bytes;
    }

    public int byteSize() {
        return utf8Bytes.length;
    }

    public void set(byte[] value) {
        utf8Bytes = value;
    }

    public boolean equals(StringValue other) {
        return Arrays.equals(this.utf8Bytes, other.utf8Bytes);
    }

    public int compareTo(StringValue right) {
        return Arrays.compare(this.utf8Bytes, right.utf8Bytes);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        return equals((StringValue) o);
    }

    @Override
    public int hashCode() {
        return Arrays.hashCode(utf8Bytes);
    }
}

