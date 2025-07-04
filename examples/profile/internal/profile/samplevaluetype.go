// Code generated by stefgen. DO NOT EDIT.
package profile

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"strings"
	"unsafe"

	"modernc.org/b/v2"

	"github.com/splunk/stef/go/pkg"
	"github.com/splunk/stef/go/pkg/encoders"
	"github.com/splunk/stef/go/pkg/schema"
)

var _ = strings.Compare
var _ = encoders.StringEncoder{}
var _ = schema.WireSchema{}
var _ = bytes.NewBuffer

type SampleValueType struct {
	type_ string
	unit  string

	// modifiedFields keeps track of which fields are modified.
	modifiedFields modifiedFields
}

const SampleValueTypeStructName = "SampleValueType"

// Bitmasks for "modified" flags for each field.
const (
	fieldModifiedSampleValueTypeType = uint64(1 << iota)
	fieldModifiedSampleValueTypeUnit
)

// Init must be called once, before the SampleValueType is used.
func (s *SampleValueType) Init() {
	s.init(nil, 0)
}

func NewSampleValueType() *SampleValueType {
	var s SampleValueType
	s.init(nil, 0)
	return &s
}

func (s *SampleValueType) init(parentModifiedFields *modifiedFields, parentModifiedBit uint64) {
	s.modifiedFields.parent = parentModifiedFields
	s.modifiedFields.parentBit = parentModifiedBit

}

func (s *SampleValueType) Type() string {
	return s.type_
}

// SetType sets the value of Type field.
func (s *SampleValueType) SetType(v string) {
	if !pkg.StringEqual(s.type_, v) {
		s.type_ = v
		s.markTypeModified()
	}
}

func (s *SampleValueType) markTypeModified() {
	s.modifiedFields.markModified(fieldModifiedSampleValueTypeType)
}

// IsTypeModified returns true the value of Type field was modified since
// SampleValueType was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *SampleValueType) IsTypeModified() bool {
	return s.modifiedFields.mask&fieldModifiedSampleValueTypeType != 0
}

func (s *SampleValueType) Unit() string {
	return s.unit
}

// SetUnit sets the value of Unit field.
func (s *SampleValueType) SetUnit(v string) {
	if !pkg.StringEqual(s.unit, v) {
		s.unit = v
		s.markUnitModified()
	}
}

func (s *SampleValueType) markUnitModified() {
	s.modifiedFields.markModified(fieldModifiedSampleValueTypeUnit)
}

// IsUnitModified returns true the value of Unit field was modified since
// SampleValueType was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *SampleValueType) IsUnitModified() bool {
	return s.modifiedFields.mask&fieldModifiedSampleValueTypeUnit != 0
}

func (s *SampleValueType) markModifiedRecursively() {

	s.modifiedFields.mask =
		fieldModifiedSampleValueTypeType |
			fieldModifiedSampleValueTypeUnit | 0
}

func (s *SampleValueType) markUnmodifiedRecursively() {

	if s.IsTypeModified() {
	}

	if s.IsUnitModified() {
	}

	s.modifiedFields.mask = 0
}

// markDiffModified marks fields in this struct modified if they differ from
// the corresponding fields in v.
func (s *SampleValueType) markDiffModified(v *SampleValueType) (modified bool) {
	if !pkg.StringEqual(s.type_, v.type_) {
		s.markTypeModified()
		modified = true
	}

	if !pkg.StringEqual(s.unit, v.unit) {
		s.markUnitModified()
		modified = true
	}

	return modified
}

func (s *SampleValueType) Clone() *SampleValueType {
	return &SampleValueType{
		type_: s.type_,
		unit:  s.unit,
	}
}

// ByteSize returns approximate memory usage in bytes. Used to calculate
// memory used by dictionaries.
func (s *SampleValueType) byteSize() uint {
	return uint(unsafe.Sizeof(*s)) +
		0
}

func copySampleValueType(dst *SampleValueType, src *SampleValueType) {
	dst.SetType(src.type_)
	dst.SetUnit(src.unit)
}

// CopyFrom() performs a deep copy from src.
func (s *SampleValueType) CopyFrom(src *SampleValueType) {
	copySampleValueType(s, src)
}

func (s *SampleValueType) markParentModified() {
	s.modifiedFields.parent.markModified(s.modifiedFields.parentBit)
}

func (s *SampleValueType) markUnmodified() {
	s.modifiedFields.markUnmodified()
}

