package metrics

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/oteltef"
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
		writer, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{})
		require.NoError(t, err)

		toTef := NewOtlpToSortedTree()
		sortedByMetric, err := toTef.FromOtlp(otlpMetricSrc.ResourceMetrics())
		require.NoError(t, err)

		err = sortedByMetric.ToTef(writer)
		require.NoError(t, err)

		assert.EqualValues(t, srcCount, int(writer.RecordCount()))

		sortedByMetric = nil

		err = writer.Flush()
		require.NoError(t, err)

		reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
		require.NoError(t, err)

		toOtlp := NewTefToSortedTree()
		sortedByResource, err := toOtlp.FromTef(reader)
		require.NoError(t, err)

		assert.EqualValues(t, writer.RecordCount(), reader.RecordCount())

		otlpMetricCopy, err := sortedByResource.ToOtlp()
		require.NoError(t, err)

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

func FuzzReader(f *testing.F) {
	otlpMetrics, err := pict.GenerateMetrics("testdata/generated_pict_pairs_metrics.txt")
	require.NoError(f, err)

	for _, otlpMetricSrc := range otlpMetrics {
		buf := &pkg.MemChunkWriter{}
		writer, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{})
		require.NoError(f, err)

		toStef := NewOtlpToSortedTree()
		sortedByMetric, err := toStef.FromOtlp(otlpMetricSrc.ResourceMetrics())
		require.NoError(f, err)

		err = sortedByMetric.ToTef(writer)
		require.NoError(f, err)

		err = writer.Flush()
		require.NoError(f, err)

		f.Add(buf.Bytes())
	}

	f.Fuzz(
		func(t *testing.T, data []byte) {
			reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(data))
			if err != nil {
				return
			}
			for {
				record, err := reader.Read()
				if err != nil {
					break
				}
				require.NotNil(t, record)
			}
		},
	)
}
