// Code generated by stefgen. DO NOT EDIT.
package profile

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

type Sample struct {
	metadata  ProfileMetadata
	locations LocationArray
	values    SampleValueArray
	labels    Labels

	// modifiedFields keeps track of which fields are modified.
	modifiedFields modifiedFields
}

const SampleStructName = "Sample"

// Bitmasks for "modified" flags for each field.
const (
	fieldModifiedSampleMetadata = uint64(1 << iota)
	fieldModifiedSampleLocations
	fieldModifiedSampleValues
	fieldModifiedSampleLabels
)

// Init must be called once, before the Sample is used.
func (s *Sample) Init() {
	s.init(nil, 0)
}

func NewSample() *Sample {
	var s Sample
	s.init(nil, 0)
	return &s
}

func (s *Sample) init(parentModifiedFields *modifiedFields, parentModifiedBit uint64) {
	s.modifiedFields.parent = parentModifiedFields
	s.modifiedFields.parentBit = parentModifiedBit

	s.metadata.init(&s.modifiedFields, fieldModifiedSampleMetadata)
	s.locations.init(&s.modifiedFields, fieldModifiedSampleLocations)
	s.values.init(&s.modifiedFields, fieldModifiedSampleValues)
	s.labels.init(&s.modifiedFields, fieldModifiedSampleLabels)
}

func (s *Sample) Metadata() *ProfileMetadata {
	return &s.metadata
}

// IsMetadataModified returns true the value of Metadata field was modified since
// Sample was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Sample) IsMetadataModified() bool {
	return s.modifiedFields.mask&fieldModifiedSampleMetadata != 0
}

func (s *Sample) Locations() *LocationArray {
	return &s.locations
}

// IsLocationsModified returns true the value of Locations field was modified since
// Sample was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Sample) IsLocationsModified() bool {
	return s.modifiedFields.mask&fieldModifiedSampleLocations != 0
}

func (s *Sample) Values() *SampleValueArray {
	return &s.values
}

// IsValuesModified returns true the value of Values field was modified since
// Sample was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Sample) IsValuesModified() bool {
	return s.modifiedFields.mask&fieldModifiedSampleValues != 0
}

func (s *Sample) Labels() *Labels {
	return &s.labels
}

// IsLabelsModified returns true the value of Labels field was modified since
// Sample was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Sample) IsLabelsModified() bool {
	return s.modifiedFields.mask&fieldModifiedSampleLabels != 0
}

func (s *Sample) markModifiedRecursively() {

	s.metadata.markModifiedRecursively()

	s.locations.markModifiedRecursively()

	s.values.markModifiedRecursively()

	s.labels.markModifiedRecursively()

	s.modifiedFields.mask =
		fieldModifiedSampleMetadata |
			fieldModifiedSampleLocations |
			fieldModifiedSampleValues |
			fieldModifiedSampleLabels | 0
}

func (s *Sample) markUnmodifiedRecursively() {

	if s.IsMetadataModified() {
		s.metadata.markUnmodifiedRecursively()
	}

	if s.IsLocationsModified() {
		s.locations.markUnmodifiedRecursively()
	}

	if s.IsValuesModified() {
		s.values.markUnmodifiedRecursively()
	}

	if s.IsLabelsModified() {
		s.labels.markUnmodifiedRecursively()
	}

	s.modifiedFields.mask = 0
}

// markDiffModified marks fields in this struct modified if they differ from
// the corresponding fields in v.
func (s *Sample) markDiffModified(v *Sample) (modified bool) {
	if s.metadata.markDiffModified(&v.metadata) {
		s.modifiedFields.markModified(fieldModifiedSampleMetadata)
		modified = true
	}

	if s.locations.markDiffModified(&v.locations) {
		s.modifiedFields.markModified(fieldModifiedSampleLocations)
		modified = true
	}

	if s.values.markDiffModified(&v.values) {
		s.modifiedFields.markModified(fieldModifiedSampleValues)
		modified = true
	}

	if s.labels.markDiffModified(&v.labels) {
		s.modifiedFields.markModified(fieldModifiedSampleLabels)
		modified = true
	}

	return modified
}

func (s *Sample) Clone() Sample {
	return Sample{
		metadata:  s.metadata.Clone(),
		locations: s.locations.Clone(),
		values:    s.values.Clone(),
		labels:    s.labels.Clone(),
	}
}

