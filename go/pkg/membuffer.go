package pkg

import (
	"encoding/binary"
	"io"
	"math/bits"
	"unsafe"
)

type BytesReader struct {
	buf       []byte
	byteIndex int
}

func (r *BytesReader) Reset(buf []byte) {
	r.buf = buf
	r.byteIndex = 0
}

func (r *BytesReader) ReadByte() (byte, error) {
	if r.byteIndex >= len(r.buf) {
		return 0, io.EOF
	}
	b := r.buf[r.byteIndex]
	r.byteIndex++
	return b, nil
}

func (r *BytesReader) ReadUvarint() (value uint64, err error) {
	val, n := binary.Uvarint(r.buf[r.byteIndex:])
	if n <= 0 {
		return 0, io.EOF
	}
	r.byteIndex += n
	return val, nil
}

func (r *BytesReader) ReadVarint() (value int64, err error) {
	x, n := binary.Uvarint(r.buf[r.byteIndex:])
	if n <= 0 {
		return 0, io.EOF
	}
	r.byteIndex += n
	return int64((x >> 1) ^ (-(x & 1))), err
}

func (r *BytesReader) ReadStringBytes(byteSize int) (string, error) {
	if len(r.buf[r.byteIndex:]) < byteSize {
		return "", io.EOF
	}
	str := string(r.buf[r.byteIndex : r.byteIndex+byteSize])
	r.byteIndex += byteSize
	return str, nil
}

func (r *BytesReader) ReadBytesMapped(byteSize int) ([]byte, error) {
	if len(r.buf[r.byteIndex:]) < byteSize {
		return nil, io.EOF
	}

	// Map instead of copying.
	mappedBuf := r.buf[r.byteIndex : r.byteIndex+byteSize]
	r.byteIndex += byteSize

	return mappedBuf, nil
}

func (r *BytesReader) ReadStringMapped(byteSize int) (string, error) {
	if byteSize == 0 {
		return "", nil
	}

	if r.byteIndex+byteSize > len(r.buf) {
		return "", io.EOF
	}

	// Map instead of copying.
	str := unsafe.String(&r.buf[r.byteIndex], byteSize)
	r.byteIndex += byteSize

	return str, nil
}

func (r *BytesReader) MapBytesFromMemBuf(src *BytesReader, byteSize int) error {
	buf, err := src.ReadBytesMapped(byteSize)
	if err != nil {
		return err
	}

	r.buf = buf
	r.byteIndex = 0
	return nil
}

