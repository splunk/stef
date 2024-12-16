package tests

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/tigrannajaryan/stef/benchmarks/encodings/stef"
	"github.com/tigrannajaryan/stef/benchmarks/testutils"
	"github.com/tigrannajaryan/stef/stef-go/pkg"
	"github.com/tigrannajaryan/stef/stef-otel/oteltef"
	"github.com/tigrannajaryan/stef/stef-pdata/metrics"
	"github.com/tigrannajaryan/stef/stef-pdata/metrics/testtools"
)

func TestConvertTEFFromToOTLP(t *testing.T) {
	tests := []struct {
		file           string
		withSizePrefix bool
	}{
		{
			file:           "testdata/oteldemo-with-histogram.otlp.zst",
			withSizePrefix: true,
		},
		{
			file: "testdata/hipstershop.pb.zst",
		},
		{
			file:           "testdata/astronomyshop.pb.zst",
			withSizePrefix: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.file, func(t *testing.T) {
				otlpDataSrc, err := testutils.ReadOTLPFile(test.file, test.withSizePrefix)
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

				toOtlp := metrics.NewTefToSortedTree()
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
		"testdata/oteldemo-with-histogram.otlp.zst",
		"testdata/astronomyshop.pb.zst",
		"testdata/hostandcollectormetrics.pb.zst",
	}

	//stefEncoding := stef.STEFEncoding{}
	tefEncoding := stef.STEFEncoding{}

	for _, inputFile := range testInputOtlpFiles {
		t.Run(
			inputFile, func(t *testing.T) {

				parts, err := testutils.ReadMultipartOTLPFile(inputFile)
				require.NoError(t, err)

				//stefStream, err := stefEncoding.StartMultipart("")
				//require.NoError(t, err)

				tefStream, err := tefEncoding.StartMultipart("")
				require.NoError(t, err)

				for _, part := range parts {
					//err = stefStream.AppendPart(part)
					//require.NoError(t, err)
					err = tefStream.AppendPart(part)
					require.NoError(t, err)
				}
				//stefBytes, err := stefStream.FinishStream()
				//require.NoError(t, err)

				tefBytes, err := tefStream.FinishStream()
				require.NoError(t, err)

				//stefReader, err := metrics.NewReader(bytes.NewBuffer(stefBytes))
				//require.NoError(t, err)

				tefReader, err := oteltef.NewMetricsReader(bytes.NewBuffer(tefBytes))
				require.NoError(t, err)

				i := 0
				for {
					//stefRec, err := stefReader.Read()
					//stefRec = stefRec
					//if err == io.EOF {
					//	break
					//}
					//require.NoError(t, err)

					tefRec, err := tefReader.Read()
					if err == io.EOF {
						break
					}
					require.NoError(t, err, i)
					require.NotNil(t, tefRec, i)
					//oteltef.EqualRecord(t, stefRec, tefRec)
					i++
				}
			},
		)
	}
}
