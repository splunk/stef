// Code generated by stefgen. DO NOT EDIT.
package oteltef

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"strings"
	"unsafe"

	"github.com/splunk/stef/go/pkg"
	"github.com/splunk/stef/go/pkg/encoders"
	"github.com/splunk/stef/go/pkg/schema"
)

var _ = strings.Compare
var _ = encoders.StringEncoder{}
var _ = schema.WireSchema{}
var _ = bytes.NewBuffer

type SpanStatus struct {
	message string
	code    uint64

	// modifiedFields keeps track of which fields are modified.
	modifiedFields modifiedFields
}

const SpanStatusStructName = "SpanStatus"

// Bitmasks for "modified" flags for each field.
const (
	fieldModifiedSpanStatusMessage = uint64(1 << iota)
	fieldModifiedSpanStatusCode
)

// Init must be called once, before the SpanStatus is used.
func (s *SpanStatus) Init() {
	s.init(nil, 0)
}

func NewSpanStatus() *SpanStatus {
	var s SpanStatus
	s.init(nil, 0)
	return &s
}

func (s *SpanStatus) init(parentModifiedFields *modifiedFields, parentModifiedBit uint64) {
	s.modifiedFields.parent = parentModifiedFields
	s.modifiedFields.parentBit = parentModifiedBit

}

func (s *SpanStatus) Message() string {
	return s.message
}

// SetMessage sets the value of Message field.
func (s *SpanStatus) SetMessage(v string) {
	if !pkg.StringEqual(s.message, v) {
		s.message = v
		s.markMessageModified()
	}
}

func (s *SpanStatus) markMessageModified() {
	s.modifiedFields.markModified(fieldModifiedSpanStatusMessage)
}

// IsMessageModified returns true the value of Message field was modified since
// SpanStatus was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *SpanStatus) IsMessageModified() bool {
	return s.modifiedFields.mask&fieldModifiedSpanStatusMessage != 0
}

func (s *SpanStatus) Code() uint64 {
	return s.code
}

// SetCode sets the value of Code field.
func (s *SpanStatus) SetCode(v uint64) {
	if !pkg.Uint64Equal(s.code, v) {
		s.code = v
		s.markCodeModified()
	}
}

func (s *SpanStatus) markCodeModified() {
	s.modifiedFields.markModified(fieldModifiedSpanStatusCode)
}

// IsCodeModified returns true the value of Code field was modified since
// SpanStatus was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *SpanStatus) IsCodeModified() bool {
	return s.modifiedFields.mask&fieldModifiedSpanStatusCode != 0
}

func (s *SpanStatus) markModifiedRecursively() {

	s.modifiedFields.mask =
		fieldModifiedSpanStatusMessage |
			fieldModifiedSpanStatusCode | 0
}

func (s *SpanStatus) markUnmodifiedRecursively() {

	if s.IsMessageModified() {
	}

	if s.IsCodeModified() {
	}

	s.modifiedFields.mask = 0
}

// markDiffModified marks fields in this struct modified if they differ from
// the corresponding fields in v.
func (s *SpanStatus) markDiffModified(v *SpanStatus) (modified bool) {
	if !pkg.StringEqual(s.message, v.message) {
		s.markMessageModified()
		modified = true
	}

	if !pkg.Uint64Equal(s.code, v.code) {
		s.markCodeModified()
		modified = true
	}

	return modified
}

func (s *SpanStatus) Clone() SpanStatus {
	return SpanStatus{
		message: s.message,
		code:    s.code,
	}
}

// ByteSize returns approximate memory usage in bytes. Used to calculate
// memory used by dictionaries.
func (s *SpanStatus) byteSize() uint {
	return uint(unsafe.Sizeof(*s)) +
		0
}

func copySpanStatus(dst *SpanStatus, src *SpanStatus) {
	dst.SetMessage(src.message)
	dst.SetCode(src.code)
}