// ByteSize returns approximate memory usage in bytes. Used to calculate
// memory used by dictionaries.
func (s *Sample) byteSize() uint {
	return uint(unsafe.Sizeof(*s)) +
		s.metadata.byteSize() + s.locations.byteSize() + s.values.byteSize() + s.labels.byteSize() + 0
}

func copySample(dst *Sample, src *Sample) {
	copyProfileMetadata(&dst.metadata, &src.metadata)
	copyLocationArray(&dst.locations, &src.locations)
	copySampleValueArray(&dst.values, &src.values)
	copyLabels(&dst.labels, &src.labels)
}

// CopyFrom() performs a deep copy from src.
func (s *Sample) CopyFrom(src *Sample) {
	copySample(s, src)
}

func (s *Sample) markParentModified() {
	s.modifiedFields.parent.markModified(s.modifiedFields.parentBit)
}

func (s *Sample) markUnmodified() {
	s.modifiedFields.markUnmodified()
	s.metadata.markUnmodified()
	s.locations.markUnmodified()
	s.values.markUnmodified()
	s.labels.markUnmodified()
}

// mutateRandom mutates fields in a random, deterministic manner using
// random parameter as a deterministic generator.
func (s *Sample) mutateRandom(random *rand.Rand) {
	const fieldCount = 4
	if random.IntN(fieldCount) == 0 {
		s.metadata.mutateRandom(random)
	}
	if random.IntN(fieldCount) == 0 {
		s.locations.mutateRandom(random)
	}
	if random.IntN(fieldCount) == 0 {
		s.values.mutateRandom(random)
	}
	if random.IntN(fieldCount) == 0 {
		s.labels.mutateRandom(random)
	}
}

// IsEqual performs deep comparison and returns true if struct is equal to val.
func (e *Sample) IsEqual(val *Sample) bool {
	if !e.metadata.IsEqual(&val.metadata) {
		return false
	}
	if !e.locations.IsEqual(&val.locations) {
		return false
	}
	if !e.values.IsEqual(&val.values) {
		return false
	}
	if !e.labels.IsEqual(&val.labels) {
		return false
	}

	return true
}

func SampleEqual(left, right *Sample) bool {
	return left.IsEqual(right)
}

// CmpSample performs deep comparison and returns an integer that
// will be 0 if left == right, negative if left < right, positive if left > right.
func CmpSample(left, right *Sample) int {
	if left == nil {
		if right == nil {
			return 0
		}
		return -1
	}
	if right == nil {
		return 1
	}

	if c := CmpProfileMetadata(&left.metadata, &right.metadata); c != 0 {
		return c
	}
	if c := CmpLocationArray(&left.locations, &right.locations); c != 0 {
		return c
	}
	if c := CmpSampleValueArray(&left.values, &right.values); c != 0 {
		return c
	}
	if c := CmpLabels(&left.labels, &right.labels); c != 0 {
		return c
	}

	return 0
}

// SampleEncoder implements encoding of Sample
type SampleEncoder struct {
	buf     pkg.BitsWriter
	limiter *pkg.SizeLimiter

	// forceModifiedFields is set to true if the next encoding operation
	// must write all fields, whether they are modified or no.
	// This is used after frame restarts so that the data can be decoded
	// from the frame start.
	forceModifiedFields bool

	metadataEncoder  ProfileMetadataEncoder
	locationsEncoder LocationArrayEncoder
	valuesEncoder    SampleValueArrayEncoder
	labelsEncoder    LabelsEncoder

	keepFieldMask uint64
	fieldCount    uint
}

