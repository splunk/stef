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

type Link struct {
	traceID                pkg.Bytes
	spanID                 pkg.Bytes
	traceState             string
	flags                  uint64
	attributes             Attributes
	droppedAttributesCount uint64

	// modifiedFields keeps track of which fields are modified.
	modifiedFields modifiedFields
}

const LinkStructName = "Link"

// Bitmasks for "modified" flags for each field.
const (
	fieldModifiedLinkTraceID = uint64(1 << iota)
	fieldModifiedLinkSpanID
	fieldModifiedLinkTraceState
	fieldModifiedLinkFlags
	fieldModifiedLinkAttributes
	fieldModifiedLinkDroppedAttributesCount
)

// Init must be called once, before the Link is used.
func (s *Link) Init() {
	s.init(nil, 0)
}

func NewLink() *Link {
	var s Link
	s.init(nil, 0)
	return &s
}

func (s *Link) init(parentModifiedFields *modifiedFields, parentModifiedBit uint64) {
	s.modifiedFields.parent = parentModifiedFields
	s.modifiedFields.parentBit = parentModifiedBit

	s.attributes.init(&s.modifiedFields, fieldModifiedLinkAttributes)
}

func (s *Link) TraceID() pkg.Bytes {
	return s.traceID
}

// SetTraceID sets the value of TraceID field.
func (s *Link) SetTraceID(v pkg.Bytes) {
	if !pkg.BytesEqual(s.traceID, v) {
		s.traceID = v
		s.markTraceIDModified()
	}
}

func (s *Link) markTraceIDModified() {
	s.modifiedFields.markModified(fieldModifiedLinkTraceID)
}

// IsTraceIDModified returns true the value of TraceID field was modified since
// Link was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Link) IsTraceIDModified() bool {
	return s.modifiedFields.mask&fieldModifiedLinkTraceID != 0
}

func (s *Link) SpanID() pkg.Bytes {
	return s.spanID
}

// SetSpanID sets the value of SpanID field.
func (s *Link) SetSpanID(v pkg.Bytes) {
	if !pkg.BytesEqual(s.spanID, v) {
		s.spanID = v
		s.markSpanIDModified()
	}
}

func (s *Link) markSpanIDModified() {
	s.modifiedFields.markModified(fieldModifiedLinkSpanID)
}

// IsSpanIDModified returns true the value of SpanID field was modified since
// Link was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Link) IsSpanIDModified() bool {
	return s.modifiedFields.mask&fieldModifiedLinkSpanID != 0
}

func (s *Link) TraceState() string {
	return s.traceState
}

// SetTraceState sets the value of TraceState field.
func (s *Link) SetTraceState(v string) {
	if !pkg.StringEqual(s.traceState, v) {
		s.traceState = v
		s.markTraceStateModified()
	}
}

func (s *Link) markTraceStateModified() {
	s.modifiedFields.markModified(fieldModifiedLinkTraceState)
}

// IsTraceStateModified returns true the value of TraceState field was modified since
// Link was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Link) IsTraceStateModified() bool {
	return s.modifiedFields.mask&fieldModifiedLinkTraceState != 0
}

func (s *Link) Flags() uint64 {
	return s.flags
}

// SetFlags sets the value of Flags field.
func (s *Link) SetFlags(v uint64) {
	if !pkg.Uint64Equal(s.flags, v) {
		s.flags = v
		s.markFlagsModified()
	}
}

func (s *Link) markFlagsModified() {
	s.modifiedFields.markModified(fieldModifiedLinkFlags)
}

// IsFlagsModified returns true the value of Flags field was modified since
// Link was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Link) IsFlagsModified() bool {
	return s.modifiedFields.mask&fieldModifiedLinkFlags != 0
}

func (s *Link) Attributes() *Attributes {
	return &s.attributes
}

// IsAttributesModified returns true the value of Attributes field was modified since
// Link was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Link) IsAttributesModified() bool {
	return s.modifiedFields.mask&fieldModifiedLinkAttributes != 0
}

func (s *Link) DroppedAttributesCount() uint64 {
	return s.droppedAttributesCount
}

