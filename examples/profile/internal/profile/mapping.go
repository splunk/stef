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

type Mapping struct {
	memoryStart     uint64
	memoryLimit     uint64
	fileOffset      uint64
	filename        string
	buildId         string
	hasFunctions    bool
	hasFilenames    bool
	hasLineNumbers  bool
	hasInlineFrames bool

	// modifiedFields keeps track of which fields are modified.
	modifiedFields modifiedFields
}

const MappingStructName = "Mapping"

// Bitmasks for "modified" flags for each field.
const (
	fieldModifiedMappingMemoryStart = uint64(1 << iota)
	fieldModifiedMappingMemoryLimit
	fieldModifiedMappingFileOffset
	fieldModifiedMappingFilename
	fieldModifiedMappingBuildId
	fieldModifiedMappingHasFunctions
	fieldModifiedMappingHasFilenames
	fieldModifiedMappingHasLineNumbers
	fieldModifiedMappingHasInlineFrames
)

// Init must be called once, before the Mapping is used.
func (s *Mapping) Init() {
	s.init(nil, 0)
}

func NewMapping() *Mapping {
	var s Mapping
	s.init(nil, 0)
	return &s
}

func (s *Mapping) init(parentModifiedFields *modifiedFields, parentModifiedBit uint64) {
	s.modifiedFields.parent = parentModifiedFields
	s.modifiedFields.parentBit = parentModifiedBit

}

func (s *Mapping) MemoryStart() uint64 {
	return s.memoryStart
}

// SetMemoryStart sets the value of MemoryStart field.
func (s *Mapping) SetMemoryStart(v uint64) {
	if !pkg.Uint64Equal(s.memoryStart, v) {
		s.memoryStart = v
		s.markMemoryStartModified()
	}
}

func (s *Mapping) markMemoryStartModified() {
	s.modifiedFields.markModified(fieldModifiedMappingMemoryStart)
}

// IsMemoryStartModified returns true the value of MemoryStart field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsMemoryStartModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingMemoryStart != 0
}

func (s *Mapping) MemoryLimit() uint64 {
	return s.memoryLimit
}

// SetMemoryLimit sets the value of MemoryLimit field.
func (s *Mapping) SetMemoryLimit(v uint64) {
	if !pkg.Uint64Equal(s.memoryLimit, v) {
		s.memoryLimit = v
		s.markMemoryLimitModified()
	}
}

func (s *Mapping) markMemoryLimitModified() {
	s.modifiedFields.markModified(fieldModifiedMappingMemoryLimit)
}

// IsMemoryLimitModified returns true the value of MemoryLimit field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsMemoryLimitModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingMemoryLimit != 0
}

func (s *Mapping) FileOffset() uint64 {
	return s.fileOffset
}

// SetFileOffset sets the value of FileOffset field.
func (s *Mapping) SetFileOffset(v uint64) {
	if !pkg.Uint64Equal(s.fileOffset, v) {
		s.fileOffset = v
		s.markFileOffsetModified()
	}
}

func (s *Mapping) markFileOffsetModified() {
	s.modifiedFields.markModified(fieldModifiedMappingFileOffset)
}

// IsFileOffsetModified returns true the value of FileOffset field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsFileOffsetModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingFileOffset != 0
}

func (s *Mapping) Filename() string {
	return s.filename
}

// SetFilename sets the value of Filename field.
func (s *Mapping) SetFilename(v string) {
	if !pkg.StringEqual(s.filename, v) {
		s.filename = v
		s.markFilenameModified()
	}
}

func (s *Mapping) markFilenameModified() {
	s.modifiedFields.markModified(fieldModifiedMappingFilename)
}

// IsFilenameModified returns true the value of Filename field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsFilenameModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingFilename != 0
}

func (s *Mapping) BuildId() string {
	return s.buildId
}

// SetBuildId sets the value of BuildId field.
func (s *Mapping) SetBuildId(v string) {
	if !pkg.StringEqual(s.buildId, v) {
		s.buildId = v
		s.markBuildIdModified()
	}
}