func (r *BytesReader) readUvar64x4Scalar() ([4]uint64, error) {
	// Lookup tables for branchless operation
	var byteLengths = [4]int{1, 2, 4, 8}
	var masks = [4]uint64{0xFF, 0xFFFF, 0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF}

	// Check if we have at least 1 byte for control byte
	if r.byteIndex >= len(r.buf) {
		return [4]uint64{}, io.EOF
	}

	// Read control byte
	controlByte := r.buf[r.byteIndex]
	r.byteIndex++

	// Extract 2-bit length codes for each value (unrolled)
	code0 := controlByte & 0x3        // bits 0-1
	code1 := (controlByte >> 2) & 0x3 // bits 2-3
	code2 := (controlByte >> 4) & 0x3 // bits 4-5
	code3 := (controlByte >> 6) & 0x3 // bits 6-7

	// Get actual byte lengths for each value
	len0 := byteLengths[code0]
	len1 := byteLengths[code1]
	len2 := byteLengths[code2]
	len3 := byteLengths[code3]

	// Calculate total bytes needed
	totalBytes := len0 + len1 + len2 + len3

	// Check if we have enough bytes remaining
	if len(r.buf[r.byteIndex:]) < totalBytes {
		return [4]uint64{}, io.EOF
	}

	// Calculate offsets for each value
	offset0 := r.byteIndex
	offset1 := offset0 + len0
	offset2 := offset1 + len1
	offset3 := offset2 + len2

	// Fast path: check if we have at least 8 bytes available after the last value's offset
	// This allows us to safely use Uint64 reads without bounds checking
	maxOffset := offset3 + 8
	if len(r.buf) >= maxOffset {
		// Fast path: use direct Uint64 reads with masking (no bounds checks needed)
		val0 := binary.LittleEndian.Uint64(r.buf[offset0:]) & masks[code0]
		val1 := binary.LittleEndian.Uint64(r.buf[offset1:]) & masks[code1]
		val2 := binary.LittleEndian.Uint64(r.buf[offset2:]) & masks[code2]
		val3 := binary.LittleEndian.Uint64(r.buf[offset3:]) & masks[code3]

		r.byteIndex += totalBytes
		return [4]uint64{val0, val1, val2, val3}, nil
	}

	// Slow path: carefully read each value with proper bounds checking
	var val0, val1, val2, val3 uint64

	// Value 0 - read exactly len0 bytes
	switch len0 {
	case 1:
		val0 = uint64(r.buf[offset0])
	case 2:
		val0 = uint64(binary.LittleEndian.Uint16(r.buf[offset0 : offset0+2]))
	case 4:
		val0 = uint64(binary.LittleEndian.Uint32(r.buf[offset0 : offset0+4]))
	case 8:
		val0 = binary.LittleEndian.Uint64(r.buf[offset0 : offset0+8])
	}

	// Value 1 - read exactly len1 bytes
	switch len1 {
	case 1:
		val1 = uint64(r.buf[offset1])
	case 2:
		val1 = uint64(binary.LittleEndian.Uint16(r.buf[offset1 : offset1+2]))
	case 4:
		val1 = uint64(binary.LittleEndian.Uint32(r.buf[offset1 : offset1+4]))
	case 8:
		val1 = binary.LittleEndian.Uint64(r.buf[offset1 : offset1+8])
	}

	// Value 2 - read exactly len2 bytes
	switch len2 {
	case 1:
		val2 = uint64(r.buf[offset2])
	case 2:
		val2 = uint64(binary.LittleEndian.Uint16(r.buf[offset2 : offset2+2]))
	case 4:
		val2 = uint64(binary.LittleEndian.Uint32(r.buf[offset2 : offset2+4]))
	case 8:
		val2 = binary.LittleEndian.Uint64(r.buf[offset2 : offset2+8])
	}

	// Value 3 - read exactly len3 bytes
	switch len3 {
	case 1:
		val3 = uint64(r.buf[offset3])
	case 2:
		val3 = uint64(binary.LittleEndian.Uint16(r.buf[offset3 : offset3+2]))
	case 4:
		val3 = uint64(binary.LittleEndian.Uint32(r.buf[offset3 : offset3+4]))
	case 8:
		val3 = binary.LittleEndian.Uint64(r.buf[offset3 : offset3+8])
	}

	r.byteIndex += totalBytes
	return [4]uint64{val0, val1, val2, val3}, nil
}