// CopyFrom() performs a deep copy from src.
func (s *SpanStatus) CopyFrom(src *SpanStatus) {
	copySpanStatus(s, src)
}

func (s *SpanStatus) markParentModified() {
	s.modifiedFields.parent.markModified(s.modifiedFields.parentBit)
}

func (s *SpanStatus) markUnmodified() {
	s.modifiedFields.markUnmodified()
}

// mutateRandom mutates fields in a random, deterministic manner using
// random parameter as a deterministic generator.
func (s *SpanStatus) mutateRandom(random *rand.Rand) {
	const fieldCount = 2
	if random.IntN(fieldCount) == 0 {
		s.SetMessage(pkg.StringRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetCode(pkg.Uint64Random(random))
	}
}

// IsEqual performs deep comparison and returns true if struct is equal to val.
func (e *SpanStatus) IsEqual(val *SpanStatus) bool {
	if !pkg.StringEqual(e.message, val.message) {
		return false
	}
	if !pkg.Uint64Equal(e.code, val.code) {
		return false
	}

	return true
}

func SpanStatusEqual(left, right *SpanStatus) bool {
	return left.IsEqual(right)
}

// CmpSpanStatus performs deep comparison and returns an integer that
// will be 0 if left == right, negative if left < right, positive if left > right.
func CmpSpanStatus(left, right *SpanStatus) int {
	if left == nil {
		if right == nil {
			return 0
		}
		return -1
	}
	if right == nil {
		return 1
	}

	if c := strings.Compare(left.message, right.message); c != 0 {
		return c
	}
	if c := pkg.Uint64Compare(left.code, right.code); c != 0 {
		return c
	}

	return 0
}

// SpanStatusEncoder implements encoding of SpanStatus
type SpanStatusEncoder struct {
	buf     pkg.BitsWriter
	limiter *pkg.SizeLimiter

	// forceModifiedFields is set to true if the next encoding operation
	// must write all fields, whether they are modified or no.
	// This is used after frame restarts so that the data can be decoded
	// from the frame start.
	forceModifiedFields bool

	messageEncoder encoders.StringEncoder
	codeEncoder    encoders.Uint64Encoder

	keepFieldMask uint64
	fieldCount    uint
}

func (e *SpanStatusEncoder) Init(state *WriterState, columns *pkg.WriteColumnSet) error {
	// Remember this encoder in the state so that we can detect recursion.
	if state.SpanStatusEncoder != nil {
		panic("cannot initialize SpanStatusEncoder: already initialized")
	}
	state.SpanStatusEncoder = e
	defer func() { state.SpanStatusEncoder = nil }()

	e.limiter = &state.limiter

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("SpanStatus")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "SpanStatus")
		}

		// Number of fields in the target schema.
		e.fieldCount = fieldCount

		// Set that many 1 bits in the keepFieldMask. All fields with higher number
		// will be skipped when encoding.
		e.keepFieldMask = ^(^uint64(0) << e.fieldCount)
	} else {
		// Keep all fields when encoding.
		e.fieldCount = 2
		e.keepFieldMask = ^uint64(0)
	}

	if e.fieldCount <= 0 {
		return nil // Message and subsequent fields are skipped.
	}
	if err := e.messageEncoder.Init(nil, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 1 {
		return nil // Code and subsequent fields are skipped.
	}
	if err := e.codeEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}

	return nil
}

func (e *SpanStatusEncoder) Reset() {
	// Since we are resetting the state of encoder make sure the next Encode()
	// call forcedly writes all fields and does not attempt to skip.
	e.forceModifiedFields = true
	e.messageEncoder.Reset()
	e.codeEncoder.Reset()
}

