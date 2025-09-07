//go:build amd64 && goexperiment.simd

package pkg

import (
	"fmt"
	"math/bits"
	"simd"
	"unsafe"
)

var cpuSupportsAVX2 bool
var cpuSupportsAVX512 bool

// Function pointers set during init() to avoid CPU detection overhead
var readUvar64x4Func func(*BytesReader) ([4]uint64, error)
var readUvar32x4Func func(*BytesReader) ([4]uint32, error)
var writeUvar32x4Func func(*BytesWriter, [4]uint32)
var readUvar64x2Func func(*BytesReader) ([2]uint64, error)
var writeUvar64x2Func func(*BytesWriter, [2]uint64)

func init() {
	// Check for AVX2 support using SIMD package capabilities
	cpuSupportsAVX512 = simd.HasAVX512()
	cpuSupportsAVX2 = simd.HasAVX2()

	fmt.Printf("cpuSupportsAVX2   = %t\n", cpuSupportsAVX2)
	fmt.Printf("cpuSupportsAVX512 = %t\n", cpuSupportsAVX512)

	// Set function pointers based on CPU capabilities
	if cpuSupportsAVX512 {
		readUvar64x4Func = (*BytesReader).readUvar64x4_AVX512
	} else if cpuSupportsAVX2 {
		readUvar64x4Func = (*BytesReader).readUvar64x4_AVX2
	} else {
		readUvar64x4Func = (*BytesReader).readUvar64x4Scalar
	}

	if cpuSupportsAVX2 {
		readUvar32x4Func = (*BytesReader).readUvar32x4_AVX2
		writeUvar32x4Func = (*BytesWriter).writeUvar32x4_AVX2
		readUvar64x2Func = (*BytesReader).readUvar64x2_AVX2
		writeUvar64x2Func = (*BytesWriter).writeUvar64x2_AVX2
	} else {
		readUvar32x4Func = (*BytesReader).readUvar32x4Scalar
		writeUvar32x4Func = (*BytesWriter).writeUvar32x4Scalar
		readUvar64x2Func = (*BytesReader).readUvar64x2Scalar
		writeUvar64x2Func = (*BytesWriter).writeUvar64x2Scalar
	}
}

// ReadUvar64x4 reads 4 variable length integers using the optimal implementation
func (r *BytesReader) ReadUvar64x4() ([4]uint64, error) {
	return readUvar64x4Func(r)
}

// ReadUvar64x2 reads 2 variable length integers using the optimal implementation
func (r *BytesReader) ReadUvar64x2() ([2]uint64, error) {
	return readUvar64x2Func(r)
}

// readUvar64x4_AVX512 handles SIMD-optimized reading with buffer size checks
func (r *BytesReader) readUvar64x4_AVX512() ([4]uint64, error) {
	if len(r.buf[r.byteIndex:]) >= 32 {
		// SIMD-optimized path: we have at least 32 bytes available for SIMD loads

		// Read control byte from r.buf
		controlByte := r.buf[r.byteIndex]
		r.byteIndex++

		// Get total length for this control byte pattern
		totalBytes := uvar64x4ReadLen512[controlByte]

		// Read next 32 bytes from r.buf using 256-bit SIMD
		bufPtr := unsafe.Pointer(&r.buf[r.byteIndex])
		next32Bytes := simd.LoadUint8x32((*[32]uint8)((bufPtr)))

		// Load shuffle indices into 256-bit SIMD register
		shuffleVec := simd.LoadUint8x32(&uvar64x4ReadPermute512[controlByte])

		// Perform 256-bit SIMD permute operation: values = next32Bytes.Permute(shuffleIndices)
		permutedVec := next32Bytes.Permute(shuffleVec)

		// Store the permuted result - this gives us 4 uint64 values perfectly aligned
		var permutedValues [4]uint64
		permutedVec.Store((*[32]uint8)(unsafe.Pointer(&permutedValues)))

		r.byteIndex += totalBytes

		return permutedValues, nil
	}

	return r.readUvar64x4Scalar()
}

