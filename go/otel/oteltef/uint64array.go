// Code generated by stefgen. DO NOT EDIT.
package oteltef

import (
	"math/rand/v2"

	"slices"

	"strings"
	"unsafe"

	"github.com/splunk/stef/go/pkg"
	"github.com/splunk/stef/go/pkg/encoders"
)

var _ = (*encoders.StringEncoder)(nil)
var _ = (*strings.Builder)(nil)

// Uint64Array is a variable size array.
type Uint64Array struct {
	elems []uint64

	parentModifiedFields *modifiedFields
	parentModifiedBit    uint64
}

func (e *Uint64Array) init(parentModifiedFields *modifiedFields, parentModifiedBit uint64) {
	e.parentModifiedFields = parentModifiedFields
	e.parentModifiedBit = parentModifiedBit
}

// Clone() creates a deep copy of Uint64Array
func (e *Uint64Array) Clone() Uint64Array {
	var clone Uint64Array
	copyUint64Array(&clone, e)
	return clone
}

// ByteSize returns approximate memory usage in bytes. Used to calculate
// memory used by dictionaries.
func (e *Uint64Array) byteSize() uint {
	if len(e.elems) == 0 {
		return 0
	}
	// TODO: add size of elements if they are clonable.
	size := uint(unsafe.Sizeof(e.elems[0]))*uint(len(e.elems)) + uint(unsafe.Sizeof(e))

	return size
}

// CopyFromSlice copies from a slice into this array. The length
// of the array will be equal to the length of slice and elements of
// the array will be assigned from elements of the slice.
func (e *Uint64Array) CopyFromSlice(src []uint64) {
	if !slices.Equal(e.elems, src) {
		e.elems = pkg.EnsureLen(e.elems, len(src))
		copy(e.elems, src)
		e.markModified()
	}
}

// Append a new element at the end of the array.
func (e *Uint64Array) Append(val uint64) {
	e.elems = append(e.elems, val)
	e.markModified()
}

func (e *Uint64Array) markModified() {
	e.parentModifiedFields.markModified(e.parentModifiedBit)
}

func (e *Uint64Array) markUnmodified() {
	e.parentModifiedFields.markUnmodified()
}

func (e *Uint64Array) markModifiedRecursively() {

}

func (e *Uint64Array) markUnmodifiedRecursively() {

}

// markDiffModified marks fields in each element of this array modified if they differ from
// the corresponding fields in v.
func (e *Uint64Array) markDiffModified(v *Uint64Array) (modified bool) {
	if len(e.elems) != len(v.elems) {
		// Array lengths are different, so they are definitely different.
		modified = true
	}

	// Scan the elements and mark them as modified if they are different.
	minLen := min(len(e.elems), len(v.elems))
	for i := 0; i < minLen; i++ {
		if !pkg.Uint64Equal(e.elems[i], v.elems[i]) {
			modified = true
		}

	}

	if modified {
		e.markModified()
	}

	return modified
}

func copyUint64Array(dst *Uint64Array, src *Uint64Array) {
	isModified := false

	minLen := min(len(dst.elems), len(src.elems))
	if len(dst.elems) != len(src.elems) {
		dst.elems = pkg.EnsureLen(dst.elems, len(src.elems))
		isModified = true
	}

	i := 0

	// Copy elements in the part of the array that already had the necessary room.
	for ; i < minLen; i++ {
		if dst.elems[i] != src.elems[i] {
			dst.elems[i] = src.elems[i]
			isModified = true
		}
	}
	for ; i < len(dst.elems); i++ {
		if dst.elems[i] != src.elems[i] {
			dst.elems[i] = src.elems[i]
			isModified = true
		}
	}
	if isModified {
		dst.markModified()
	}
}

// Len returns the number of elements in the array.
func (e *Uint64Array) Len() int {
	return len(e.elems)
}

// At returns element at index i.
func (m *Uint64Array) At(i int) uint64 {
	return m.elems[i]
}

// EnsureLen ensures the length of the array is equal to newLen.
// It will grow or shrink the array if needed.
func (e *Uint64Array) EnsureLen(newLen int) {
	oldLen := len(e.elems)
	if newLen > oldLen {
		// Grow the array
		e.elems = append(e.elems, make([]uint64, newLen-oldLen)...)
		e.markModified()
	} else if oldLen > newLen {
		// Shrink it
		e.elems = e.elems[:newLen]
		e.markModified()
	}
}