func (s *Mapping) markBuildIdModified() {
	s.modifiedFields.markModified(fieldModifiedMappingBuildId)
}

// IsBuildIdModified returns true the value of BuildId field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsBuildIdModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingBuildId != 0
}

func (s *Mapping) HasFunctions() bool {
	return s.hasFunctions
}

// SetHasFunctions sets the value of HasFunctions field.
func (s *Mapping) SetHasFunctions(v bool) {
	if !pkg.BoolEqual(s.hasFunctions, v) {
		s.hasFunctions = v
		s.markHasFunctionsModified()
	}
}

func (s *Mapping) markHasFunctionsModified() {
	s.modifiedFields.markModified(fieldModifiedMappingHasFunctions)
}

// IsHasFunctionsModified returns true the value of HasFunctions field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsHasFunctionsModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingHasFunctions != 0
}

func (s *Mapping) HasFilenames() bool {
	return s.hasFilenames
}

// SetHasFilenames sets the value of HasFilenames field.
func (s *Mapping) SetHasFilenames(v bool) {
	if !pkg.BoolEqual(s.hasFilenames, v) {
		s.hasFilenames = v
		s.markHasFilenamesModified()
	}
}

func (s *Mapping) markHasFilenamesModified() {
	s.modifiedFields.markModified(fieldModifiedMappingHasFilenames)
}

// IsHasFilenamesModified returns true the value of HasFilenames field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsHasFilenamesModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingHasFilenames != 0
}

func (s *Mapping) HasLineNumbers() bool {
	return s.hasLineNumbers
}

// SetHasLineNumbers sets the value of HasLineNumbers field.
func (s *Mapping) SetHasLineNumbers(v bool) {
	if !pkg.BoolEqual(s.hasLineNumbers, v) {
		s.hasLineNumbers = v
		s.markHasLineNumbersModified()
	}
}

func (s *Mapping) markHasLineNumbersModified() {
	s.modifiedFields.markModified(fieldModifiedMappingHasLineNumbers)
}

// IsHasLineNumbersModified returns true the value of HasLineNumbers field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsHasLineNumbersModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingHasLineNumbers != 0
}

func (s *Mapping) HasInlineFrames() bool {
	return s.hasInlineFrames
}

// SetHasInlineFrames sets the value of HasInlineFrames field.
func (s *Mapping) SetHasInlineFrames(v bool) {
	if !pkg.BoolEqual(s.hasInlineFrames, v) {
		s.hasInlineFrames = v
		s.markHasInlineFramesModified()
	}
}

func (s *Mapping) markHasInlineFramesModified() {
	s.modifiedFields.markModified(fieldModifiedMappingHasInlineFrames)
}

// IsHasInlineFramesModified returns true the value of HasInlineFrames field was modified since
// Mapping was created, encoded or decoded. If the field is modified
// it will be encoded by the next Write() operation. If the field is decoded by the
// next Read() operation the modified flag will be set.
func (s *Mapping) IsHasInlineFramesModified() bool {
	return s.modifiedFields.mask&fieldModifiedMappingHasInlineFrames != 0
}

func (s *Mapping) markModifiedRecursively() {

	s.modifiedFields.mask =
		fieldModifiedMappingMemoryStart |
			fieldModifiedMappingMemoryLimit |
			fieldModifiedMappingFileOffset |
			fieldModifiedMappingFilename |
			fieldModifiedMappingBuildId |
			fieldModifiedMappingHasFunctions |
			fieldModifiedMappingHasFilenames |
			fieldModifiedMappingHasLineNumbers |
			fieldModifiedMappingHasInlineFrames | 0
}

