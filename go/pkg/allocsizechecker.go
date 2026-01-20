package pkg

import (
	"math"
	"math/bits"
)

type AllocSizeChecker struct {
	// allocatedSize tracks the total allocated size in bytes since last resetAllocSize call.
	// Normally used by Allocator. This tracking is independent of the Allocator's calls
	// to Alloc(), which DO NOT result in allocatedSize being updated automatically.
	allocatedSize uint
}

// ResetAllocSize resets the allocated size counter to zero.
func (a *AllocSizeChecker) ResetAllocSize() {
	a.allocatedSize = 0
}

// AddAllocSize adds the size to the allocated size counter.
func (a *AllocSizeChecker) AddAllocSize(size uint) {
	var carry uint
	a.allocatedSize, carry = bits.Add(a.allocatedSize, size, 0)
	if carry != 0 {
		// Overflow, saturate to max value.
		a.allocatedSize = math.MaxUint
	}
}

// IsOverLimit checks if the allocated size exceeds the allocation limit.
func (a *AllocSizeChecker) IsOverLimit() bool {
	return a.allocatedSize > RecordAllocLimit
}

// PrepAllocSize checks if allocating size bytes would exceed the allocation limit.
// It adds the size to the total allocated so far.
// Returns ErrRecordAllocLimitExceeded if the limit is exceeded.
func (a *AllocSizeChecker) PrepAllocSize(size uint) error {
	var carry uint
	a.allocatedSize, carry = bits.Add(a.allocatedSize, size, 0)
	if carry != 0 {
		// Overflow, saturate to max value.
		a.allocatedSize = math.MaxUint
		return ErrRecordAllocLimitExceeded
	}
	if a.IsOverLimit() {
		return ErrRecordAllocLimitExceeded
	}
	return nil
}

// PrepAllocSizeN checks if allocating size*count bytes would exceed the allocation limit.
// It adds the size*count to the total allocated so far.
// Returns ErrRecordAllocLimitExceeded if the limit is exceeded.
func (a *AllocSizeChecker) PrepAllocSizeN(size uint, count uint) error {
	carry, totalSize := bits.Mul(size, count)
	if carry != 0 {
		// Overflow, saturate to max value.
		a.allocatedSize = math.MaxUint
		return ErrRecordAllocLimitExceeded
	}
	a.allocatedSize, carry = bits.Add(a.allocatedSize, totalSize, 0)
	if carry != 0 {
		// Overflow, saturate to max value.
		a.allocatedSize = math.MaxUint
		return ErrRecordAllocLimitExceeded
	}

	if a.IsOverLimit() {
		return ErrRecordAllocLimitExceeded
	}

	return nil
}
