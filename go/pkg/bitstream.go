//go:generate tablegen

package pkg

// Originally from https://github.com/dgryski/go-tsz/blob/master/bstream.go
//
// Copyright (c) 2015,2016 Damian Gryski <damian@gryski.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

import (
	"encoding/binary"
	"math/bits"
)

// BitsWriter is a stream of bits
type BitsWriter struct {
	// The output byte stream.
	stream []byte

	// Temporary buffer of bits.
	bitsBuf uint64

	// Number of written bits in bitsBuf.
	bitsBufUsed uint
}

func NewBitsWriter(size int) BitsWriter {
	return BitsWriter{stream: make([]byte, 0, size), bitsBufUsed: 0}
}

func (b *BitsWriter) Reset() {
	b.bitsBuf = 0
	b.bitsBufUsed = 0
	b.stream = b.stream[:0]
}

// Close finalizes pending write operations. Must be called before Bytes().
// If you want to reuse the BitsWriter after calling Close() then call
// Reset() before using this writer.
func (b *BitsWriter) Close() {
	targetLen := len(b.stream) + int(b.bitsBufUsed+7)/8
	b.stream = binary.BigEndian.AppendUint64(b.stream, b.bitsBuf)
	b.stream = b.stream[:targetLen]
}

func (b *BitsWriter) Bytes() []byte {
	return b.stream
}

func (b *BitsWriter) BitCount() uint {
	return uint(len(b.stream))*8 + b.bitsBufUsed
}

func (b *BitsWriter) WriteBit(bit uint) {
	if b.bitsBufUsed <= 63 {
		b.bitsBuf |= uint64(bit) << (63 - b.bitsBufUsed)
		b.bitsBufUsed++
		return
	}
	b.writeBitsSlow(uint64(bit), 1)
}

func (b *BitsWriter) WriteBits(val uint64, nbits uint) {
	nbitsComplement := 64 - nbits
	if b.bitsBufUsed <= nbitsComplement {
		b.bitsBuf |= val << (nbitsComplement - b.bitsBufUsed)
		b.bitsBufUsed += nbits
		return
	}
	b.writeBitsSlow(val, nbits)
}

func (b *BitsWriter) writeBitsSlow(val uint64, nbits uint) {
	// Complete bitsBuf to 64 bits.
	bitsBufFree := 64 - b.bitsBufUsed
	b.bitsBuf |= val >> (nbits - bitsBufFree)

	// And append 64 bits to stream.
	b.stream = binary.BigEndian.AppendUint64(b.stream, b.bitsBuf)

	// Write the rest of bits
	nbits -= bitsBufFree
	b.bitsBuf = val << (64 - nbits)
	b.bitsBufUsed = nbits
}

// WriteVarintCompact reads a variable-bit encoded signed value.
// The range of supported values is [-2^47..2^47-1].
// Returns the number of bits written.
func (b *BitsWriter) WriteVarintCompact(val int64) uint {
	ux := uint64((val >> 63) ^ (val << 1))
	return b.WriteUvarintCompact(ux)
}

// WriteUvarintCompact writes a variable-bit encoded value.
// The range of supported values is [0..2^48-1].
// Returns the number of bits written.
func (b *BitsWriter) WriteUvarintCompact(val uint64) uint {
	// The format is the following:
	// Prefix Bits:   Followed by big endian bits:
	// 1              Nothing. Encodes value of 0.
	// 01             2 bit value
	// 001            5 bit value
	// 0001           12 bit value
	// 00001          19 bit value
	// 000001         26 bit value
	// 0000001        33 bit value
	// 00000001       48 bit value

	zeros := bits.LeadingZeros64(val)
	val |= writeMaskByZeros[zeros]
	bitCount := writeBitsCountByZeros[zeros]
	b.WriteBits(val, bitCount)
	return bitCount
}

// BitsReader is a reader of bits.
type BitsReader struct {
	// Always contains 64 usable bits to serve from.
	bitBuf uint64
	// Contains up to 64 bits to be shifted into bitBuf as needed.
	bitBufNext uint64
	// Number of usable bits in bitBufNext.
	availBitCount int

	// Input byte buffer.
	buf []byte
	// Position to read next from the buf.
	byteIndex uint

	// True if attempt to read past buf was detected.
	isEOF bool
}