func (r *BytesReader) readUvar32x4Scalar() ([4]uint32, error) {
	// Lookup tables for branchless operation
	var byteLengths = [4]int{0, 1, 2, 4}
	var masks = [4]uint32{0x0, 0xFF, 0xFFFF, 0xFFFFFFFF}

	// Check if we have at least 1 byte for control byte
	if r.byteIndex >= len(r.buf) {
		return [4]uint32{}, io.EOF
	}

	// Read control byte
	controlByte := r.buf[r.byteIndex]
	r.byteIndex++

	// Extract 2-bit length codes for each value (unrolled)
	code0 := controlByte & 0x3        // bits 0-1
	code1 := (controlByte >> 2) & 0x3 // bits 2-3
	code2 := (controlByte >> 4) & 0x3 // bits 4-5
	code3 := (controlByte >> 6) & 0x3 // bits 6-7

	// Get actual byte lengths for each value
	len0 := byteLengths[code0]
	len1 := byteLengths[code1]
	len2 := byteLengths[code2]
	len3 := byteLengths[code3]

	// Calculate total bytes needed
	totalBytes := len0 + len1 + len2 + len3

	// Check if we have enough bytes remaining
	if len(r.buf[r.byteIndex:]) < totalBytes {
		return [4]uint32{}, io.EOF
	}

	// Calculate offsets for each value
	offset0 := r.byteIndex
	offset1 := offset0 + len0
	offset2 := offset1 + len1
	offset3 := offset2 + len2

	// Initialize values to 0 (for when length is 0)
	var val0, val1, val2, val3 uint32

	// Fast path: check if we have at least 4 bytes available after the last value's offset
	// This allows us to safely use Uint32 reads without bounds checking
	maxOffset := offset3 + 4
	if len(r.buf) >= maxOffset {
		// Fast path: use direct Uint32 reads with masking (no bounds checks needed)
		val0 = binary.LittleEndian.Uint32(r.buf[offset0:]) & masks[code0]
		val1 = binary.LittleEndian.Uint32(r.buf[offset1:]) & masks[code1]
		val2 = binary.LittleEndian.Uint32(r.buf[offset2:]) & masks[code2]
		val3 = binary.LittleEndian.Uint32(r.buf[offset3:]) & masks[code3]

		r.byteIndex += totalBytes
		return [4]uint32{val0, val1, val2, val3}, nil
	}

	// Slow path: carefully read each value with proper bounds checking
	// Value 0 - read exactly len0 bytes (or 0 if len0 is 0)
	if len0 > 0 {
		switch len0 {
		case 1:
			val0 = uint32(r.buf[offset0])
		case 2:
			val0 = uint32(binary.LittleEndian.Uint16(r.buf[offset0 : offset0+2]))
		case 4:
			val0 = binary.LittleEndian.Uint32(r.buf[offset0 : offset0+4])
		}
	}

	// Value 1 - read exactly len1 bytes (or 0 if len1 is 0)
	switch len1 {
	case 1:
		val1 = uint32(r.buf[offset1])
	case 2:
		val1 = uint32(binary.LittleEndian.Uint16(r.buf[offset1 : offset1+2]))
	case 4:
		val1 = binary.LittleEndian.Uint32(r.buf[offset1 : offset1+4])
	}

	// Value 2 - read exactly len2 bytes (or 0 if len2 is 0)
	switch len2 {
	case 1:
		val2 = uint32(r.buf[offset2])
	case 2:
		val2 = uint32(binary.LittleEndian.Uint16(r.buf[offset2 : offset2+2]))
	case 4:
		val2 = binary.LittleEndian.Uint32(r.buf[offset2 : offset2+4])
	}

	// Value 3 - read exactly len3 bytes (or 0 if len3 is 0)
	switch len3 {
	case 1:
		val3 = uint32(r.buf[offset3])
	case 2:
		val3 = uint32(binary.LittleEndian.Uint16(r.buf[offset3 : offset3+2]))
	case 4:
		val3 = binary.LittleEndian.Uint32(r.buf[offset3 : offset3+4])
	}

	r.byteIndex += totalBytes
	return [4]uint32{val0, val1, val2, val3}, nil
}