// Encode encodes val into buf
func (e *SpanStatusEncoder) Encode(val *SpanStatus) {
	var bitCount uint

	// Mask that describes what fields are encoded. Start with all modified fields.
	fieldMask := val.modifiedFields.mask

	// If forceModifiedFields we need to set to 1 all bits so that we
	// force writing of all fields.
	if e.forceModifiedFields {
		fieldMask =
			fieldModifiedSpanStatusMessage |
				fieldModifiedSpanStatusCode | 0
	}

	// Only write fields that we want to write. See Init() for keepFieldMask.
	fieldMask &= e.keepFieldMask

	// Write bits to indicate which fields follow.
	e.buf.WriteBits(fieldMask, e.fieldCount)
	bitCount += e.fieldCount

	// Encode modified, present fields.

	if fieldMask&fieldModifiedSpanStatusMessage != 0 {
		// Encode Message
		e.messageEncoder.Encode(val.message)
	}

	if fieldMask&fieldModifiedSpanStatusCode != 0 {
		// Encode Code
		e.codeEncoder.Encode(val.code)
	}

	// Account written bits in the limiter.
	e.limiter.AddFrameBits(bitCount)

	// Mark all fields non-modified so that next Encode() correctly
	// encodes only fields that change after this.
	val.modifiedFields.mask = 0
}

// CollectColumns collects all buffers from all encoders into buf.
func (e *SpanStatusEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBits(&e.buf)

	if e.fieldCount <= 0 {
		return // Message and subsequent fields are skipped.
	}
	e.messageEncoder.CollectColumns(columnSet.At(0))
	if e.fieldCount <= 1 {
		return // Code and subsequent fields are skipped.
	}
	e.codeEncoder.CollectColumns(columnSet.At(1))
}

// SpanStatusDecoder implements decoding of SpanStatus
type SpanStatusDecoder struct {
	buf        pkg.BitsReader
	column     *pkg.ReadableColumn
	lastValPtr *SpanStatus
	lastVal    SpanStatus
	fieldCount uint

	messageDecoder encoders.StringDecoder
	codeDecoder    encoders.Uint64Decoder
}

// Init is called once in the lifetime of the stream.
func (d *SpanStatusDecoder) Init(state *ReaderState, columns *pkg.ReadColumnSet) error {
	// Remember this decoder in the state so that we can detect recursion.
	if state.SpanStatusDecoder != nil {
		panic("cannot initialize SpanStatusDecoder: already initialized")
	}
	state.SpanStatusDecoder = d
	defer func() { state.SpanStatusDecoder = nil }()

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("SpanStatus")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "SpanStatus")
		}

		// Number of fields in the target schema.
		d.fieldCount = fieldCount
	} else {
		// Keep all fields when encoding.
		d.fieldCount = 2
	}

	d.column = columns.Column()

	d.lastVal.init(nil, 0)
	d.lastValPtr = &d.lastVal

	var err error

	if d.fieldCount <= 0 {
		return nil // Message and subsequent fields are skipped.
	}
	err = d.messageDecoder.Init(nil, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 1 {
		return nil // Code and subsequent fields are skipped.
	}
	err = d.codeDecoder.Init(columns.AddSubColumn())
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
func (d *SpanStatusDecoder) Continue() {
	d.buf.Reset(d.column.Data())

	if d.fieldCount <= 0 {
		return // Message and subsequent fields are skipped.
	}
	d.messageDecoder.Continue()
	if d.fieldCount <= 1 {
		return // Code and subsequent fields are skipped.
	}
	d.codeDecoder.Continue()
}

func (d *SpanStatusDecoder) Reset() {
	d.messageDecoder.Reset()
	d.codeDecoder.Reset()
}

func (d *SpanStatusDecoder) Decode(dstPtr *SpanStatus) error {
	val := dstPtr

	var err error

	// Read bits that indicate which fields follow.
	val.modifiedFields.mask = d.buf.ReadBits(d.fieldCount)

	if val.modifiedFields.mask&fieldModifiedSpanStatusMessage != 0 {
		// Field is changed and is present, decode it.
		err = d.messageDecoder.Decode(&val.message)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedSpanStatusCode != 0 {
		// Field is changed and is present, decode it.
		err = d.codeDecoder.Decode(&val.code)
		if err != nil {
			return err
		}
	}

	return nil
}