// readUvar64x4_AVX2 handles SIMD-optimized reading using two 128-bit operations
func (r *BytesReader) readUvar64x4_AVX2() ([4]uint64, error) {
	if len(r.buf[r.byteIndex:]) >= 32 {
		// SIMD-optimized path: we have at least 32 bytes available for SIMD loads

		// Read control byte from r.buf
		controlByte := r.buf[r.byteIndex]
		r.byteIndex++

		// Get total length for this control byte pattern
		totalBytes := uvar64x4ReadLen512[controlByte]

		// Get length for the first pair
		firstPairLength := uvar64x4ReadLen256Part1[controlByte]

		// Process first pair of values (value0 and value1) using 128-bit SIMD
		bufPtr1 := unsafe.Pointer(&r.buf[r.byteIndex])
		firstChunk := simd.LoadUint8x16((*[16]uint8)(bufPtr1))

		// Load shuffle indices for the first pair
		shuffleVec1 := simd.LoadUint8x16(&uvar64x4ReadPermute256Part1[controlByte])

		// Perform 128-bit SIMD permute operation for first pair
		permutedVec1 := firstChunk.Permute(shuffleVec1)

		// Store the first pair of uint64 values
		var result [4]uint64
		permutedVec1.Store((*[16]uint8)(unsafe.Pointer(&result)))

		// Process second pair of values (value2 and value3) using 128-bit SIMD
		// The second chunk starts after the first pair's data
		bufPtr2 := unsafe.Pointer(&r.buf[r.byteIndex+firstPairLength])
		secondChunk := simd.LoadUint8x16((*[16]uint8)(bufPtr2))

		// Load shuffle indices for the second pair
		shuffleVec2 := simd.LoadUint8x16(&uvar64x4ReadPermute256Part2[controlByte])

		// Perform 128-bit SIMD permute operation for second pair
		permutedVec2 := secondChunk.Permute(shuffleVec2)

		// Store the second pair of uint64 values
		permutedVec2.Store((*[16]uint8)(unsafe.Pointer(&result[2])))

		r.byteIndex += totalBytes

		return result, nil
	}

	return r.readUvar64x4Scalar()
}

// ReadUvar32x4 reads 4 variable length 32-bit integers using the optimal implementation
func (r *BytesReader) ReadUvar32x4() ([4]uint32, error) {
	return readUvar32x4Func(r)
}

func (r *BytesReader) readUvar32x4_AVX2() ([4]uint32, error) {
	if len(r.buf[r.byteIndex:]) >= 16 {
		// Read control byte from r.buf
		controlByte := r.buf[r.byteIndex]
		r.byteIndex++

		// Get total length for this control byte pattern
		totalBytes := uvar32x4ReadLen256[controlByte]

		// Read next 16 bytes from r.buf using 128-bit SIMD
		bufPtr := unsafe.Pointer(&r.buf[r.byteIndex])
		next16Bytes := simd.LoadUint8x16((*[16]uint8)((bufPtr)))

		// Load shuffle indices into 128-bit SIMD register
		shuffleVec := simd.LoadUint8x16(&uvar32x4ReadPermute256[controlByte])

		// Perform 128-bit SIMD permute operation: values = next16Bytes.Permute(shuffleIndices)
		permutedVec := next16Bytes.Permute(shuffleVec)

		// Store the permuted result - this gives us 4 uint32 values perfectly aligned
		var permutedValues [4]uint32
		permutedVec.Store((*[16]uint8)(unsafe.Pointer(&permutedValues)))

		r.byteIndex += totalBytes

		return permutedValues, nil
	}

	return r.readUvar32x4Scalar()
}

func (w *BytesWriter) WriteUvar32x4(values [4]uint32) {
	writeUvar32x4Func(w, values)
}

func (w *BytesWriter) writeUvar32x4_AVX2(values [4]uint32) {
	// Lookup table for converting leading zeros to byte length encoding (0-3 for 0,1,2,4 bytes)
	// For zero values: bits.LeadingZeros32(0|1) = 31, so we map index 31 to code 0
	var lengthLookup = [33]byte{
		3, 3, 3, 3, 3, 3, 3, 3, // 0-7   (values >= 2^24: 4 bytes)
		3, 3, 3, 3, 3, 3, 3, 3, // 8-15  (values >= 2^16: 4 bytes)
		2, 2, 2, 2, 2, 2, 2, 2, // 16-23 (values >= 2^8: 2 bytes)
		1, 1, 1, 1, 1, 1, 1, 1, // 24-31 (values >= 2^0: 1 byte)
		0, // 32 (for zero value)
	}

	// Unrolled length calculation for all 4 values
	val0, val1, val2, val3 := values[0], values[1], values[2], values[3]

	// Get length codes using lookup table (branchless)
	code0 := lengthLookup[bits.LeadingZeros32(val0)]
	code1 := lengthLookup[bits.LeadingZeros32(val1)]
	code2 := lengthLookup[bits.LeadingZeros32(val2)]
	code3 := lengthLookup[bits.LeadingZeros32(val3)]

	// Pack control byte (2 bits per value)
	controlByte := code0 | code1<<2 | code2<<4 | code3<<6

	// Get total size needed using single lookup instead of 4 lookups in byteLengths array
	totalSize := uvar32x4WriteLenByControl256[controlByte]

	// Calculate maximum space needed for PutUint32 operations (worst case: all 4-byte values + control byte)
	const maxSpaceNeeded = 1 + 4 + 4 + 4 + 4

	// Pre-allocate buffer space in one operation with enough room for PutUint32 writes
	startIdx := len(w.buf)
	w.buf = EnsureLen(w.buf, len(w.buf)+maxSpaceNeeded)

	_ = w.buf[startIdx+maxSpaceNeeded-1] // bounds check hint to compiler

	// Write control byte
	w.buf[startIdx] = controlByte

	// Load the packed values into SIMD register
	packedValues := simd.LoadUint8x16((*[16]uint8)(unsafe.Pointer(&values[0])))

	// Load the write permutation pattern for this control byte
	// This determines which bytes to extract from our packed 16-byte values
	shuffleVec := simd.LoadUint8x16(&uvar32x4WritePermute256[controlByte])

	// Use SIMD permute to extract only the needed bytes in the correct order
	extractedBytes := packedValues.Permute(shuffleVec)

	// Store the extracted bytes to the output buffer
	extractedBytes.Store((*[16]uint8)(unsafe.Pointer(&w.buf[startIdx+1])))

	w.buf = w.buf[:startIdx+totalSize]
}

