package pkg

import "math/bits"

// OptionalFieldCount calculates how many optional fields are within the
// fields that are kept according to keepFieldMask.
//
// This is used to determine how many bits are needed to encode the presence
// bits of optional fields when schema override is used to keep not all fields.
// optionalsMask has 1 bits set for every optional field in the original schema.
// keepFieldCount is the number of fields that we want to keep (all fields: optional and regular).
// Returns the number of optional fields within the kept fields.
func OptionalFieldCount(optionalsMask uint64, keepFieldCount uint) uint {
	// Bit mask with 1 bit set for every field that we want to keep.
	keepFieldMask := ^(^uint64(0) << keepFieldCount)

	// Zero out bits for fields that we do not want to keep.
	optionalsMask &= keepFieldMask

	// Count the number of remaining 1 bits in the optionalsMask, that's the number
	// of optional fields within the overall number of kept fields.
	return uint(bits.OnesCount64(optionalsMask))
}
