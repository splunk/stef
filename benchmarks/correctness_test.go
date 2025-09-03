package tests

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/benchmarks/encodings/stef"
	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/metrics"
	"github.com/splunk/stef/go/pdata/metrics/testtools"
	"github.com/splunk/stef/go/pkg"
)

func TestConvertTEFFromToOTLP(t *testing.T) {
	tests := []struct {
		file string
	}{
		{
			file: "testdata/hipstershop-otelmetrics.zst",
		},
		{
			file: "testdata/astronomy-otelmetrics.zst",
		},
	}

	for _, test := range tests {
		t.Run(
			test.file, func(t *testing.T) {
				otlpDataSrc, err := testutils.ReadOTLPFile(test.file)
				require.NoError(t, err)

				testtools.NormalizeMetrics(otlpDataSrc)
				srcCount := otlpDataSrc.DataPointCount()

				buf := &pkg.MemChunkWriter{}
				writer, err := oteltef.NewMetricsWriter(buf, pkg.WriterOptions{})
				require.NoError(t, err)

				toStef := metrics.NewOtlpToSortedTree()
				sortedByMetric, err := toStef.FromOtlp(otlpDataSrc.ResourceMetrics())
				require.NoError(t, err)

				err = sortedByMetric.ToTef(writer)
				require.NoError(t, err)

				//assert.EqualValues(t, srcCount, int(writer.Stats().Datapoints))

				sortedByMetric = nil

				err = writer.Flush()
				require.NoError(t, err)

				reader, err := oteltef.NewMetricsReader(bytes.NewBuffer(buf.Bytes()))
				require.NoError(t, err)

				toOtlp := metrics.NewSTEFToSortedTree()
				sortedByResource, err := toOtlp.FromTef(reader)
				require.NoError(t, err)

				//assert.EqualValues(t, writer.Stats().Datapoints, reader.Stats().Datapoints)

				otlpDataCopy, err := sortedByResource.ToOtlp()
				require.NoError(t, err)

				testtools.NormalizeMetrics(otlpDataCopy)

				copyCount := otlpDataCopy.DataPointCount()
				assert.EqualValues(t, srcCount, copyCount)

				assert.NoError(t, testtools.DiffMetrics(otlpDataSrc, otlpDataCopy))
				assert.True(t, bytes.Equal(toBytes(t, otlpDataSrc), toBytes(t, otlpDataCopy)))
			},
		)
	}
}

func toBytes(t *testing.T, data pmetric.Metrics) []byte {
	marshaler := pmetric.ProtoMarshaler{}
	bytes, err := marshaler.MarshalMetrics(data)
	require.NoError(t, err)
	return bytes
}

func TestTEFMultiPart(t *testing.T) {
	var testInputOtlpFiles = []string{
		"testdata/astronomy-otelmetrics.zst",
		"testdata/hostandcollector-otelmetrics.zst",
	}

	tefEncoding := stef.STEFEncoding{}

	for _, inputFile := range testInputOtlpFiles {
		t.Run(
			inputFile, func(t *testing.T) {

				parts, err := testutils.ReadMultipartOTLPFile(inputFile)
				require.NoError(t, err)

				tefStream, err := tefEncoding.StartMultipart("")
				require.NoError(t, err)

				for _, part := range parts {
					err = tefStream.AppendPart(part)
					require.NoError(t, err)
				}

				tefBytes, err := tefStream.FinishStream()
				require.NoError(t, err)

				tefReader, err := oteltef.NewMetricsReader(bytes.NewBuffer(tefBytes))
				require.NoError(t, err)

				i := 0
				for {
					err := tefReader.Read(pkg.ReadOptions{})
					if err == io.EOF {
						break
					}
					require.NoError(t, err, i)
					i++
				}
			},
		)
	}
}