// mutateRandom mutates fields in a random, deterministic manner using
// random parameter as a deterministic generator.
func (s *SampleValueType) mutateRandom(random *rand.Rand) {
	const fieldCount = 2
	if random.IntN(fieldCount) == 0 {
		s.SetType(pkg.StringRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetUnit(pkg.StringRandom(random))
	}
}

// IsEqual performs deep comparison and returns true if struct is equal to val.
func (e *SampleValueType) IsEqual(val *SampleValueType) bool {
	if !pkg.StringEqual(e.type_, val.type_) {
		return false
	}
	if !pkg.StringEqual(e.unit, val.unit) {
		return false
	}

	return true
}

func SampleValueTypeEqual(left, right *SampleValueType) bool {
	return left.IsEqual(right)
}

// CmpSampleValueType performs deep comparison and returns an integer that
// will be 0 if left == right, negative if left < right, positive if left > right.
func CmpSampleValueType(left, right *SampleValueType) int {
	if left == nil {
		if right == nil {
			return 0
		}
		return -1
	}
	if right == nil {
		return 1
	}

	if c := strings.Compare(left.type_, right.type_); c != 0 {
		return c
	}
	if c := strings.Compare(left.unit, right.unit); c != 0 {
		return c
	}

	return 0
}

// SampleValueTypeEncoder implements encoding of SampleValueType
type SampleValueTypeEncoder struct {
	buf     pkg.BitsWriter
	limiter *pkg.SizeLimiter

	// forceModifiedFields is set to true if the next encoding operation
	// must write all fields, whether they are modified or no.
	// This is used after frame restarts so that the data can be decoded
	// from the frame start.
	forceModifiedFields bool

	type_Encoder encoders.StringEncoder
	unitEncoder  encoders.StringEncoder

	dict *SampleValueTypeEncoderDict

	keepFieldMask uint64
	fieldCount    uint
}

type SampleValueTypeEntry struct {
	refNum uint64
	val    *SampleValueType
}

// SampleValueTypeEncoderDict is the dictionary used by SampleValueTypeEncoder
type SampleValueTypeEncoderDict struct {
	dict    b.Tree[*SampleValueType, SampleValueTypeEntry]
	limiter *pkg.SizeLimiter
}

func (d *SampleValueTypeEncoderDict) Init(limiter *pkg.SizeLimiter) {
	d.dict = *b.TreeNew[*SampleValueType, SampleValueTypeEntry](CmpSampleValueType)
	d.dict.Set(nil, SampleValueTypeEntry{}) // nil SampleValueType is RefNum 0
	d.limiter = limiter
}

func (d *SampleValueTypeEncoderDict) Reset() {
	d.dict.Clear()
	d.dict.Set(nil, SampleValueTypeEntry{}) // nil SampleValueType is RefNum 0
}

func (e *SampleValueTypeEncoder) Init(state *WriterState, columns *pkg.WriteColumnSet) error {
	// Remember this encoder in the state so that we can detect recursion.
	if state.SampleValueTypeEncoder != nil {
		panic("cannot initialize SampleValueTypeEncoder: already initialized")
	}
	state.SampleValueTypeEncoder = e
	defer func() { state.SampleValueTypeEncoder = nil }()

	e.limiter = &state.limiter
	e.dict = &state.SampleValueType

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("SampleValueType")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "SampleValueType")
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
		return nil // Type and subsequent fields are skipped.
	}
	if err := e.type_Encoder.Init(nil, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 1 {
		return nil // Unit and subsequent fields are skipped.
	}
	if err := e.unitEncoder.Init(nil, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}

	return nil
}

func (e *SampleValueTypeEncoder) Reset() {
	// Since we are resetting the state of encoder make sure the next Encode()
	// call forcedly writes all fields and does not attempt to skip.
	e.forceModifiedFields = true
	e.type_Encoder.Reset()
	e.unitEncoder.Reset()
}

// Encode encodes val into buf
func (e *SampleValueTypeEncoder) Encode(val *SampleValueType) {
	var bitCount uint

	// Check if the SampleValueType exists in the dictionary.
	entry, exists := e.dict.dict.Get(val)
	if exists {
		// The SampleValueType exists, we will reference it.
		// Indicate a RefNum follows.
		e.buf.WriteBit(0)
		// Encode refNum.
		bitCount = e.buf.WriteUvarintCompact(entry.refNum)

		// Account written bits in the limiter.
		e.limiter.AddFrameBits(1 + bitCount)

		// Mark all fields non-modified recursively so that next Encode() correctly
		// encodes only fields that change after this.
		val.markUnmodifiedRecursively()
		return
	}

	// The SampleValueType does not exist in the dictionary. Add it to the dictionary.
	valInDict := val.Clone()
	entry = SampleValueTypeEntry{refNum: uint64(e.dict.dict.Len()), val: valInDict}
	e.dict.dict.Set(valInDict, entry)
	e.dict.limiter.AddDictElemSize(valInDict.byteSize())

	// Indicate that an encoded SampleValueType follows.
	e.buf.WriteBit(1)
	bitCount += 1
	// TODO: optimize and merge WriteBit with the following WriteBits.
	// Mask that describes what fields are encoded. Start with all modified fields.
	fieldMask := val.modifiedFields.mask

	// If forceModifiedFields we need to set to 1 all bits so that we
	// force writing of all fields.
	if e.forceModifiedFields {
		fieldMask =
			fieldModifiedSampleValueTypeType |
				fieldModifiedSampleValueTypeUnit | 0
	}

	// Only write fields that we want to write. See Init() for keepFieldMask.
	fieldMask &= e.keepFieldMask

	// Write bits to indicate which fields follow.
	e.buf.WriteBits(fieldMask, e.fieldCount)
	bitCount += e.fieldCount

	// Encode modified, present fields.

	if fieldMask&fieldModifiedSampleValueTypeType != 0 {
		// Encode Type
		e.type_Encoder.Encode(val.type_)
	}

	if fieldMask&fieldModifiedSampleValueTypeUnit != 0 {
		// Encode Unit
		e.unitEncoder.Encode(val.unit)
	}

	// Account written bits in the limiter.
	e.limiter.AddFrameBits(bitCount)

	// Mark all fields non-modified so that next Encode() correctly
	// encodes only fields that change after this.
	val.modifiedFields.mask = 0
}

