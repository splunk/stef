package main

import (
	"bytes"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg"
	"github.com/splunk/stef/go/pkg/schema"

	"github.com/splunk/stef/go/otel/oteltef"
)

type countingChunkWriter struct {
	pkg.MemChunkWriter
	chunkCount int
}

func (m *countingChunkWriter) WriteChunk(header []byte, content []byte) error {
	m.chunkCount++
	return m.MemChunkWriter.WriteChunk(header, content)
}

func TestWriterDictLimit(t *testing.T) {
	buf := &countingChunkWriter{}
	w, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{MaxTotalDictSize: 2000})
	require.NoError(t, err)

	// Fixed header is chunk 1, var header is chunk 2.
	assert.EqualValues(t, 2, buf.chunkCount)

	w.Record.Metric().SetName("cpu.utilization")

	err = w.Write()
	require.NoError(t, err)

	// Metric too small to trigger a new frame, so it is still chunk 2.
	assert.EqualValues(t, 2, buf.chunkCount)

	schemaUrl1 := strings.Repeat("s", 2000)
	w.Record.Resource().SetSchemaURL(schemaUrl1)
	err = w.Write()
	require.NoError(t, err)

	// Resource was large enough to trigger dictionary limit and result in
	// a new frame, so it is chunk 2.
	assert.EqualValues(t, 3, buf.chunkCount)

	schemaUrl2 := "small"
	w.Record.Resource().SetSchemaURL(schemaUrl2)
	err = w.Write()
	require.NoError(t, err)

	// Record is not large enough to trigger a new frame.
	assert.EqualValues(t, 3, buf.chunkCount)

	// Write the same record again to use dictionary encoding.
	err = w.Write()
	require.NoError(t, err)

	err = w.Flush()
	require.NoError(t, err)

	// Flush must trigger a new frame, so it is chunk 3.
	assert.EqualValues(t, 4, buf.chunkCount)

	reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
	require.NoError(t, err)

	readRecord, err := reader.Read()
	require.NoError(t, err)
	require.NotNil(t, readRecord)
	assert.EqualValues(t, "cpu.utilization", readRecord.Metric().Name())

	readRecord, err = reader.Read()
	require.NoError(t, err)
	require.NotNil(t, readRecord)
	assert.EqualValues(t, schemaUrl1, readRecord.Resource().SchemaURL())

	readRecord, err = reader.Read()
	require.NoError(t, err)
	require.NotNil(t, readRecord)
	assert.EqualValues(t, schemaUrl2, readRecord.Resource().SchemaURL())

	readRecord, err = reader.Read()
	require.NoError(t, err)
	require.NotNil(t, readRecord)
	assert.EqualValues(t, schemaUrl2, readRecord.Resource().SchemaURL())
}

func TestWriterFrameLimit(t *testing.T) {
	buf := &countingChunkWriter{}
	w, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{MaxUncompressedFrameByteSize: 2000})
	require.NoError(t, err)

	// Header is the first chunk.
	assert.EqualValues(t, 2, buf.chunkCount)

	w.Record.Metric().SetName("cpu.utilization")

	err = w.Write()
	require.NoError(t, err)

	// Metric too small to trigger a new frame, so it is still chunk 1.
	assert.EqualValues(t, 2, buf.chunkCount)

	w.Record.Resource().SetSchemaURL(strings.Repeat("s", 2000))
	err = w.Write()
	require.NoError(t, err)

	// Resource was large enough to trigger a new frame, so it is chunk 2.
	assert.EqualValues(t, 3, buf.chunkCount)

	w.Record.Resource().SetSchemaURL("small")
	err = w.Write()
	require.NoError(t, err)

	// Scope is not large enough to trigger a new frame.
	assert.EqualValues(t, 3, buf.chunkCount)

	err = w.Flush()
	require.NoError(t, err)

	// Flush must trigger a new frame, so it is chunk 3.
	assert.EqualValues(t, 4, buf.chunkCount)
}

func mapToTef(m map[string]any, out *oteltef.Attributes) {
	out.EnsureLen(len(m))
	i := 0
	for k, v := range m {
		valueToTef(v, out.At(i).Value())
		out.SetKey(i, k)
		i++
	}
	out.Sort()
}