// SetDroppedAttributesCount sets the value of DroppedAttributesCount field.
func (s *Link) SetDroppedAttributesCount(v uint64) {
	if !pkg.Uint64Equal(s.droppedAttributesCount, v) {
		s.droppedAttributesCount = v
		s.markDroppedAttributesCountModified()
	}
}

func (s *Link) markDroppedAttributesCountModified() {
	s.modifiedFields.markModified(fieldModifiedLinkDroppedAttributesCount)
}

// IsDroppedAttributesCountModified returns true the value of DroppedAttributesCount field was modified since
// Link was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Link) IsDroppedAttributesCountModified() bool {
	return s.modifiedFields.mask&fieldModifiedLinkDroppedAttributesCount != 0
}

func (s *Link) markModifiedRecursively() {

	s.attributes.markModifiedRecursively()

	s.modifiedFields.mask =
		fieldModifiedLinkTraceID |
			fieldModifiedLinkSpanID |
			fieldModifiedLinkTraceState |
			fieldModifiedLinkFlags |
			fieldModifiedLinkAttributes |
			fieldModifiedLinkDroppedAttributesCount | 0
}

func (s *Link) markUnmodifiedRecursively() {

	if s.IsTraceIDModified() {
	}

	if s.IsSpanIDModified() {
	}

	if s.IsTraceStateModified() {
	}

	if s.IsFlagsModified() {
	}

	if s.IsAttributesModified() {
		s.attributes.markUnmodifiedRecursively()
	}

	if s.IsDroppedAttributesCountModified() {
	}

	s.modifiedFields.mask = 0
}

// markDiffModified marks fields in this struct modified if they differ from
// the corresponding fields in v.
func (s *Link) markDiffModified(v *Link) (modified bool) {
	if !pkg.BytesEqual(s.traceID, v.traceID) {
		s.markTraceIDModified()
		modified = true
	}

	if !pkg.BytesEqual(s.spanID, v.spanID) {
		s.markSpanIDModified()
		modified = true
	}

	if !pkg.StringEqual(s.traceState, v.traceState) {
		s.markTraceStateModified()
		modified = true
	}

	if !pkg.Uint64Equal(s.flags, v.flags) {
		s.markFlagsModified()
		modified = true
	}

	if s.attributes.markDiffModified(&v.attributes) {
		s.modifiedFields.markModified(fieldModifiedLinkAttributes)
		modified = true
	}

	if !pkg.Uint64Equal(s.droppedAttributesCount, v.droppedAttributesCount) {
		s.markDroppedAttributesCountModified()
		modified = true
	}

	return modified
}

func (s *Link) Clone() Link {
	return Link{
		traceID:                s.traceID,
		spanID:                 s.spanID,
		traceState:             s.traceState,
		flags:                  s.flags,
		attributes:             s.attributes.Clone(),
		droppedAttributesCount: s.droppedAttributesCount,
	}
}

// ByteSize returns approximate memory usage in bytes. Used to calculate
// memory used by dictionaries.
func (s *Link) byteSize() uint {
	return uint(unsafe.Sizeof(*s)) +
		s.attributes.byteSize() + 0
}

func copyLink(dst *Link, src *Link) {
	dst.SetTraceID(src.traceID)
	dst.SetSpanID(src.spanID)
	dst.SetTraceState(src.traceState)
	dst.SetFlags(src.flags)
	copyAttributes(&dst.attributes, &src.attributes)
	dst.SetDroppedAttributesCount(src.droppedAttributesCount)
}

// CopyFrom() performs a deep copy from src.
func (s *Link) CopyFrom(src *Link) {
	copyLink(s, src)
}

func (s *Link) markParentModified() {
	s.modifiedFields.parent.markModified(s.modifiedFields.parentBit)
}

func (s *Link) markUnmodified() {
	s.modifiedFields.markUnmodified()
	s.attributes.markUnmodified()
}

