package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pkg"
)

func TestCopy(t *testing.T) {
	files := []string{
		"hipstershop-otelmetrics.stefz",
		"hostandcollector-otelmetrics.stefz",
		"astronomy-otelmetrics.stefz",
	}

	fmt.Printf(
		"%-30s %12s %12s\n",
		"File", "Uncompressed", "Zstd Bytes",
	)

	for _, file := range files {

		tefBytes, err := os.ReadFile("testdata/generated/" + file)
		require.NoError(t, err)

		tefReader, err := otelstef.NewMetricsReader(bytes.NewBuffer(tefBytes))
		require.NoError(t, err)

		cw := &pkg.MemChunkWriter{}
		tefWriter, err := otelstef.NewMetricsWriter(cw, pkg.WriterOptions{})
		require.NoError(t, err)

		recCount := 0
		for {
			err := tefReader.Read(pkg.ReadOptions{})
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			if recCount == 1 {
				_ = recCount
			}

			copyModified(&tefWriter.Record, &tefReader.Record)

			err = tefWriter.Write()
			require.NoError(t, err)
			recCount++
		}

		err = tefWriter.Flush()
		require.NoError(t, err)

		zstdBuf := bytes.NewBuffer(nil)
		var zstdEncoder, _ = zstd.NewWriter(zstdBuf, zstd.WithEncoderLevel(zstd.SpeedDefault))
		_, err = zstdEncoder.Write(cw.Bytes())
		require.NoError(t, err)
		err = zstdEncoder.Flush()
		require.NoError(t, err)

		fmt.Printf(
			"%-30s %12d %12d\n",
			file,
			len(cw.Bytes()),
			len(zstdBuf.Bytes()),
		)
	}
}

func BenchmarkReadSTEF(b *testing.B) {
	tefBytes, err := os.ReadFile("testdata/generated/hipstershop-otelmetrics.stefz")
	require.NoError(b, err)

	tefSrc, err := otelstef.NewMetricsReader(bytes.NewBuffer(tefBytes))
	require.NoError(b, err)

	cw := &pkg.MemChunkWriter{}
	tefWriter, err := otelstef.NewMetricsWriter(cw, pkg.WriterOptions{})
	require.NoError(b, err)

	recCount := 0
	for {
		err := tefSrc.Read(pkg.ReadOptions{})
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		copyModified(&tefWriter.Record, &tefSrc.Record)

		err = tefWriter.Write()
		if err != nil {
			panic(err)
		}
		recCount++
	}
	err = tefWriter.Flush()
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, err := otelstef.NewMetricsReader(bytes.NewBuffer(cw.Bytes()))
		if err != nil {
			panic(err)
		}

		for i := 0; i < recCount; i++ {
			err := reader.Read(pkg.ReadOptions{})
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/point")
}

func BenchmarkReadSTEFZ(b *testing.B) {
	tefBytes, err := os.ReadFile("testdata/generated/hipstershop-otelmetrics.stefz")
	require.NoError(b, err)

	recCount := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, err := otelstef.NewMetricsReader(bytes.NewBuffer(tefBytes))
		if err != nil {
			panic(err)
		}

		recCount = 0
		for {
			err := reader.Read(pkg.ReadOptions{})
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			recCount++
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/point")
}

func BenchmarkReadSTEFZWriteSTEF(b *testing.B) {
	tefBytes, err := os.ReadFile("testdata/generated/hipstershop-otelmetrics.stefz")
	require.NoError(b, err)

	recCount := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tefReader, err := otelstef.NewMetricsReader(bytes.NewBuffer(tefBytes))
		if err != nil {
			panic(err)
		}

		cw := &pkg.MemChunkWriter{}
		tefWriter, err := otelstef.NewMetricsWriter(cw, pkg.WriterOptions{})
		if err != nil {
			panic(err)
		}

		recCount = 0
		for {
			err := tefReader.Read(pkg.ReadOptions{})
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}

			copyModified(&tefWriter.Record, &tefReader.Record)

			err = tefWriter.Write()
			if err != nil {
				panic(err)
			}
			recCount++
		}
		err = tefWriter.Flush()
		if err != nil {
			panic(err)
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/point")
}

func copyModified(dst *otelstef.Metrics, src *otelstef.Metrics) {
	if src.IsEnvelopeModified() {
		dst.Envelope().CopyFrom(src.Envelope())
	}

	if src.IsResourceModified() {
		dst.SetResource(src.Resource())
	}

	if src.IsScopeModified() {
		dst.SetScope(src.Scope())
	}

	if src.IsMetricModified() {
		dst.SetMetric(src.Metric())
	}

	if src.IsAttributesModified() {
		dst.Attributes().CopyFrom(src.Attributes())
	}

	if src.IsPointModified() {
		dst.Point().CopyFrom(src.Point())
	}
}