func (r *BytesReader) readUvar64x2Scalar() ([2]uint64, error) {
	var masks = [9]uint64{
		0x0, 0xFF, 0xFFFF, 0xFFFFFF, 0xFFFFFFFF, 0xFFFFFFFFFF, 0xFFFFFFFFFFFF, 0xFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF,
	}

	// Check if we have at least 1 byte for control byte
	if r.byteIndex >= len(r.buf) {
		return [2]uint64{}, io.EOF
	}

	// Read control byte
	controlByte := r.buf[r.byteIndex]
	r.byteIndex++

	// Extract 4-bit length codes for each value
	code0 := controlByte & 0xF        // bits 0-3
	code1 := (controlByte >> 4) & 0xF // bits 4-7

	// Get actual byte lengths for each value
	len0 := code0
	len1 := code1

	// Calculate total bytes needed
	totalBytes := int(len0 + len1)

	// Check if we have enough bytes remaining
	if len(r.buf[r.byteIndex:]) < totalBytes {
		return [2]uint64{}, io.EOF
	}

	// Calculate offsets for each value
	offset0 := r.byteIndex
	offset1 := offset0 + int(len0)

	// Fast path: check if we have at least 8 bytes available after the last value's offset
	// This allows us to safely use Uint64 reads without bounds checking
	maxOffset := offset1 + 8
	if len(r.buf) >= maxOffset {
		// Fast path: use direct Uint32 reads with masking (no bounds checks needed)
		val0 := binary.LittleEndian.Uint64(r.buf[offset0:]) & masks[code0]
		val1 := binary.LittleEndian.Uint64(r.buf[offset1:]) & masks[code1]

		r.byteIndex += totalBytes
		return [2]uint64{val0, val1}, nil
	}

	// Slow path: read exact bytes

	var val0, val1 uint64

	// Read value 0
	switch len0 {
	case 1:
		val0 = uint64(r.buf[offset0])
	case 2:
		val0 = uint64(binary.LittleEndian.Uint16(r.buf[offset0 : offset0+2]))
	case 3:
		val0 = uint64(r.buf[offset0]) | uint64(binary.LittleEndian.Uint16(r.buf[offset0+1:offset0+3]))<<8
	case 4:
		val0 = uint64(binary.LittleEndian.Uint32(r.buf[offset0 : offset0+4]))
	case 5:
		val0 = uint64(binary.LittleEndian.Uint32(r.buf[offset0:offset0+4])) | uint64(r.buf[offset0+4])<<32
	case 6:
		val0 = uint64(binary.LittleEndian.Uint32(r.buf[offset0:offset0+4])) | uint64(binary.LittleEndian.Uint16(r.buf[offset0+4:offset0+6]))<<32
	case 7:
		val0 = uint64(binary.LittleEndian.Uint32(r.buf[offset0:offset0+4])) | uint64(binary.LittleEndian.Uint16(r.buf[offset0+4:offset0+6]))<<32 | uint64(r.buf[offset0+6])<<48
	case 8:
		val0 = binary.LittleEndian.Uint64(r.buf[offset0 : offset0+8])
	}

	// Read value 1
	switch len1 {
	case 1:
		val1 = uint64(r.buf[offset1])
	case 2:
		val1 = uint64(binary.LittleEndian.Uint16(r.buf[offset1 : offset1+2]))
	case 3:
		val1 = uint64(r.buf[offset1]) | uint64(binary.LittleEndian.Uint16(r.buf[offset1+1:offset1+3]))<<8
	case 4:
		val1 = uint64(binary.LittleEndian.Uint32(r.buf[offset1 : offset1+4]))
	case 5:
		val1 = uint64(binary.LittleEndian.Uint32(r.buf[offset1:offset1+4])) | uint64(r.buf[offset1+4])<<32
	case 6:
		val1 = uint64(binary.LittleEndian.Uint32(r.buf[offset1:offset1+4])) | uint64(binary.LittleEndian.Uint16(r.buf[offset1+4:offset1+6]))<<32
	case 7:
		val1 = uint64(binary.LittleEndian.Uint32(r.buf[offset1:offset1+4])) | uint64(binary.LittleEndian.Uint16(r.buf[offset1+4:offset1+6]))<<32 | uint64(r.buf[offset1+6])<<48
	case 8:
		val1 = binary.LittleEndian.Uint64(r.buf[offset1 : offset1+8])
	}

	r.byteIndex += totalBytes
	return [2]uint64{val0, val1}, nil
}

type BytesWriter struct {
	buf       []byte
	byteIndex int
}

func NewBytesWriter(cap int) BytesWriter {
	return BytesWriter{buf: make([]byte, 0, cap)}
}

func (w *BytesWriter) WriteByte(b byte) {
	w.buf = append(w.buf, b)
}

func (w *BytesWriter) WriteBytes(bytes []byte) {
	w.buf = append(w.buf, bytes...)
}

func (w *BytesWriter) WriteStringBytes(val string) {
	w.buf = append(w.buf, val...)
}

func (w *BytesWriter) WriteUvarint(value uint64) {
	w.buf = binary.AppendUvarint(w.buf, value)
}

func (w *BytesWriter) WriteVarint(x int64) {
	ux := uint64((x >> 63) ^ (x << 1))
	w.WriteUvarint(ux)
}

