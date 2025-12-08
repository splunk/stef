package metrics

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/internal/pict"
	"github.com/splunk/stef/go/pdata/metrics/testtools"
)

func TestConvertFromToOTLP(t *testing.T) {
	otlpMetrics, err := pict.GenerateMetrics("testdata/generated_pict_pairs_metrics.txt")
	require.NoError(t, err)

	for _, otlpMetricSrc := range otlpMetrics {
		testtools.NormalizeMetrics(otlpMetricSrc)
		srcCount := otlpMetricSrc.DataPointCount()

		buf := &pkg.MemChunkWriter{}
		writer, err := otelstef.NewMetricsWriter(buf, pkg.WriterOptions{})
		require.NoError(t, err)

		toStef := OtlpToStefSorted{}
		err = toStef.Convert(otlpMetricSrc, writer)
		require.NoError(t, err)

		assert.EqualValues(t, srcCount, int(writer.RecordCount()))

		err = writer.Flush()
		require.NoError(t, err)

		reader, err := otelstef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
		require.NoError(t, err)

		toOtlp := StefToOtlpUnsorted{}
		otlpMetricCopy, err := toOtlp.Convert(reader, true)
		require.NoError(t, err)

		assert.EqualValues(t, writer.RecordCount(), reader.RecordCount())

		testtools.NormalizeMetrics(otlpMetricCopy)

		copyCount := otlpMetricCopy.DataPointCount()
		assert.EqualValues(t, srcCount, copyCount)

		assert.NoError(t, testtools.DiffMetrics(otlpMetricSrc, otlpMetricCopy))
		assert.True(t, bytes.Equal(toBytes(t, otlpMetricSrc), toBytes(t, otlpMetricCopy)))
	}
}

func toBytes(t *testing.T, data pmetric.Metrics) []byte {
	marshaler := pmetric.ProtoMarshaler{}
	bytes, err := marshaler.MarshalMetrics(data)
	require.NoError(t, err)
	return bytes
}

func TestFromOTLPToWriter(t *testing.T) {
	otlpMetrics, err := pict.GenerateMetrics("testdata/generated_pict_pairs_metrics.txt")
	require.NoError(t, err)

	for _, otlpMetricSrc := range otlpMetrics {
		testtools.NormalizeMetrics(otlpMetricSrc)
		srcCount := otlpMetricSrc.DataPointCount()

		buf := &pkg.MemChunkWriter{}
		writer, err := otelstef.NewMetricsWriter(buf, pkg.WriterOptions{})
		require.NoError(t, err)

		// Convert from OTLP to STEF
		converter := OtlpToStefUnsorted{}
		err = converter.Convert(otlpMetricSrc, writer)
		require.NoError(t, err)

		assert.EqualValues(t, srcCount, int(writer.RecordCount()))

		err = writer.Flush()
		require.NoError(t, err)

		reader, err := otelstef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
		require.NoError(t, err)

		toOtlp := StefToOtlpUnsorted{}
		otlpMetricCopy, err := toOtlp.Convert(reader, true)
		require.NoError(t, err)

		assert.EqualValues(t, writer.RecordCount(), reader.RecordCount())

		testtools.NormalizeMetrics(otlpMetricCopy)

		copyCount := otlpMetricCopy.DataPointCount()
		assert.EqualValues(t, srcCount, copyCount)

		assert.NoError(t, testtools.DiffMetrics(otlpMetricSrc, otlpMetricCopy))
		assert.True(t, bytes.Equal(toBytes(t, otlpMetricSrc), toBytes(t, otlpMetricCopy)))
	}
}

func FuzzReader(f *testing.F) {
	otlpMetrics, err := pict.GenerateMetrics("testdata/generated_pict_pairs_metrics.txt")
	require.NoError(f, err)

	for _, otlpMetricSrc := range otlpMetrics {
		buf := &pkg.MemChunkWriter{}
		writer, err := otelstef.NewMetricsWriter(buf, pkg.WriterOptions{})
		require.NoError(f, err)

		converter := OtlpToStefSorted{}
		err = converter.Convert(otlpMetricSrc, writer)
		require.NoError(f, err)

		err = writer.Flush()
		require.NoError(f, err)

		f.Add(buf.Bytes())
	}

	f.Fuzz(
		func(t *testing.T, data []byte) {
			reader, err := otelstef.NewMetricsReader(bytes.NewBuffer(data))
			if err != nil {
				return
			}
			for {
				err = reader.Read(pkg.ReadOptions{})
				if err != nil {
					break
				}
			}
		},
	)
}