// mutateRandom mutates fields in a random, deterministic manner using
// random parameter as a deterministic generator.
func (s *Link) mutateRandom(random *rand.Rand) {
	const fieldCount = 6
	if random.IntN(fieldCount) == 0 {
		s.SetTraceID(pkg.BytesRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetSpanID(pkg.BytesRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetTraceState(pkg.StringRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetFlags(pkg.Uint64Random(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.attributes.mutateRandom(random)
	}
	if random.IntN(fieldCount) == 0 {
		s.SetDroppedAttributesCount(pkg.Uint64Random(random))
	}
}

// IsEqual performs deep comparison and returns true if struct is equal to val.
func (e *Link) IsEqual(val *Link) bool {
	if !pkg.BytesEqual(e.traceID, val.traceID) {
		return false
	}
	if !pkg.BytesEqual(e.spanID, val.spanID) {
		return false
	}
	if !pkg.StringEqual(e.traceState, val.traceState) {
		return false
	}
	if !pkg.Uint64Equal(e.flags, val.flags) {
		return false
	}
	if !e.attributes.IsEqual(&val.attributes) {
		return false
	}
	if !pkg.Uint64Equal(e.droppedAttributesCount, val.droppedAttributesCount) {
		return false
	}

	return true
}

func LinkEqual(left, right *Link) bool {
	return left.IsEqual(right)
}

// CmpLink performs deep comparison and returns an integer that
// will be 0 if left == right, negative if left < right, positive if left > right.
func CmpLink(left, right *Link) int {
	if left == nil {
		if right == nil {
			return 0
		}
		return -1
	}
	if right == nil {
		return 1
	}

	if c := pkg.BytesCompare(left.traceID, right.traceID); c != 0 {
		return c
	}
	if c := pkg.BytesCompare(left.spanID, right.spanID); c != 0 {
		return c
	}
	if c := strings.Compare(left.traceState, right.traceState); c != 0 {
		return c
	}
	if c := pkg.Uint64Compare(left.flags, right.flags); c != 0 {
		return c
	}
	if c := CmpAttributes(&left.attributes, &right.attributes); c != 0 {
		return c
	}
	if c := pkg.Uint64Compare(left.droppedAttributesCount, right.droppedAttributesCount); c != 0 {
		return c
	}

	return 0
}

// LinkEncoder implements encoding of Link
type LinkEncoder struct {
	buf     pkg.BitsWriter
	limiter *pkg.SizeLimiter

	// forceModifiedFields is set to true if the next encoding operation
	// must write all fields, whether they are modified or no.
	// This is used after frame restarts so that the data can be decoded
	// from the frame start.
	forceModifiedFields bool

	traceIDEncoder                encoders.BytesEncoder
	spanIDEncoder                 encoders.BytesEncoder
	traceStateEncoder             encoders.StringEncoder
	flagsEncoder                  encoders.Uint64Encoder
	attributesEncoder             AttributesEncoder
	droppedAttributesCountEncoder encoders.Uint64Encoder

	keepFieldMask uint64
	fieldCount    uint
}

func (e *LinkEncoder) Init(state *WriterState, columns *pkg.WriteColumnSet) error {
	// Remember this encoder in the state so that we can detect recursion.
	if state.LinkEncoder != nil {
		panic("cannot initialize LinkEncoder: already initialized")
	}
	state.LinkEncoder = e
	defer func() { state.LinkEncoder = nil }()

	e.limiter = &state.limiter

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("Link")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "Link")
		}

		// Number of fields in the target schema.
		e.fieldCount = fieldCount

		// Set that many 1 bits in the keepFieldMask. All fields with higher number
		// will be skipped when encoding.
		e.keepFieldMask = ^(^uint64(0) << e.fieldCount)
	} else {
		// Keep all fields when encoding.
		e.fieldCount = 6
		e.keepFieldMask = ^uint64(0)
	}

	if e.fieldCount <= 0 {
		return nil // TraceID and subsequent fields are skipped.
	}
	if err := e.traceIDEncoder.Init(nil, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 1 {
		return nil // SpanID and subsequent fields are skipped.
	}
	if err := e.spanIDEncoder.Init(nil, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 2 {
		return nil // TraceState and subsequent fields are skipped.
	}
	if err := e.traceStateEncoder.Init(nil, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 3 {
		return nil // Flags and subsequent fields are skipped.
	}
	if err := e.flagsEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 4 {
		return nil // Attributes and subsequent fields are skipped.
	}
	if err := e.attributesEncoder.Init(state, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 5 {
		return nil // DroppedAttributesCount and subsequent fields are skipped.
	}
	if err := e.droppedAttributesCountEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}

	return nil
}

func (e *LinkEncoder) Reset() {
	// Since we are resetting the state of encoder make sure the next Encode()
	// call forcedly writes all fields and does not attempt to skip.
	e.forceModifiedFields = true
	e.traceIDEncoder.Reset()
	e.spanIDEncoder.Reset()
	e.traceStateEncoder.Reset()
	e.flagsEncoder.Reset()
	e.attributesEncoder.Reset()
	e.droppedAttributesCountEncoder.Reset()
}

// Encode encodes val into buf
func (e *LinkEncoder) Encode(val *Link) {
	var bitCount uint

	// Mask that describes what fields are encoded. Start with all modified fields.
	fieldMask := val.modifiedFields.mask

	// If forceModifiedFields we need to set to 1 all bits so that we
	// force writing of all fields.
	if e.forceModifiedFields {
		fieldMask =
			fieldModifiedLinkTraceID |
				fieldModifiedLinkSpanID |
				fieldModifiedLinkTraceState |
				fieldModifiedLinkFlags |
				fieldModifiedLinkAttributes |
				fieldModifiedLinkDroppedAttributesCount | 0
	}

	// Only write fields that we want to write. See Init() for keepFieldMask.
	fieldMask &= e.keepFieldMask

	// Write bits to indicate which fields follow.
	e.buf.WriteBits(fieldMask, e.fieldCount)
	bitCount += e.fieldCount

	// Encode modified, present fields.

	if fieldMask&fieldModifiedLinkTraceID != 0 {
		// Encode TraceID
		e.traceIDEncoder.Encode(val.traceID)
	}

	if fieldMask&fieldModifiedLinkSpanID != 0 {
		// Encode SpanID
		e.spanIDEncoder.Encode(val.spanID)
	}

	if fieldMask&fieldModifiedLinkTraceState != 0 {
		// Encode TraceState
		e.traceStateEncoder.Encode(val.traceState)
	}

	if fieldMask&fieldModifiedLinkFlags != 0 {
		// Encode Flags
		e.flagsEncoder.Encode(val.flags)
	}

	if fieldMask&fieldModifiedLinkAttributes != 0 {
		// Encode Attributes
		e.attributesEncoder.Encode(&val.attributes)
	}

	if fieldMask&fieldModifiedLinkDroppedAttributesCount != 0 {
		// Encode DroppedAttributesCount
		e.droppedAttributesCountEncoder.Encode(val.droppedAttributesCount)
	}

	// Account written bits in the limiter.
	e.limiter.AddFrameBits(bitCount)

	// Mark all fields non-modified so that next Encode() correctly
	// encodes only fields that change after this.
	val.modifiedFields.mask = 0
}

// CollectColumns collects all buffers from all encoders into buf.
func (e *LinkEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBits(&e.buf)

	if e.fieldCount <= 0 {
		return // TraceID and subsequent fields are skipped.
	}
	e.traceIDEncoder.CollectColumns(columnSet.At(0))
	if e.fieldCount <= 1 {
		return // SpanID and subsequent fields are skipped.
	}
	e.spanIDEncoder.CollectColumns(columnSet.At(1))
	if e.fieldCount <= 2 {
		return // TraceState and subsequent fields are skipped.
	}
	e.traceStateEncoder.CollectColumns(columnSet.At(2))
	if e.fieldCount <= 3 {
		return // Flags and subsequent fields are skipped.
	}
	e.flagsEncoder.CollectColumns(columnSet.At(3))
	if e.fieldCount <= 4 {
		return // Attributes and subsequent fields are skipped.
	}
	e.attributesEncoder.CollectColumns(columnSet.At(4))
	if e.fieldCount <= 5 {
		return // DroppedAttributesCount and subsequent fields are skipped.
	}
	e.droppedAttributesCountEncoder.CollectColumns(columnSet.At(5))
}

// LinkDecoder implements decoding of Link
type LinkDecoder struct {
	buf        pkg.BitsReader
	column     *pkg.ReadableColumn
	lastValPtr *Link
	lastVal    Link
	fieldCount uint

	traceIDDecoder                encoders.BytesDecoder
	spanIDDecoder                 encoders.BytesDecoder
	traceStateDecoder             encoders.StringDecoder
	flagsDecoder                  encoders.Uint64Decoder
	attributesDecoder             AttributesDecoder
	droppedAttributesCountDecoder encoders.Uint64Decoder
}

// Init is called once in the lifetime of the stream.
func (d *LinkDecoder) Init(state *ReaderState, columns *pkg.ReadColumnSet) error {
	// Remember this decoder in the state so that we can detect recursion.
	if state.LinkDecoder != nil {
		panic("cannot initialize LinkDecoder: already initialized")
	}
	state.LinkDecoder = d
	defer func() { state.LinkDecoder = nil }()

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("Link")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "Link")
		}

		// Number of fields in the target schema.
		d.fieldCount = fieldCount
	} else {
		// Keep all fields when encoding.
		d.fieldCount = 6
	}

	d.column = columns.Column()

	d.lastVal.init(nil, 0)
	d.lastValPtr = &d.lastVal

	var err error

	if d.fieldCount <= 0 {
		return nil // TraceID and subsequent fields are skipped.
	}
	err = d.traceIDDecoder.Init(nil, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 1 {
		return nil // SpanID and subsequent fields are skipped.
	}
	err = d.spanIDDecoder.Init(nil, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 2 {
		return nil // TraceState and subsequent fields are skipped.
	}
	err = d.traceStateDecoder.Init(nil, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 3 {
		return nil // Flags and subsequent fields are skipped.
	}
	err = d.flagsDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 4 {
		return nil // Attributes and subsequent fields are skipped.
	}
	err = d.attributesDecoder.Init(state, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 5 {
		return nil // DroppedAttributesCount and subsequent fields are skipped.
	}
	err = d.droppedAttributesCountDecoder.Init(columns.AddSubColumn())
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
func (d *LinkDecoder) Continue() {
	d.buf.Reset(d.column.Data())

	if d.fieldCount <= 0 {
		return // TraceID and subsequent fields are skipped.
	}
	d.traceIDDecoder.Continue()
	if d.fieldCount <= 1 {
		return // SpanID and subsequent fields are skipped.
	}
	d.spanIDDecoder.Continue()
	if d.fieldCount <= 2 {
		return // TraceState and subsequent fields are skipped.
	}
	d.traceStateDecoder.Continue()
	if d.fieldCount <= 3 {
		return // Flags and subsequent fields are skipped.
	}
	d.flagsDecoder.Continue()
	if d.fieldCount <= 4 {
		return // Attributes and subsequent fields are skipped.
	}
	d.attributesDecoder.Continue()
	if d.fieldCount <= 5 {
		return // DroppedAttributesCount and subsequent fields are skipped.
	}
	d.droppedAttributesCountDecoder.Continue()
}

func (d *LinkDecoder) Reset() {
	d.traceIDDecoder.Reset()
	d.spanIDDecoder.Reset()
	d.traceStateDecoder.Reset()
	d.flagsDecoder.Reset()
	d.attributesDecoder.Reset()
	d.droppedAttributesCountDecoder.Reset()
}

func (d *LinkDecoder) Decode(dstPtr *Link) error {
	val := dstPtr

	var err error

	// Read bits that indicate which fields follow.
	val.modifiedFields.mask = d.buf.ReadBits(d.fieldCount)

	if val.modifiedFields.mask&fieldModifiedLinkTraceID != 0 {
		// Field is changed and is present, decode it.
		err = d.traceIDDecoder.Decode(&val.traceID)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedLinkSpanID != 0 {
		// Field is changed and is present, decode it.
		err = d.spanIDDecoder.Decode(&val.spanID)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedLinkTraceState != 0 {
		// Field is changed and is present, decode it.
		err = d.traceStateDecoder.Decode(&val.traceState)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedLinkFlags != 0 {
		// Field is changed and is present, decode it.
		err = d.flagsDecoder.Decode(&val.flags)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedLinkAttributes != 0 {
		// Field is changed and is present, decode it.
		err = d.attributesDecoder.Decode(&val.attributes)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedLinkDroppedAttributesCount != 0 {
		// Field is changed and is present, decode it.
		err = d.droppedAttributesCountDecoder.Decode(&val.droppedAttributesCount)
		if err != nil {
			return err
		}
	}

	return nil
}