func (w *BytesWriter) Reset() {
	w.buf = w.buf[:0]
	w.byteIndex = 0
}

func (w *BytesWriter) ResetAndReserve(len int) {
	needCap := len + 8
	if cap(w.buf) < needCap {
		w.buf = append(w.buf[0:cap(w.buf)], make([]byte, int(needCap)-cap(w.buf))...)
	}
	w.buf = w.buf[0:int(len)]
	w.byteIndex = 0
}

func (w *BytesWriter) MapBytesToBitsReader(dest *BitsReader, byteSize int) error {
	if len(w.buf[w.byteIndex:]) < byteSize {
		return io.EOF
	}

	// Map instead of copying.
	dest.buf = w.buf[w.byteIndex : w.byteIndex+byteSize]
	dest.byteIndex = 0
	dest.availBitCount = 0

	w.byteIndex += byteSize
	return nil
}

func (w *BytesWriter) Bytes() []byte {
	return w.buf
}

func (w *BytesWriter) WriteUvar64x4(values [4]uint64) {
	// Lookup table for converting leading zeros to byte length encoding (0-3 for 1,2,4,8 bytes)
	var lengthLookup = [64]byte{
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, // 0-15
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, // 16-31
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, // 32-47
		1, 1, 1, 1, 1, 1, 1, 1, // 48-55
		0, 0, 0, 0, 0, 0, 0, 0, // 56-63
	}

	// Lookup table for converting length code to actual byte count
	var byteLengths = [4]int{1, 2, 4, 8}

	// Unrolled length calculation for all 4 values
	val0, val1, val2, val3 := values[0], values[1], values[2], values[3]

	// Get length codes using lookup table (branchless)
	code0 := lengthLookup[bits.LeadingZeros64(val0|1)]
	code1 := lengthLookup[bits.LeadingZeros64(val1|1)]
	code2 := lengthLookup[bits.LeadingZeros64(val2|1)]
	code3 := lengthLookup[bits.LeadingZeros64(val3|1)]

	// Pack control byte (2 bits per value)
	controlByte := code0 | code1<<2 | code2<<4 | code3<<6

	// Calculate total size needed: 1 control byte + sum of all value lengths
	len0 := byteLengths[code0]
	len1 := byteLengths[code1]
	len2 := byteLengths[code2]
	len3 := byteLengths[code3]
	totalSize := 1 + len0 + len1 + len2 + len3

	// Calculate maximum space needed for PutUint64 operations (worst case: all 8-byte values + control byte)
	maxSpaceNeeded := 1 + 8 + 8 + 8 + 8

	// Pre-allocate buffer space in one operation with enough room for PutUint64 writes
	startIdx := len(w.buf)
	if cap(w.buf) < len(w.buf)+maxSpaceNeeded {
		// Grow buffer capacity if needed
		newCap := (cap(w.buf) + maxSpaceNeeded) * 2
		newBuf := make([]byte, len(w.buf), newCap)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}

	// Extend buffer to maximum needed size first
	w.buf = w.buf[:len(w.buf)+maxSpaceNeeded]

	// Write control byte
	w.buf[startIdx] = byte(controlByte)

	// Calculate offsets for each value
	offset0 := startIdx + 1
	offset1 := offset0 + len0
	offset2 := offset1 + len1
	offset3 := offset2 + len2

	// Write all values using PutUint64 only - each will write 8 bytes but we only use what we need
	binary.LittleEndian.PutUint64(w.buf[offset0:], val0)
	binary.LittleEndian.PutUint64(w.buf[offset1:], val1)
	binary.LittleEndian.PutUint64(w.buf[offset2:], val2)
	binary.LittleEndian.PutUint64(w.buf[offset3:], val3)

	// Resize buffer to actual needed size
	w.buf = w.buf[:startIdx+totalSize]
}