func NewBitsReader() *BitsReader {
	return &BitsReader{}
}

func (b *BitsReader) Reset(buf []byte) {
	b.buf = buf
	b.byteIndex = 0
	b.isEOF = false
	b.bitBuf = 0
	b.bitBufNext = 0
	b.availBitCount = 0
	b.fillInitial()
}

// fillInitial fills bitBuf and bitBufNext from the input buffer.
func (b *BitsReader) fillInitial() {
	b.bitBuf = b.read64Bits()
	b.bitBufNext = b.read64Bits()
	b.availBitCount = 64
}

// read64Bits reads up to 8 bytes from the buffer and returns as uint64 (big endian, zero-padded).
func (b *BitsReader) read64Bits() uint64 {
	var val uint64
	remaining := uint(len(b.buf)) - b.byteIndex
	if remaining >= 8 {
		val = binary.BigEndian.Uint64(b.buf[b.byteIndex:])
		b.byteIndex += 8
	} else if remaining > 0 {
		for i := uint(0); i < remaining; i++ {
			val |= uint64(b.buf[b.byteIndex+i]) << (56 - 8*i)
		}
		b.byteIndex += remaining
	} else {
		val = 0
	}
	return val
}

func (b *BitsReader) MapBytesFromMemBuf(src *BytesReader, byteSize int) error {
	buf, err := src.ReadBytesMapped(byteSize)
	if err != nil {
		return err
	}

	b.buf = buf
	b.byteIndex = 0
	b.isEOF = false
	b.bitBuf = 0
	b.bitBufNext = 0
	b.availBitCount = 0
	b.fillInitial()
	return nil
}

func (b *BitsReader) IsEOF() bool {
	return b.isEOF
}

// PeekBits always serves from bitBuf, which always has 64 bits.
// nbits must be <= 56.
func (b *BitsReader) PeekBits(nbits uint) uint64 {
	if nbits > 64 {
		panic("at most 64 bits can be peeked")
	}
	return b.bitBuf >> (64 - nbits)
}

// Consume advances the bit pointer by nbits bits, shifting in bits from bitBufNext as needed.
// nbits must be <= 56.
func (b *BitsReader) Consume(nbits uint) {
	if b.availBitCount >= int(nbits) {
		// Fast path: enough bits in bitBufNext to satisfy the request.
		b.bitBuf = (b.bitBuf << nbits) | (b.bitBufNext >> (64 - nbits))
		b.bitBufNext <<= nbits
		b.availBitCount -= int(nbits)
		return
	}
	b.consumeSlow(nbits)
}

//go:noinline
func (b *BitsReader) consumeSlow(nbits uint) {
	b.bitBuf = (b.bitBuf << b.availBitCount) | (b.bitBufNext >> (64 - b.availBitCount))
	nbits -= uint(b.availBitCount)
	b.bitBufNext = b.read64Bits()
	b.availBitCount = 64
	b.bitBuf = (b.bitBuf << nbits) | (b.bitBufNext >> (64 - nbits))
	b.bitBufNext <<= nbits
	b.availBitCount -= int(nbits)
}

// ReadBits reads bits. nbits should be in [0..64] range.
// Reading past EOF produces additional 0 bits on the least significant side.
// IsEOF condition will be set if ReadBits is called again after
// consuming all available bits.
func (b *BitsReader) ReadBits(nbits uint) uint64 {
	val := b.PeekBits(nbits)
	b.Consume(nbits)
	return val
}

func (b *BitsReader) ReadBit() uint64 {
	return b.ReadBits(1)
}

// ReadVarintCompact reads a variable-bit encoded signed value.
// The range of supported values is [-2^47..2^47-1].
func (b *BitsReader) ReadVarintCompact() int64 {
	x := b.ReadUvarintCompact()
	// zigzag decode
	return int64((x >> 1) ^ (-(x & 1)))
}

// ReadUvarintCompact reads a variable-bit encoded unsigned value.
// The range of supported values is [0..2^48-1].
func (b *BitsReader) ReadUvarintCompact() uint64 {
	val := b.PeekBits(56)
	zeros := bits.LeadingZeros64(val)
	ret := (val >> readShiftByZeros[zeros]) & readMaskByZeros[zeros]
	b.Consume(readConsumeCountByZeros[zeros])
	return ret
}