func valueToTef(v any, into *oteltef.AnyValue) {
	if v == nil {
		into.SetType(oteltef.AnyValueTypeNone)
		return
	}

	switch val := v.(type) {
	case string:
		into.SetString(val)

	case bool:
		into.SetBool(val)

	case float64:
		into.SetFloat64(val)

	case int64:
		into.SetInt64(val)

	case []byte:
		into.SetBytes(pkg.Bytes(val))

	case []any:
		into.SetType(oteltef.AnyValueTypeArray)
		arr := into.Array()
		arr.EnsureLen(len(val))

		for i := 0; i < len(val); i++ {
			valueToTef(val[i], arr.At(i))
		}

	case map[string]any:
		into.SetType(oteltef.AnyValueTypeKVList)
		kvList := into.KVList()
		kvList.EnsureLen(len(val))

		i := 0
		for k, v := range val {
			kvList.SetKey(i, k)
			valueToTef(v, kvList.At(i).Value())
		}

	default:
		panic("unknown anyValue type")
	}
}

func tefToMap(in *oteltef.Attributes) map[string]any {
	out := map[string]any{}

	for i := 0; i < in.Len(); i++ {
		kv := in.At(i)
		val := tefToValue(kv.Value())
		out[kv.Key()] = val
	}
	return out
}

func tefToValue(src *oteltef.AnyValue) any {
	switch src.Type() {
	case oteltef.AnyValueTypeString:
		return src.String()

	case oteltef.AnyValueTypeBytes:
		return []byte(src.Bytes())

	case oteltef.AnyValueTypeInt64:
		return src.Int64()

	case oteltef.AnyValueTypeBool:
		return src.Bool()

	case oteltef.AnyValueTypeNone:
		return nil

	case oteltef.AnyValueTypeFloat64:
		return src.Float64()

	case oteltef.AnyValueTypeArray:
		values := []any{}
		arr := src.Array()
		for i := 0; i < arr.Len(); i++ {
			val := tefToValue(arr.At(i))
			values = append(values, val)
		}
		return values

	case oteltef.AnyValueTypeKVList:
		values := map[string]any{}
		kvList := src.KVList()
		for i := 0; i < kvList.Len(); i++ {
			pair := kvList.At(i)
			val := tefToValue(pair.Value())
			values[pair.Key()] = val
		}
		return values

	default:
		panic("unknown anyValue type")
	}
}

type attrGenerator struct {
	r *rand.Rand
}

func (g *attrGenerator) randKey() string {
	keys := []string{"", "abc", "def"}
	return keys[g.r.IntN(len(keys))]
}

func (g *attrGenerator) genAttr() map[string]any {
	m := map[string]any{}

	ln := g.r.IntN(2)
	for i := 0; i < ln; i++ {
		vals := []string{"", "foo", "bar"}
		var val any
		switch g.r.IntN(3) {
		case 0:
		case 1:
			val = vals[g.r.IntN(len(vals))]
		case 2:
			val = g.genAttr()
		}
		m[g.randKey()] = val
	}
	return m
}

func TestAnyValue(t *testing.T) {
	buf := &countingChunkWriter{}
	w, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)

	g := attrGenerator{r: rand.New(rand.NewPCG(0, 0))}
	var writeAttrs []map[string]any
	for i := 0; i < 10000; i++ {
		writeAttrs = append(writeAttrs, g.genAttr())
	}

	for i := 0; i < len(writeAttrs); i++ {
		mapToTef(writeAttrs[i], w.Record.Attributes())
		err = w.Write()
		require.NoError(t, err)
	}

	err = w.Flush()
	require.NoError(t, err)

	reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
	require.NoError(t, err)

	for i := 0; i < len(writeAttrs); i++ {
		readRecord, err := reader.Read()
		require.NoError(t, err, i)
		require.NotNil(t, readRecord, i)

		readAttr := tefToMap(readRecord.Attributes())
		require.EqualValues(t, writeAttrs[i], readAttr, i)
	}
}

func writeReadRecord(t *testing.T, withSchema *schema.WireSchema) *oteltef.Metrics {
	buf := &countingChunkWriter{}
	writer, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{Schema: withSchema})
	require.NoError(t, err)

	writer.Record.Metric().SetName("abc")
	writer.Record.Scope().SetName("scope")
	writer.Record.Point().SetTimestamp(123)
	writer.Record.Point().Value().SetFloat64(4.5)

	err = writer.Write()
	require.NoError(t, err)

	err = writer.Flush()
	require.NoError(t, err)

	reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
	require.NoError(t, err)

	readRecord, err := reader.Read()
	require.NoError(t, err)

	return readRecord
}