func (e *SampleEncoder) Init(state *WriterState, columns *pkg.WriteColumnSet) error {
	// Remember this encoder in the state so that we can detect recursion.
	if state.SampleEncoder != nil {
		panic("cannot initialize SampleEncoder: already initialized")
	}
	state.SampleEncoder = e
	defer func() { state.SampleEncoder = nil }()

	e.limiter = &state.limiter

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("Sample")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "Sample")
		}

		// Number of fields in the target schema.
		e.fieldCount = fieldCount

		// Set that many 1 bits in the keepFieldMask. All fields with higher number
		// will be skipped when encoding.
		e.keepFieldMask = ^(^uint64(0) << e.fieldCount)
	} else {
		// Keep all fields when encoding.
		e.fieldCount = 4
		e.keepFieldMask = ^uint64(0)
	}

	if e.fieldCount <= 0 {
		return nil // Metadata and subsequent fields are skipped.
	}
	if err := e.metadataEncoder.Init(state, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 1 {
		return nil // Locations and subsequent fields are skipped.
	}
	if err := e.locationsEncoder.Init(state, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 2 {
		return nil // Values and subsequent fields are skipped.
	}
	if err := e.valuesEncoder.Init(state, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 3 {
		return nil // Labels and subsequent fields are skipped.
	}
	if err := e.labelsEncoder.Init(state, columns.AddSubColumn()); err != nil {
		return err
	}

	return nil
}

func (e *SampleEncoder) Reset() {
	// Since we are resetting the state of encoder make sure the next Encode()
	// call forcedly writes all fields and does not attempt to skip.
	e.forceModifiedFields = true
	e.metadataEncoder.Reset()
	e.locationsEncoder.Reset()
	e.valuesEncoder.Reset()
	e.labelsEncoder.Reset()
}

// Encode encodes val into buf
func (e *SampleEncoder) Encode(val *Sample) {
	var bitCount uint

	// Mask that describes what fields are encoded. Start with all modified fields.
	fieldMask := val.modifiedFields.mask

	// If forceModifiedFields we need to set to 1 all bits so that we
	// force writing of all fields.
	if e.forceModifiedFields {
		fieldMask =
			fieldModifiedSampleMetadata |
				fieldModifiedSampleLocations |
				fieldModifiedSampleValues |
				fieldModifiedSampleLabels | 0
	}

	// Only write fields that we want to write. See Init() for keepFieldMask.
	fieldMask &= e.keepFieldMask

	// Write bits to indicate which fields follow.
	e.buf.WriteBits(fieldMask, e.fieldCount)
	bitCount += e.fieldCount

	// Encode modified, present fields.

	if fieldMask&fieldModifiedSampleMetadata != 0 {
		// Encode Metadata
		e.metadataEncoder.Encode(&val.metadata)
	}

	if fieldMask&fieldModifiedSampleLocations != 0 {
		// Encode Locations
		e.locationsEncoder.Encode(&val.locations)
	}

	if fieldMask&fieldModifiedSampleValues != 0 {
		// Encode Values
		e.valuesEncoder.Encode(&val.values)
	}

	if fieldMask&fieldModifiedSampleLabels != 0 {
		// Encode Labels
		e.labelsEncoder.Encode(&val.labels)
	}

	// Account written bits in the limiter.
	e.limiter.AddFrameBits(bitCount)

	// Mark all fields non-modified so that next Encode() correctly
	// encodes only fields that change after this.
	val.modifiedFields.mask = 0
}

// CollectColumns collects all buffers from all encoders into buf.
func (e *SampleEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBits(&e.buf)

	if e.fieldCount <= 0 {
		return // Metadata and subsequent fields are skipped.
	}
	e.metadataEncoder.CollectColumns(columnSet.At(0))
	if e.fieldCount <= 1 {
		return // Locations and subsequent fields are skipped.
	}
	e.locationsEncoder.CollectColumns(columnSet.At(1))
	if e.fieldCount <= 2 {
		return // Values and subsequent fields are skipped.
	}
	e.valuesEncoder.CollectColumns(columnSet.At(2))
	if e.fieldCount <= 3 {
		return // Labels and subsequent fields are skipped.
	}
	e.labelsEncoder.CollectColumns(columnSet.At(3))
}

// SampleDecoder implements decoding of Sample
type SampleDecoder struct {
	buf        pkg.BitsReader
	column     *pkg.ReadableColumn
	lastValPtr *Sample
	lastVal    Sample
	fieldCount uint

	metadataDecoder  ProfileMetadataDecoder
	locationsDecoder LocationArrayDecoder
	valuesDecoder    SampleValueArrayDecoder
	labelsDecoder    LabelsDecoder
}

// Init is called once in the lifetime of the stream.
func (d *SampleDecoder) Init(state *ReaderState, columns *pkg.ReadColumnSet) error {
	// Remember this decoder in the state so that we can detect recursion.
	if state.SampleDecoder != nil {
		panic("cannot initialize SampleDecoder: already initialized")
	}
	state.SampleDecoder = d
	defer func() { state.SampleDecoder = nil }()

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("Sample")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "Sample")
		}

		// Number of fields in the target schema.
		d.fieldCount = fieldCount
	} else {
		// Keep all fields when encoding.
		d.fieldCount = 4
	}

	d.column = columns.Column()

	d.lastVal.Init()
	d.lastValPtr = &d.lastVal

	var err error

	if d.fieldCount <= 0 {
		return nil // Metadata and subsequent fields are skipped.
	}
	err = d.metadataDecoder.Init(state, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 1 {
		return nil // Locations and subsequent fields are skipped.
	}
	err = d.locationsDecoder.Init(state, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 2 {
		return nil // Values and subsequent fields are skipped.
	}
	err = d.valuesDecoder.Init(state, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 3 {
		return nil // Labels and subsequent fields are skipped.
	}
	err = d.labelsDecoder.Init(state, columns.AddSubColumn())
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
func (d *SampleDecoder) Continue() {
	d.buf.Reset(d.column.Data())

	if d.fieldCount <= 0 {
		return // Metadata and subsequent fields are skipped.
	}
	d.metadataDecoder.Continue()
	if d.fieldCount <= 1 {
		return // Locations and subsequent fields are skipped.
	}
	d.locationsDecoder.Continue()
	if d.fieldCount <= 2 {
		return // Values and subsequent fields are skipped.
	}
	d.valuesDecoder.Continue()
	if d.fieldCount <= 3 {
		return // Labels and subsequent fields are skipped.
	}
	d.labelsDecoder.Continue()
}

func (d *SampleDecoder) Reset() {
	d.metadataDecoder.Reset()
	d.locationsDecoder.Reset()
	d.valuesDecoder.Reset()
	d.labelsDecoder.Reset()
}

func (d *SampleDecoder) Decode(dstPtr *Sample) error {
	val := dstPtr

	var err error

	// Read bits that indicate which fields follow.
	val.modifiedFields.mask = d.buf.ReadBits(d.fieldCount)

	if val.modifiedFields.mask&fieldModifiedSampleMetadata != 0 {
		// Field is changed and is present, decode it.
		err = d.metadataDecoder.Decode(&val.metadata)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedSampleLocations != 0 {
		// Field is changed and is present, decode it.
		err = d.locationsDecoder.Decode(&val.locations)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedSampleValues != 0 {
		// Field is changed and is present, decode it.
		err = d.valuesDecoder.Decode(&val.values)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedSampleLabels != 0 {
		// Field is changed and is present, decode it.
		err = d.labelsDecoder.Decode(&val.labels)
		if err != nil {
			return err
		}
	}

	return nil
}

var wireSchemaSample = []byte{0x0A, 0x08, 0x46, 0x75, 0x6E, 0x63, 0x74, 0x69, 0x6F, 0x6E, 0x04, 0x0A, 0x4C, 0x61, 0x62, 0x65, 0x6C, 0x56, 0x61, 0x6C, 0x75, 0x65, 0x02, 0x04, 0x4C, 0x69, 0x6E, 0x65, 0x03, 0x08, 0x4C, 0x6F, 0x63, 0x61, 0x74, 0x69, 0x6F, 0x6E, 0x04, 0x07, 0x4D, 0x61, 0x70, 0x70, 0x69, 0x6E, 0x67, 0x09, 0x08, 0x4E, 0x75, 0x6D, 0x56, 0x61, 0x6C, 0x75, 0x65, 0x02, 0x0F, 0x50, 0x72, 0x6F, 0x66, 0x69, 0x6C, 0x65, 0x4D, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x08, 0x06, 0x53, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x04, 0x0B, 0x53, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x56, 0x61, 0x6C, 0x75, 0x65, 0x02, 0x0F, 0x53, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x56, 0x61, 0x6C, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x02}

func SampleWireSchema() (schema.WireSchema, error) {
	var w schema.WireSchema
	if err := w.Deserialize(bytes.NewBuffer([]byte(wireSchemaSample))); err != nil {
		return w, err
	}
	return w, nil
}
