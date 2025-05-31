package net.stef;

import java.util.Arrays;
import java.util.Random;

public class Types {

    public static final byte[] emptyBytes = new byte[0];

    public static int Uint64Compare(long left, long right) {
        return Long.compareUnsigned(left, right);
    }

    public static int Int64Compare(long left, long right) {
        return Long.compare(left, right);
    }

    public static int BoolCompare(boolean left, boolean right) {
        return Boolean.compare(left, right);
    }

    public static int Float64Compare(double left, double right) {
        return Double.compare(left, right);
    }

    public static int StringCompare(StringValue left, StringValue right) {
        return left.compareTo(right);
    }

    public static int BytesCompare(byte[] left, byte[] right) {
        return Arrays.compare(left, right);
    }

    public static boolean Uint64Equal(long left, long right) {
        return left == right;
    }

    public static boolean Int64Equal(long left, long right) {
        return left == right;
    }

    public static boolean BoolEqual(boolean left, boolean right) {
        return left == right;
    }

    public static boolean Float64Equal(double left, double right) {
        return Double.compare(left, right) == 0;
    }

    public static boolean StringEqual(StringValue left, StringValue right) {
        if (left == null || right == null) {
            return left == right; // handles null cases
        }
        return left.equals(right);
    }

    public static boolean BytesEqual(byte[] left, byte[] right) {
        return Arrays.equals(left, right);
    }

    public static long Uint64Random(Random random) {
        return random.nextLong();
    }

    public static long Int64Random(Random random) {
        return random.nextLong();
    }

    public static boolean BoolRandom(Random random) {
        return random.nextBoolean();
    }

    public static double Float64Random(Random random) {
        return random.nextDouble();
    }

    public static StringValue StringRandom(Random random) {
        return new StringValue(String.valueOf(random.nextInt(10)));
    }

    public static byte[] BytesRandom(Random random) {
        byte[] randomBytes = new byte[4]; // Example size, can be adjusted
        random.nextBytes(randomBytes);
        return randomBytes;
    }
}