// WriteUvar64x2 writes 2 variable length integers using the optimal implementation
func (w *BytesWriter) WriteUvar64x2(values [2]uint64) {
	writeUvar64x2Func(w, values)
}

// readUvar64x2_AVX2 handles SIMD-optimized reading for 2 uint64 values using 128-bit operations
func (r *BytesReader) readUvar64x2_AVX2() ([2]uint64, error) {
	if len(r.buf[r.byteIndex:]) >= 16 {
		// SIMD-optimized path: we have at least 16 bytes available for SIMD loads

		// Read control byte from r.buf
		controlByte := r.buf[r.byteIndex]
		r.byteIndex++

		// Get total length for this control byte pattern
		totalBytes := uvar64x2ReadLen128[controlByte]

		// Read next 16 bytes from r.buf using 128-bit SIMD
		bufPtr := unsafe.Pointer(&r.buf[r.byteIndex])
		next16Bytes := simd.LoadUint8x16((*[16]uint8)(bufPtr))

		// Load shuffle indices into 128-bit SIMD register
		shuffleVec := simd.LoadUint8x16(&uvar64x2ReadPermute128[controlByte])

		// Perform 128-bit SIMD permute operation: values = next16Bytes.Permute(shuffleIndices)
		permutedVec := next16Bytes.Permute(shuffleVec)

		// Store the permuted result - this gives us 2 uint64 values perfectly aligned
		var permutedValues [2]uint64
		permutedVec.Store((*[16]uint8)(unsafe.Pointer(&permutedValues)))

		r.byteIndex += totalBytes

		return permutedValues, nil
	}

	return r.readUvar64x2Scalar()
}

// writeUvar64x2_AVX2 handles SIMD-optimized writing for 2 uint64 values using 128-bit operations
func (w *BytesWriter) writeUvar64x2_AVX2(values [2]uint64) {
	// Lookup table for converting leading zeros to byte length (0-8 for 0,1,2,3,4,5,6,7,8 bytes)
	var lengthLookup = [65]byte{
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 0-15
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 16-31
		8, 8, 8, 8, 8, 8, 8, 8, 7, 7, 7, 7, 7, 7, 7, 7, // 32-47
		6, 6, 6, 6, 6, 6, 6, 6, 5, 5, 5, 5, 5, 5, 5, 5, // 48-63
		0, // 64 (for zero value)
	}

	// Calculate lengths for both values
	val0, val1 := values[0], values[1]

	// Get length codes using lookup table (branchless)
	code0 := lengthLookup[bits.LeadingZeros64(val0)]
	code1 := lengthLookup[bits.LeadingZeros64(val1)]

	// Pack control byte (4 bits per value)
	controlByte := code0 | code1<<4

	totalSize := 1 + int(code0+code1)

	// Calculate maximum space needed for SIMD operations (worst case: both 8-byte values + control byte)
	const maxSpaceNeeded = 1 + 8 + 8

	// Pre-allocate buffer space
	startIdx := len(w.buf)
	w.buf = EnsureLen(w.buf, len(w.buf)+maxSpaceNeeded)

	_ = w.buf[startIdx+maxSpaceNeeded-1] // bounds check hint to compiler

	// Write control byte
	w.buf[startIdx] = controlByte

	// Load the packed values into SIMD register
	packedValues := simd.LoadUint8x16((*[16]uint8)(unsafe.Pointer(&values[0])))

	// Load the write permutation pattern for this control byte
	// This determines which bytes to extract from our packed 16-byte values
	shuffleVec := simd.LoadUint8x16(&uvar64x2WritePermute128[controlByte])

	// Use SIMD permute to extract only the needed bytes in the correct order
	extractedBytes := packedValues.Permute(shuffleVec)

	// Store the extracted bytes to the output buffer
	extractedBytes.Store((*[16]uint8)(unsafe.Pointer(&w.buf[startIdx+1])))

	// Resize buffer to actual needed size
	w.buf = w.buf[:startIdx+totalSize]
}