// CollectColumns collects all buffers from all encoders into buf.
func (e *SampleValueTypeEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBits(&e.buf)

	if e.fieldCount <= 0 {
		return // Type and subsequent fields are skipped.
	}
	e.type_Encoder.CollectColumns(columnSet.At(0))
	if e.fieldCount <= 1 {
		return // Unit and subsequent fields are skipped.
	}
	e.unitEncoder.CollectColumns(columnSet.At(1))
}

// SampleValueTypeDecoder implements decoding of SampleValueType
type SampleValueTypeDecoder struct {
	buf        pkg.BitsReader
	column     *pkg.ReadableColumn
	lastValPtr *SampleValueType
	lastVal    SampleValueType
	fieldCount uint

	type_Decoder encoders.StringDecoder
	unitDecoder  encoders.StringDecoder

	dict *SampleValueTypeDecoderDict
}

// Init is called once in the lifetime of the stream.
func (d *SampleValueTypeDecoder) Init(state *ReaderState, columns *pkg.ReadColumnSet) error {
	// Remember this decoder in the state so that we can detect recursion.
	if state.SampleValueTypeDecoder != nil {
		panic("cannot initialize SampleValueTypeDecoder: already initialized")
	}
	state.SampleValueTypeDecoder = d
	defer func() { state.SampleValueTypeDecoder = nil }()

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("SampleValueType")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "SampleValueType")
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
	d.dict = &state.SampleValueType

	var err error

	if d.fieldCount <= 0 {
		return nil // Type and subsequent fields are skipped.
	}
	err = d.type_Decoder.Init(nil, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 1 {
		return nil // Unit and subsequent fields are skipped.
	}
	err = d.unitDecoder.Init(nil, columns.AddSubColumn())
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
func (d *SampleValueTypeDecoder) Continue() {
	d.buf.Reset(d.column.Data())

	if d.fieldCount <= 0 {
		return // Type and subsequent fields are skipped.
	}
	d.type_Decoder.Continue()
	if d.fieldCount <= 1 {
		return // Unit and subsequent fields are skipped.
	}
	d.unitDecoder.Continue()
}

func (d *SampleValueTypeDecoder) Reset() {
	d.type_Decoder.Reset()
	d.unitDecoder.Reset()
}

func (d *SampleValueTypeDecoder) Decode(dstPtr **SampleValueType) error {
	// Check if the SampleValueType exists in the dictionary.
	dictFlag := d.buf.ReadBit()
	if dictFlag == 0 {
		refNum := d.buf.ReadUvarintCompact()
		if refNum >= uint64(len(d.dict.dict)) {
			return pkg.ErrInvalidRefNum
		}
		d.lastValPtr = d.dict.dict[refNum]
		*dstPtr = d.lastValPtr
		return nil
	}

	// lastValPtr here is pointing to a element in the dictionary. We are not allowed
	// to modify it. Make a clone of it and decode into the clone.
	val := d.lastValPtr.Clone()
	d.lastValPtr = val
	*dstPtr = val

	var err error

	// Read bits that indicate which fields follow.
	val.modifiedFields.mask = d.buf.ReadBits(d.fieldCount)

	if val.modifiedFields.mask&fieldModifiedSampleValueTypeType != 0 {
		// Field is changed and is present, decode it.
		err = d.type_Decoder.Decode(&val.type_)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedSampleValueTypeUnit != 0 {
		// Field is changed and is present, decode it.
		err = d.unitDecoder.Decode(&val.unit)
		if err != nil {
			return err
		}
	}

	d.dict.dict = append(d.dict.dict, val)

	return nil
}

// SampleValueTypeDecoderDict is the dictionary used by SampleValueTypeDecoder
type SampleValueTypeDecoderDict struct {
	dict []*SampleValueType
}

func (d *SampleValueTypeDecoderDict) Init() {
	d.dict = d.dict[:0]
	d.dict = append(d.dict, nil) // nil SampleValueType is RefNum 0
}

// Reset the dictionary to initial state. Used when a frame is
// started with RestartDictionaries flag.
func (d *SampleValueTypeDecoderDict) Reset() {
	d.Init()
}