func (s *Mapping) markUnmodifiedRecursively() {

	if s.IsMemoryStartModified() {
	}

	if s.IsMemoryLimitModified() {
	}

	if s.IsFileOffsetModified() {
	}

	if s.IsFilenameModified() {
	}

	if s.IsBuildIdModified() {
	}

	if s.IsHasFunctionsModified() {
	}

	if s.IsHasFilenamesModified() {
	}

	if s.IsHasLineNumbersModified() {
	}

	if s.IsHasInlineFramesModified() {
	}

	s.modifiedFields.mask = 0
}

// markDiffModified marks fields in this struct modified if they differ from
// the corresponding fields in v.
func (s *Mapping) markDiffModified(v *Mapping) (modified bool) {
	if !pkg.Uint64Equal(s.memoryStart, v.memoryStart) {
		s.markMemoryStartModified()
		modified = true
	}

	if !pkg.Uint64Equal(s.memoryLimit, v.memoryLimit) {
		s.markMemoryLimitModified()
		modified = true
	}

	if !pkg.Uint64Equal(s.fileOffset, v.fileOffset) {
		s.markFileOffsetModified()
		modified = true
	}

	if !pkg.StringEqual(s.filename, v.filename) {
		s.markFilenameModified()
		modified = true
	}

	if !pkg.StringEqual(s.buildId, v.buildId) {
		s.markBuildIdModified()
		modified = true
	}

	if !pkg.BoolEqual(s.hasFunctions, v.hasFunctions) {
		s.markHasFunctionsModified()
		modified = true
	}

	if !pkg.BoolEqual(s.hasFilenames, v.hasFilenames) {
		s.markHasFilenamesModified()
		modified = true
	}

	if !pkg.BoolEqual(s.hasLineNumbers, v.hasLineNumbers) {
		s.markHasLineNumbersModified()
		modified = true
	}

	if !pkg.BoolEqual(s.hasInlineFrames, v.hasInlineFrames) {
		s.markHasInlineFramesModified()
		modified = true
	}

	return modified
}

func (s *Mapping) Clone() *Mapping {
	return &Mapping{
		memoryStart:     s.memoryStart,
		memoryLimit:     s.memoryLimit,
		fileOffset:      s.fileOffset,
		filename:        s.filename,
		buildId:         s.buildId,
		hasFunctions:    s.hasFunctions,
		hasFilenames:    s.hasFilenames,
		hasLineNumbers:  s.hasLineNumbers,
		hasInlineFrames: s.hasInlineFrames,
	}
}

// ByteSize returns approximate memory usage in bytes. Used to calculate
// memory used by dictionaries.
func (s *Mapping) byteSize() uint {
	return uint(unsafe.Sizeof(*s)) +
		0
}

func copyMapping(dst *Mapping, src *Mapping) {
	dst.SetMemoryStart(src.memoryStart)
	dst.SetMemoryLimit(src.memoryLimit)
	dst.SetFileOffset(src.fileOffset)
	dst.SetFilename(src.filename)
	dst.SetBuildId(src.buildId)
	dst.SetHasFunctions(src.hasFunctions)
	dst.SetHasFilenames(src.hasFilenames)
	dst.SetHasLineNumbers(src.hasLineNumbers)
	dst.SetHasInlineFrames(src.hasInlineFrames)
}

// CopyFrom() performs a deep copy from src.
func (s *Mapping) CopyFrom(src *Mapping) {
	copyMapping(s, src)
}

func (s *Mapping) markParentModified() {
	s.modifiedFields.parent.markModified(s.modifiedFields.parentBit)
}

func (s *Mapping) markUnmodified() {
	s.modifiedFields.markUnmodified()
}

