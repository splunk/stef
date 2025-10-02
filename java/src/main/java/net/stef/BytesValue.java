package net.stef;

import java.util.Arrays;

/**
 * BytesValue is a wrapper for byte array, representing the "bytes" STEF type.
 * <ul>
 *   <li>Use {@link #getBytes()} to get the raw bytes.</li>
 *   <li>Use {@link #byteSize()} to get the length in bytes.</li>
 *   <li>Use {@link #set(byte[])} to update the value from a byte array.</li>
 * </ul>
 */
public class BytesValue {
    private byte[] bytes;

    public BytesValue(byte[] bytes) {
        this.bytes = bytes;
    }

    public final static BytesValue empty = new BytesValue(new byte[0]);

    public byte[] getBytes() {
        return bytes;
    }

    public int byteSize() {
        return bytes.length;
    }

    public void set(byte[] value) {
        bytes = value;
    }

    public boolean equals(BytesValue other) {
        return Arrays.equals(this.bytes, other.bytes);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        return equals((BytesValue) o);
    }

    public int compareTo(BytesValue right) {
        return Arrays.compare(this.bytes, right.bytes);
    }

    @Override
    public int hashCode() {
        return Arrays.hashCode(bytes);
    }
}