func (w *BytesWriter) writeUvar32x4Scalar(values [4]uint32) {
	// Lookup table for converting leading zeros to byte length encoding (0-3 for 0,1,2,4 bytes)
	// For zero values: bits.LeadingZeros32(0|1) = 31, so we map index 31 to code 0
	var lengthLookup = [33]byte{
		3, 3, 3, 3, 3, 3, 3, 3, // 0-7   (values >= 2^24: 4 bytes)
		3, 3, 3, 3, 3, 3, 3, 3, // 8-15  (values >= 2^16: 4 bytes)
		2, 2, 2, 2, 2, 2, 2, 2, // 16-23 (values >= 2^8: 2 bytes)
		1, 1, 1, 1, 1, 1, 1, 1, // 24-31 (values >= 2^0: 1 byte)
		0, // 32 (for zero value)
	}

	// Lookup table for converting length code to actual byte count
	var byteLengths = [4]int{0, 1, 2, 4}

	// Unrolled length calculation for all 4 values
	val0, val1, val2, val3 := values[0], values[1], values[2], values[3]

	// Get length codes using lookup table (branchless)
	// bits.LeadingZeros32(0|1) = 31, which maps to code 0 (0 bytes) in our lookup table
	code0 := lengthLookup[bits.LeadingZeros32(val0)]
	code1 := lengthLookup[bits.LeadingZeros32(val1)]
	code2 := lengthLookup[bits.LeadingZeros32(val2)]
	code3 := lengthLookup[bits.LeadingZeros32(val3)]

	// Pack control byte (2 bits per value)
	controlByte := code0 | code1<<2 | code2<<4 | code3<<6

	// Calculate total size needed: 1 control byte + sum of all value lengths
	len0 := byteLengths[code0]
	len1 := byteLengths[code1]
	len2 := byteLengths[code2]
	len3 := byteLengths[code3]
	totalSize := 1 + len0 + len1 + len2 + len3

	// Calculate maximum space needed for PutUint32 operations (worst case: all 4-byte values + control byte)
	maxSpaceNeeded := 1 + 4 + 4 + 4 + 4

	// Pre-allocate buffer space in one operation with enough room for PutUint32 writes
	startIdx := len(w.buf)
	w.buf = EnsureLen(w.buf, len(w.buf)+maxSpaceNeeded)

	// Write control byte
	w.buf[startIdx] = controlByte

	// Calculate offsets for each value
	offset0 := startIdx + 1
	offset1 := offset0 + len0
	offset2 := offset1 + len1
	offset3 := offset2 + len2

	// Write all non-zero values using PutUint32 only - each will write 4 bytes but we only use what we need
	binary.LittleEndian.PutUint32(w.buf[offset0:], val0)
	binary.LittleEndian.PutUint32(w.buf[offset1:], val1)
	binary.LittleEndian.PutUint32(w.buf[offset2:], val2)
	binary.LittleEndian.PutUint32(w.buf[offset3:], val3)

	// Resize buffer to actual needed size
	w.buf = w.buf[:startIdx+totalSize]
}

func (w *BytesWriter) writeUvar64x2Scalar(values [2]uint64) {
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
	len0 := lengthLookup[bits.LeadingZeros64(val0)]
	len1 := lengthLookup[bits.LeadingZeros64(val1)]

	// Pack control byte (4 bits per value)
	controlByte := len0 | len1<<4

	// Calculate total size needed: 1 control byte + sum of both value lengths
	totalSize := 1 + int(len0+len1)

	// Calculate maximum space needed for PutUint64 operations (worst case: both 8-byte values + control byte)
	maxSpaceNeeded := 1 + 8 + 8

	// Pre-allocate buffer space in one operation with enough room for PutUint64 writes
	startIdx := len(w.buf)
	w.buf = EnsureLen(w.buf, len(w.buf)+maxSpaceNeeded)

	// Write control byte
	w.buf[startIdx] = controlByte

	// Calculate offsets for each value
	offset0 := startIdx + 1
	offset1 := offset0 + int(len0)

	binary.LittleEndian.PutUint64(w.buf[offset0:], val0)
	binary.LittleEndian.PutUint64(w.buf[offset1:], val1)

	// Resize buffer to actual needed size
	w.buf = w.buf[:startIdx+totalSize]
}