// IsEqual performs deep comparison and returns true if array is equal to val.
func (e *Uint64Array) IsEqual(val *Uint64Array) bool {
	if len(e.elems) != len(val.elems) {
		return false
	}
	for i := range e.elems {
		if !pkg.Uint64Equal(e.elems[i], val.elems[i]) {
			return false
		}
	}
	return true
}

// CmpUint64Array performs deep comparison and returns an integer that
// will be 0 if left == right, negative if left < right, positive if left > right.
func CmpUint64Array(left, right *Uint64Array) int {
	c := len(left.elems) - len(right.elems)
	if c != 0 {
		return c
	}
	for i := range left.elems {
		fc := pkg.Uint64Compare(left.elems[i], right.elems[i])
		if fc < 0 {
			return -1
		}
		if fc > 0 {
			return 1
		}
	}
	return 0
}

// mutateRandom mutates fields in a random, deterministic manner using
// random parameter as a deterministic generator.
func (a *Uint64Array) mutateRandom(random *rand.Rand) {
	if random.IntN(20) == 0 {
		a.EnsureLen(a.Len() + 1)
	}
	if random.IntN(20) == 0 && a.Len() > 0 {
		a.EnsureLen(a.Len() - 1)
	}

	for i := range a.elems {
		_ = i
		if random.IntN(2*len(a.elems)) == 0 {
			v := pkg.Uint64Random(random)
			if a.elems[i] != v {
				a.elems[i] = v
				a.markModified()
			}
		}
	}
}

type Uint64ArrayEncoder struct {
	buf         pkg.BitsWriter
	limiter     *pkg.SizeLimiter
	elemEncoder *encoders.Uint64Encoder
	isRecursive bool
	state       *WriterState
	prevLen     int
}

func (e *Uint64ArrayEncoder) Init(state *WriterState, columns *pkg.WriteColumnSet) error {
	e.state = state
	e.limiter = &state.limiter

	e.elemEncoder = new(encoders.Uint64Encoder)
	if err := e.elemEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}

	return nil
}

func (e *Uint64ArrayEncoder) Reset() {
	if !e.isRecursive {
		e.elemEncoder.Reset()
	}
	e.prevLen = 0
}

func (e *Uint64ArrayEncoder) Encode(arr *Uint64Array) {

	newLen := len(arr.elems)
	oldBitLen := e.buf.BitCount()

	lenDelta := newLen - e.prevLen
	e.prevLen = newLen

	e.buf.WriteVarintCompact(int64(lenDelta))

	if newLen > 0 {
		for i := 0; i < newLen; i++ {
			e.elemEncoder.Encode(arr.elems[i])
		}
	}

	// Account written bits in the limiter.
	newBitLen := e.buf.BitCount()
	e.limiter.AddFrameBits(newBitLen - oldBitLen)
}

func (e *Uint64ArrayEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBits(&e.buf)
	if !e.isRecursive {
		e.elemEncoder.CollectColumns(columnSet.At(0))
	}
}

type Uint64ArrayDecoder struct {
	buf         pkg.BitsReader
	column      *pkg.ReadableColumn
	elemDecoder *encoders.Uint64Decoder
	isRecursive bool
	prevLen     int
}

// Init is called once in the lifetime of the stream.
func (d *Uint64ArrayDecoder) Init(state *ReaderState, columns *pkg.ReadColumnSet) error {
	d.column = columns.Column()
	d.elemDecoder = new(encoders.Uint64Decoder)
	err := d.elemDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}

	return nil
}

// Continue is called at the start of the frame to continue decoding column data.
// This should set the decoder's source buffer, so the new decoding continues from
// the supplied column data. This should NOT reset the internal state of the decoder,
// since columns can cross frame boundaries and the new column data is considered
// continuation of that same column in the previous frame.
func (d *Uint64ArrayDecoder) Continue() {
	d.buf.Reset(d.column.Data())
	if !d.isRecursive {
		d.elemDecoder.Continue()
	}
}

func (d *Uint64ArrayDecoder) Reset() {
	if !d.isRecursive {
		d.elemDecoder.Reset()
	}
	d.prevLen = 0
}

func (d *Uint64ArrayDecoder) Decode(dst *Uint64Array) error {

	lenDelta := d.buf.ReadVarintCompact()

	newLen := d.prevLen + int(lenDelta)
	d.prevLen = newLen

	dst.EnsureLen(newLen)

	for i := 0; i < newLen; i++ {
		err := d.elemDecoder.Decode(&dst.elems[i])
		if err != nil {
			return err
		}
	}

	return nil
}
