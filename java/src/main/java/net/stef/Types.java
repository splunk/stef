package net.stef;

import java.util.Random;

public class Types {

    // Bytes is a sequence of immutable bytes.
    public static class Bytes {
        private final String value;

        public Bytes(String value) {
            this.value = value;
        }

        public String getValue() {
            return value;
        }
    }

    public static int Uint64Compare(long left, long right) {
        return Long.compare(left, right);
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

    public static int StringCompare(String left, String right) {
        return left.compareTo(right);
    }

    public static int BytesCompare(Bytes left, Bytes right) {
        return left.getValue().compareTo(right.getValue());
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

    public static boolean StringEqual(String left, String right) {
        return left.equals(right);
    }

    public static boolean BytesEqual(Bytes left, Bytes right) {
        return left.getValue().equals(right.getValue());
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

    public static String StringRandom(Random random) {
        return String.valueOf(random.nextInt(10));
    }

    public static Bytes BytesRandom(Random random) {
        return new Bytes(StringRandom(random));
    }
}