func TestWriteOverrideSchema(t *testing.T) {
	schem, err := oteltef.MetricsWireSchema()
	require.NoError(t, err)

	// Write/read using nil schema, which is equal to full schema
	readRecord := writeReadRecord(t, nil)
	assert.EqualValues(t, "abc", readRecord.Metric().Name())
	assert.EqualValues(t, "scope", readRecord.Scope().Name())
	assert.EqualValues(t, 123, readRecord.Point().Timestamp())
	assert.EqualValues(t, 4.5, readRecord.Point().Value().Float64())

	// Write/read using full, unmodified schema
	readRecord = writeReadRecord(t, &schem)
	assert.EqualValues(t, "abc", readRecord.Metric().Name())
	assert.EqualValues(t, "scope", readRecord.Scope().Name())
	assert.EqualValues(t, 123, readRecord.Point().Timestamp())
	assert.EqualValues(t, oteltef.PointValueTypeFloat64, readRecord.Point().Value().Type())
	assert.EqualValues(t, 4.5, readRecord.Point().Value().Float64())

	// Remove "Monotonic" field (field #8) from "Metric" struct in the schema.
	schem.StructFieldCount["Metric"] = 7

	// Remove "Float64" field (field #2) from "PointValue" oneof struct in the schema.
	schem.StructFieldCount["PointValue"] = 1

	// Write/read using reduced schema
	readRecord = writeReadRecord(t, &schem)
	assert.EqualValues(t, "abc", readRecord.Metric().Name())
	assert.EqualValues(t, "scope", readRecord.Scope().Name())
	assert.EqualValues(t, 123, readRecord.Point().Timestamp())

	// PointValue type is None since it was Float64 in the original and Float64 field is removed
	// from PointValue schema. Removed fields in oneof structs result in None type when decoding.
	assert.EqualValues(t, oteltef.PointValueTypeNone, readRecord.Point().Value().Type())

	// Float64 is 0.0 which is the default value.
	assert.EqualValues(t, 0.0, readRecord.Point().Value().Float64())

	// Remove the entire "Point" field (field #6) from "Record" struct in the schema.
	schem.StructFieldCount["Metrics"] = 5

	// Write/read using reduced schema
	readRecord = writeReadRecord(t, &schem)
	assert.EqualValues(t, "abc", readRecord.Metric().Name())
	assert.EqualValues(t, "scope", readRecord.Scope().Name())
	// All Point fields are default values because Point field was not encoded by Writer.
	assert.EqualValues(t, 0, readRecord.Point().Timestamp())
	assert.EqualValues(t, oteltef.PointValueTypeNone, readRecord.Point().Value().Type())
	assert.EqualValues(t, 0.0, readRecord.Point().Value().Float64())
}

func TestLargeMultimap(t *testing.T) {
	buf := &countingChunkWriter{}
	w, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)

	attrs := w.Record.Attributes()

	// Test multimap with more than 62 pairs. This is called a "large" multimap.
	const attrCount = 100
	attrs.EnsureLen(attrCount)
	for i := 0; i < attrCount; i++ {
		attrs.SetKey(i, strconv.Itoa(i))
		attrs.At(i).Value().SetInt64(int64(i))
	}
	attrs1Copy := attrs.Clone()
	err = w.Write()
	require.NoError(t, err)

	// Modify one key. This normally would result in differential encoding
	// but since the multimap is large it will use full encoding. This is
	// precisely the case that this test verifies.
	attrs.At(0).Value().SetString("abc")
	attrs2Copy := attrs.Clone()
	err = w.Write()
	require.NoError(t, err)

	err = w.Flush()
	require.NoError(t, err)

	reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
	require.NoError(t, err)

	readRecord, err := reader.Read()
	require.NoError(t, err)
	require.NotNil(t, readRecord)

	require.True(t, readRecord.Attributes().IsEqual(&attrs1Copy))

	readRecord, err = reader.Read()
	require.NoError(t, err)
	require.NotNil(t, readRecord)

	require.True(t, readRecord.Attributes().IsEqual(&attrs2Copy))
}