// mutateRandom mutates fields in a random, deterministic manner using
// random parameter as a deterministic generator.
func (s *Mapping) mutateRandom(random *rand.Rand) {
	const fieldCount = 9
	if random.IntN(fieldCount) == 0 {
		s.SetMemoryStart(pkg.Uint64Random(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetMemoryLimit(pkg.Uint64Random(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetFileOffset(pkg.Uint64Random(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetFilename(pkg.StringRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetBuildId(pkg.StringRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetHasFunctions(pkg.BoolRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetHasFilenames(pkg.BoolRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetHasLineNumbers(pkg.BoolRandom(random))
	}
	if random.IntN(fieldCount) == 0 {
		s.SetHasInlineFrames(pkg.BoolRandom(random))
	}
}

// IsEqual performs deep comparison and returns true if struct is equal to val.
func (e *Mapping) IsEqual(val *Mapping) bool {
	if !pkg.Uint64Equal(e.memoryStart, val.memoryStart) {
		return false
	}
	if !pkg.Uint64Equal(e.memoryLimit, val.memoryLimit) {
		return false
	}
	if !pkg.Uint64Equal(e.fileOffset, val.fileOffset) {
		return false
	}
	if !pkg.StringEqual(e.filename, val.filename) {
		return false
	}
	if !pkg.StringEqual(e.buildId, val.buildId) {
		return false
	}
	if !pkg.BoolEqual(e.hasFunctions, val.hasFunctions) {
		return false
	}
	if !pkg.BoolEqual(e.hasFilenames, val.hasFilenames) {
		return false
	}
	if !pkg.BoolEqual(e.hasLineNumbers, val.hasLineNumbers) {
		return false
	}
	if !pkg.BoolEqual(e.hasInlineFrames, val.hasInlineFrames) {
		return false
	}

	return true
}

func MappingEqual(left, right *Mapping) bool {
	return left.IsEqual(right)
}

// CmpMapping performs deep comparison and returns an integer that
// will be 0 if left == right, negative if left < right, positive if left > right.
func CmpMapping(left, right *Mapping) int {
	if left == nil {
		if right == nil {
			return 0
		}
		return -1
	}
	if right == nil {
		return 1
	}

	if c := pkg.Uint64Compare(left.memoryStart, right.memoryStart); c != 0 {
		return c
	}
	if c := pkg.Uint64Compare(left.memoryLimit, right.memoryLimit); c != 0 {
		return c
	}
	if c := pkg.Uint64Compare(left.fileOffset, right.fileOffset); c != 0 {
		return c
	}
	if c := strings.Compare(left.filename, right.filename); c != 0 {
		return c
	}
	if c := strings.Compare(left.buildId, right.buildId); c != 0 {
		return c
	}
	if c := pkg.BoolCompare(left.hasFunctions, right.hasFunctions); c != 0 {
		return c
	}
	if c := pkg.BoolCompare(left.hasFilenames, right.hasFilenames); c != 0 {
		return c
	}
	if c := pkg.BoolCompare(left.hasLineNumbers, right.hasLineNumbers); c != 0 {
		return c
	}
	if c := pkg.BoolCompare(left.hasInlineFrames, right.hasInlineFrames); c != 0 {
		return c
	}

	return 0
}

// MappingEncoder implements encoding of Mapping
type MappingEncoder struct {
	buf     pkg.BitsWriter
	limiter *pkg.SizeLimiter

	// forceModifiedFields is set to true if the next encoding operation
	// must write all fields, whether they are modified or no.
	// This is used after frame restarts so that the data can be decoded
	// from the frame start.
	forceModifiedFields bool

	memoryStartEncoder     encoders.Uint64Encoder
	memoryLimitEncoder     encoders.Uint64Encoder
	fileOffsetEncoder      encoders.Uint64Encoder
	filenameEncoder        encoders.StringEncoder
	buildIdEncoder         encoders.StringEncoder
	hasFunctionsEncoder    encoders.BoolEncoder
	hasFilenamesEncoder    encoders.BoolEncoder
	hasLineNumbersEncoder  encoders.BoolEncoder
	hasInlineFramesEncoder encoders.BoolEncoder

	dict *MappingEncoderDict

	keepFieldMask uint64
	fieldCount    uint
}

type MappingEntry struct {
	refNum uint64
	val    *Mapping
}

// MappingEncoderDict is the dictionary used by MappingEncoder
type MappingEncoderDict struct {
	dict    b.Tree[*Mapping, MappingEntry]
	limiter *pkg.SizeLimiter
}

func (d *MappingEncoderDict) Init(limiter *pkg.SizeLimiter) {
	d.dict = *b.TreeNew[*Mapping, MappingEntry](CmpMapping)
	d.dict.Set(nil, MappingEntry{}) // nil Mapping is RefNum 0
	d.limiter = limiter
}

func (d *MappingEncoderDict) Reset() {
	d.dict.Clear()
	d.dict.Set(nil, MappingEntry{}) // nil Mapping is RefNum 0
}

func (e *MappingEncoder) Init(state *WriterState, columns *pkg.WriteColumnSet) error {
	// Remember this encoder in the state so that we can detect recursion.
	if state.MappingEncoder != nil {
		panic("cannot initialize MappingEncoder: already initialized")
	}
	state.MappingEncoder = e
	defer func() { state.MappingEncoder = nil }()

	e.limiter = &state.limiter
	e.dict = &state.Mapping

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("Mapping")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "Mapping")
		}

		// Number of fields in the target schema.
		e.fieldCount = fieldCount

		// Set that many 1 bits in the keepFieldMask. All fields with higher number
		// will be skipped when encoding.
		e.keepFieldMask = ^(^uint64(0) << e.fieldCount)
	} else {
		// Keep all fields when encoding.
		e.fieldCount = 9
		e.keepFieldMask = ^uint64(0)
	}

	if e.fieldCount <= 0 {
		return nil // MemoryStart and subsequent fields are skipped.
	}
	if err := e.memoryStartEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 1 {
		return nil // MemoryLimit and subsequent fields are skipped.
	}
	if err := e.memoryLimitEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 2 {
		return nil // FileOffset and subsequent fields are skipped.
	}
	if err := e.fileOffsetEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 3 {
		return nil // Filename and subsequent fields are skipped.
	}
	if err := e.filenameEncoder.Init(&state.Filename, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 4 {
		return nil // BuildId and subsequent fields are skipped.
	}
	if err := e.buildIdEncoder.Init(&state.BuildID, e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 5 {
		return nil // HasFunctions and subsequent fields are skipped.
	}
	if err := e.hasFunctionsEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 6 {
		return nil // HasFilenames and subsequent fields are skipped.
	}
	if err := e.hasFilenamesEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 7 {
		return nil // HasLineNumbers and subsequent fields are skipped.
	}
	if err := e.hasLineNumbersEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}
	if e.fieldCount <= 8 {
		return nil // HasInlineFrames and subsequent fields are skipped.
	}
	if err := e.hasInlineFramesEncoder.Init(e.limiter, columns.AddSubColumn()); err != nil {
		return err
	}

	return nil
}

func (e *MappingEncoder) Reset() {
	// Since we are resetting the state of encoder make sure the next Encode()
	// call forcedly writes all fields and does not attempt to skip.
	e.forceModifiedFields = true
	e.memoryStartEncoder.Reset()
	e.memoryLimitEncoder.Reset()
	e.fileOffsetEncoder.Reset()
	e.filenameEncoder.Reset()
	e.buildIdEncoder.Reset()
	e.hasFunctionsEncoder.Reset()
	e.hasFilenamesEncoder.Reset()
	e.hasLineNumbersEncoder.Reset()
	e.hasInlineFramesEncoder.Reset()
}

// Encode encodes val into buf
func (e *MappingEncoder) Encode(val *Mapping) {
	var bitCount uint

	// Check if the Mapping exists in the dictionary.
	entry, exists := e.dict.dict.Get(val)
	if exists {
		// The Mapping exists, we will reference it.
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

	// The Mapping does not exist in the dictionary. Add it to the dictionary.
	valInDict := val.Clone()
	entry = MappingEntry{refNum: uint64(e.dict.dict.Len()), val: valInDict}
	e.dict.dict.Set(valInDict, entry)
	e.dict.limiter.AddDictElemSize(valInDict.byteSize())

	// Indicate that an encoded Mapping follows.
	e.buf.WriteBit(1)
	bitCount += 1
	// TODO: optimize and merge WriteBit with the following WriteBits.
	// Mask that describes what fields are encoded. Start with all modified fields.
	fieldMask := val.modifiedFields.mask

	// If forceModifiedFields we need to set to 1 all bits so that we
	// force writing of all fields.
	if e.forceModifiedFields {
		fieldMask =
			fieldModifiedMappingMemoryStart |
				fieldModifiedMappingMemoryLimit |
				fieldModifiedMappingFileOffset |
				fieldModifiedMappingFilename |
				fieldModifiedMappingBuildId |
				fieldModifiedMappingHasFunctions |
				fieldModifiedMappingHasFilenames |
				fieldModifiedMappingHasLineNumbers |
				fieldModifiedMappingHasInlineFrames | 0
	}

	// Only write fields that we want to write. See Init() for keepFieldMask.
	fieldMask &= e.keepFieldMask

	// Write bits to indicate which fields follow.
	e.buf.WriteBits(fieldMask, e.fieldCount)
	bitCount += e.fieldCount

	// Encode modified, present fields.

	if fieldMask&fieldModifiedMappingMemoryStart != 0 {
		// Encode MemoryStart
		e.memoryStartEncoder.Encode(val.memoryStart)
	}

	if fieldMask&fieldModifiedMappingMemoryLimit != 0 {
		// Encode MemoryLimit
		e.memoryLimitEncoder.Encode(val.memoryLimit)
	}

	if fieldMask&fieldModifiedMappingFileOffset != 0 {
		// Encode FileOffset
		e.fileOffsetEncoder.Encode(val.fileOffset)
	}

	if fieldMask&fieldModifiedMappingFilename != 0 {
		// Encode Filename
		e.filenameEncoder.Encode(val.filename)
	}

	if fieldMask&fieldModifiedMappingBuildId != 0 {
		// Encode BuildId
		e.buildIdEncoder.Encode(val.buildId)
	}

	if fieldMask&fieldModifiedMappingHasFunctions != 0 {
		// Encode HasFunctions
		e.hasFunctionsEncoder.Encode(val.hasFunctions)
	}

	if fieldMask&fieldModifiedMappingHasFilenames != 0 {
		// Encode HasFilenames
		e.hasFilenamesEncoder.Encode(val.hasFilenames)
	}

	if fieldMask&fieldModifiedMappingHasLineNumbers != 0 {
		// Encode HasLineNumbers
		e.hasLineNumbersEncoder.Encode(val.hasLineNumbers)
	}

	if fieldMask&fieldModifiedMappingHasInlineFrames != 0 {
		// Encode HasInlineFrames
		e.hasInlineFramesEncoder.Encode(val.hasInlineFrames)
	}

	// Account written bits in the limiter.
	e.limiter.AddFrameBits(bitCount)

	// Mark all fields non-modified so that next Encode() correctly
	// encodes only fields that change after this.
	val.modifiedFields.mask = 0
}

// CollectColumns collects all buffers from all encoders into buf.
func (e *MappingEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBits(&e.buf)

	if e.fieldCount <= 0 {
		return // MemoryStart and subsequent fields are skipped.
	}
	e.memoryStartEncoder.CollectColumns(columnSet.At(0))
	if e.fieldCount <= 1 {
		return // MemoryLimit and subsequent fields are skipped.
	}
	e.memoryLimitEncoder.CollectColumns(columnSet.At(1))
	if e.fieldCount <= 2 {
		return // FileOffset and subsequent fields are skipped.
	}
	e.fileOffsetEncoder.CollectColumns(columnSet.At(2))
	if e.fieldCount <= 3 {
		return // Filename and subsequent fields are skipped.
	}
	e.filenameEncoder.CollectColumns(columnSet.At(3))
	if e.fieldCount <= 4 {
		return // BuildId and subsequent fields are skipped.
	}
	e.buildIdEncoder.CollectColumns(columnSet.At(4))
	if e.fieldCount <= 5 {
		return // HasFunctions and subsequent fields are skipped.
	}
	e.hasFunctionsEncoder.CollectColumns(columnSet.At(5))
	if e.fieldCount <= 6 {
		return // HasFilenames and subsequent fields are skipped.
	}
	e.hasFilenamesEncoder.CollectColumns(columnSet.At(6))
	if e.fieldCount <= 7 {
		return // HasLineNumbers and subsequent fields are skipped.
	}
	e.hasLineNumbersEncoder.CollectColumns(columnSet.At(7))
	if e.fieldCount <= 8 {
		return // HasInlineFrames and subsequent fields are skipped.
	}
	e.hasInlineFramesEncoder.CollectColumns(columnSet.At(8))
}

// MappingDecoder implements decoding of Mapping
type MappingDecoder struct {
	buf        pkg.BitsReader
	column     *pkg.ReadableColumn
	lastValPtr *Mapping
	lastVal    Mapping
	fieldCount uint

	memoryStartDecoder     encoders.Uint64Decoder
	memoryLimitDecoder     encoders.Uint64Decoder
	fileOffsetDecoder      encoders.Uint64Decoder
	filenameDecoder        encoders.StringDecoder
	buildIdDecoder         encoders.StringDecoder
	hasFunctionsDecoder    encoders.BoolDecoder
	hasFilenamesDecoder    encoders.BoolDecoder
	hasLineNumbersDecoder  encoders.BoolDecoder
	hasInlineFramesDecoder encoders.BoolDecoder

	dict *MappingDecoderDict
}

// Init is called once in the lifetime of the stream.
func (d *MappingDecoder) Init(state *ReaderState, columns *pkg.ReadColumnSet) error {
	// Remember this decoder in the state so that we can detect recursion.
	if state.MappingDecoder != nil {
		panic("cannot initialize MappingDecoder: already initialized")
	}
	state.MappingDecoder = d
	defer func() { state.MappingDecoder = nil }()

	if state.OverrideSchema != nil {
		fieldCount, ok := state.OverrideSchema.FieldCount("Mapping")
		if !ok {
			return fmt.Errorf("cannot find struct in override schema: %s", "Mapping")
		}

		// Number of fields in the target schema.
		d.fieldCount = fieldCount
	} else {
		// Keep all fields when encoding.
		d.fieldCount = 9
	}

	d.column = columns.Column()

	d.lastVal.init(nil, 0)
	d.lastValPtr = &d.lastVal
	d.dict = &state.Mapping

	var err error

	if d.fieldCount <= 0 {
		return nil // MemoryStart and subsequent fields are skipped.
	}
	err = d.memoryStartDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 1 {
		return nil // MemoryLimit and subsequent fields are skipped.
	}
	err = d.memoryLimitDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 2 {
		return nil // FileOffset and subsequent fields are skipped.
	}
	err = d.fileOffsetDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 3 {
		return nil // Filename and subsequent fields are skipped.
	}
	err = d.filenameDecoder.Init(&state.Filename, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 4 {
		return nil // BuildId and subsequent fields are skipped.
	}
	err = d.buildIdDecoder.Init(&state.BuildID, columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 5 {
		return nil // HasFunctions and subsequent fields are skipped.
	}
	err = d.hasFunctionsDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 6 {
		return nil // HasFilenames and subsequent fields are skipped.
	}
	err = d.hasFilenamesDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 7 {
		return nil // HasLineNumbers and subsequent fields are skipped.
	}
	err = d.hasLineNumbersDecoder.Init(columns.AddSubColumn())
	if err != nil {
		return err
	}
	if d.fieldCount <= 8 {
		return nil // HasInlineFrames and subsequent fields are skipped.
	}
	err = d.hasInlineFramesDecoder.Init(columns.AddSubColumn())
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
func (d *MappingDecoder) Continue() {
	d.buf.Reset(d.column.Data())

	if d.fieldCount <= 0 {
		return // MemoryStart and subsequent fields are skipped.
	}
	d.memoryStartDecoder.Continue()
	if d.fieldCount <= 1 {
		return // MemoryLimit and subsequent fields are skipped.
	}
	d.memoryLimitDecoder.Continue()
	if d.fieldCount <= 2 {
		return // FileOffset and subsequent fields are skipped.
	}
	d.fileOffsetDecoder.Continue()
	if d.fieldCount <= 3 {
		return // Filename and subsequent fields are skipped.
	}
	d.filenameDecoder.Continue()
	if d.fieldCount <= 4 {
		return // BuildId and subsequent fields are skipped.
	}
	d.buildIdDecoder.Continue()
	if d.fieldCount <= 5 {
		return // HasFunctions and subsequent fields are skipped.
	}
	d.hasFunctionsDecoder.Continue()
	if d.fieldCount <= 6 {
		return // HasFilenames and subsequent fields are skipped.
	}
	d.hasFilenamesDecoder.Continue()
	if d.fieldCount <= 7 {
		return // HasLineNumbers and subsequent fields are skipped.
	}
	d.hasLineNumbersDecoder.Continue()
	if d.fieldCount <= 8 {
		return // HasInlineFrames and subsequent fields are skipped.
	}
	d.hasInlineFramesDecoder.Continue()
}

func (d *MappingDecoder) Reset() {
	d.memoryStartDecoder.Reset()
	d.memoryLimitDecoder.Reset()
	d.fileOffsetDecoder.Reset()
	d.filenameDecoder.Reset()
	d.buildIdDecoder.Reset()
	d.hasFunctionsDecoder.Reset()
	d.hasFilenamesDecoder.Reset()
	d.hasLineNumbersDecoder.Reset()
	d.hasInlineFramesDecoder.Reset()
}

func (d *MappingDecoder) Decode(dstPtr **Mapping) error {
	// Check if the Mapping exists in the dictionary.
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

	if val.modifiedFields.mask&fieldModifiedMappingMemoryStart != 0 {
		// Field is changed and is present, decode it.
		err = d.memoryStartDecoder.Decode(&val.memoryStart)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingMemoryLimit != 0 {
		// Field is changed and is present, decode it.
		err = d.memoryLimitDecoder.Decode(&val.memoryLimit)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingFileOffset != 0 {
		// Field is changed and is present, decode it.
		err = d.fileOffsetDecoder.Decode(&val.fileOffset)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingFilename != 0 {
		// Field is changed and is present, decode it.
		err = d.filenameDecoder.Decode(&val.filename)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingBuildId != 0 {
		// Field is changed and is present, decode it.
		err = d.buildIdDecoder.Decode(&val.buildId)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingHasFunctions != 0 {
		// Field is changed and is present, decode it.
		err = d.hasFunctionsDecoder.Decode(&val.hasFunctions)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingHasFilenames != 0 {
		// Field is changed and is present, decode it.
		err = d.hasFilenamesDecoder.Decode(&val.hasFilenames)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingHasLineNumbers != 0 {
		// Field is changed and is present, decode it.
		err = d.hasLineNumbersDecoder.Decode(&val.hasLineNumbers)
		if err != nil {
			return err
		}
	}

	if val.modifiedFields.mask&fieldModifiedMappingHasInlineFrames != 0 {
		// Field is changed and is present, decode it.
		err = d.hasInlineFramesDecoder.Decode(&val.hasInlineFrames)
		if err != nil {
			return err
		}
	}

	d.dict.dict = append(d.dict.dict, val)

	return nil
}

// MappingDecoderDict is the dictionary used by MappingDecoder
type MappingDecoderDict struct {
	dict []*Mapping
}

func (d *MappingDecoderDict) Init() {
	d.dict = d.dict[:0]
	d.dict = append(d.dict, nil) // nil Mapping is RefNum 0
}

// Reset the dictionary to initial state. Used when a frame is
// started with RestartDictionaries flag.
func (d *MappingDecoderDict) Reset() {
	d.Init()
}
