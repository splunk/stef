package net.stef;

/**
 * Helper utility functions for STEF operations.
 */
public class Helper {

    /**
     * Calculates how many optional fields are within the fields that are kept
     * according to keepFieldCount.
     *
     * This is used to determine how many bits are needed to encode the presence
     * bits of optional fields when schema override is used to keep not all fields.
     *
     * @param optionalsMask has 1 bits set for every optional field in the original schema
     * @param keepFieldCount the number of fields that we want to keep (all fields: optional and regular)
     * @return the number of optional fields within the kept fields
     */
    public static int optionalFieldCount(long optionalsMask, int keepFieldCount) {
        // Bit mask with 1 bit set for every field that we want to keep.
        long keepFieldMask = ~(~0L << keepFieldCount);

        // Zero out bits for fields that we do not want to keep.
        optionalsMask &= keepFieldMask;

        // Count the number of remaining 1 bits in the optionalsMask, that's the number
        // of optional fields within the overall number of kept fields.
        return Long.bitCount(optionalsMask);
    }
}
