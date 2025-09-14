package tests

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/benchmarks/encodings/otelarrow"
	"github.com/splunk/stef/benchmarks/encodings/otlp"
	parquetenc "github.com/splunk/stef/benchmarks/encodings/parquet"
	"github.com/splunk/stef/benchmarks/encodings/stef"
	"github.com/splunk/stef/benchmarks/generators"
	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/metrics"
	"github.com/splunk/stef/go/pkg"
)

var speedEncodings = []encodings.MetricEncoding{
	&otlp.OTLPEncoding{},
	&stef.STEFEncoding{Opts: pkg.WriterOptions{Compression: pkg.CompressionNone}},
	&stef.STEFUEncoding{Opts: pkg.WriterOptions{Compression: pkg.CompressionNone}},
	&parquetenc.Encoding{},
	&otelarrow.OtelArrowEncoding{},
}

var benchmarkDataVariations = []struct {
	generator generators.Generator
}{
	//{
	//	generator: &generators.File{
	//		FilePath: "testdata/host_and_collector.pb",
	//	},
	//},
	{
		generator: &generators.File{
			FilePath: "testdata/hipstershop-otelmetrics.zst",
		},
	},
}

var chart = BarOutput{}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	chart.Begin()
	defer chart.End()
	m.Run()
}

func addZstdCompressTime(
	b *testing.B,
	encoding encodings.MetricEncoding,
	bodyBytes []byte,
	dataPointCount int,
) {
	if !chartsEnabled() {
		return
	}

	b.Run(
		fmt.Sprintf("%s/zstd", encoding.Name()),
		func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				zstdBytes := testutils.CompressZstd(bodyBytes)
				if zstdBytes == nil {
					log.Fatal("compression failed")
				}
			}
			chart.RecordStacked(
				b,
				encoding.LongName(),
				"Zstd Compress",
				float64(b.Elapsed().Nanoseconds())/float64(b.N*dataPointCount),
			)
		},
	)
}

func addZstdDecompressTime(
	b *testing.B,
	encoding encodings.MetricEncoding,
	bodyBytes []byte,
	dataPointCount int,
) {
	if !chartsEnabled() {
		return
	}

	zstdBytes := testutils.CompressZstd(bodyBytes)
	if zstdBytes == nil {
		log.Fatal("compression failed")
	}
	b.Run(
		fmt.Sprintf("%s/zstd", encoding.Name()),
		func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := testutils.DecompressZstd(zstdBytes)
				if err != nil {
					log.Fatal(err)
				}
			}
			chart.RecordStacked(
				b,
				encoding.LongName(),
				"Zstd Decompress",
				float64(b.Elapsed().Nanoseconds())/float64(b.N*dataPointCount),
			)
		},
	)

}

func BenchmarkSerializeNative(b *testing.B) {
	chart.BeginSection("Speed Benchmarks")

	chart.BeginChart("Serialization Speed", b)
	defer chart.EndChart(
		"ns/point",
		charts.WithColorsOpts(opts.Colors{"#92C5F9", "#12C5F9"}),
	)

	for _, dataVariation := range benchmarkDataVariations {
		for _, encoding := range speedEncodings {
			if _, ok := encoding.(*otelarrow.OtelArrowEncoding); ok {
				// Skip Arrow, it does not have native serialization
				continue
			}
			batch := dataVariation.generator.Generate()
			inmem, err := encoding.FromOTLP(batch)
			require.NoError(b, err)
			b.Run(
				fmt.Sprintf("%s/serialize", encoding.Name()),
				func(b *testing.B) {
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						bodyBytes, err := encoding.Encode(inmem)
						if err != nil || bodyBytes == nil {
							log.Fatal(err)
						}
					}
					chart.RecordStacked(
						b,
						encoding.LongName(),
						"Serialize",
						float64(b.Elapsed().Nanoseconds())/float64(b.N*batch.DataPointCount()),
					)
				},
			)
			bodyBytes, err := encoding.Encode(inmem)
			require.NotNil(b, bodyBytes)
			require.NoError(b, err)

			addZstdCompressTime(b, encoding, bodyBytes, batch.DataPointCount())
		}
	}
	b.ReportAllocs()
}

func BenchmarkDeserializeNative(b *testing.B) {
	chart.BeginChart("Deserialization Speed", b)
	defer chart.EndChart(
		"ns/point",
		charts.WithColorsOpts(opts.Colors{"#92C5F9", "#12C5F9"}),
	)

	for _, dataVariation := range benchmarkDataVariations {
		for _, encoding := range speedEncodings {
			if _, ok := encoding.(*otelarrow.OtelArrowEncoding); ok {
				// Skip Arrow, it does not have native serialization
				continue
			}
			batch := dataVariation.generator.Generate()
			inmem, err := encoding.FromOTLP(batch)
			require.NoError(b, err)
			b.Run(
				fmt.Sprintf("%s/deser", encoding.Name()),
				func(b *testing.B) {
					bodyBytes, err := encoding.Encode(inmem)
					if err != nil {
						log.Fatal(err)
					}

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						_, err = encoding.Decode(bodyBytes)
						if err != nil {
							log.Fatal(err)
						}
					}
					chart.RecordStacked(
						b,
						encoding.LongName(),
						"Deserialize",
						float64(b.Elapsed().Nanoseconds())/float64(b.N*batch.DataPointCount()),
					)
				},
			)
			bodyBytes, err := encoding.Encode(inmem)
			if err != nil {
				log.Fatal(err)
			}
			addZstdDecompressTime(b, encoding, bodyBytes, batch.DataPointCount())
		}
	}
	b.ReportAllocs()
}

func BenchmarkSerializeFromPdata(b *testing.B) {
	chart.BeginChart("Serialization From pdata Speed", b)
	defer chart.EndChart(
		"ns/point",
		charts.WithColorsOpts(opts.Colors{"#92C5F9", "#12C5F9"}),
	)

	for _, dataVariation := range benchmarkDataVariations {
		for _, encoding := range speedEncodings {
			if dataVariation.generator.GetName() == "hostandcollector-otelmetrics.zst" &&
				encoding.Name() == "ARROW" {
				// Skip due to bug in Arrow encoding
				continue
			}
			batch := dataVariation.generator.Generate()
			b.Run(
				fmt.Sprintf("%s/serialize", encoding.Name()),
				func(b *testing.B) {
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						inmem, err := encoding.FromOTLP(batch)
						require.NoError(b, err)
						bodyBytes, err := encoding.Encode(inmem)
						require.NotNil(b, bodyBytes)
						require.NoError(b, err)
					}
					chart.Record(
						b,
						encoding.LongName(),
						"CPU time to serialize one data point",
						float64(b.Elapsed().Nanoseconds())/float64(b.N*batch.DataPointCount()),
					)
				},
			)
			inmem, err := encoding.FromOTLP(batch)
			require.NoError(b, err)
			bodyBytes, err := encoding.Encode(inmem)
			require.NotNil(b, bodyBytes)
			require.NoError(b, err)
			addZstdCompressTime(b, encoding, bodyBytes, batch.DataPointCount())
		}
	}
	b.ReportAllocs()
}

func BenchmarkDeserializeToPdata(b *testing.B) {
	chart.BeginChart("Deserialization To pdata Speed", b)
	defer chart.EndChart(
		"ns/point",
		charts.WithColorsOpts(opts.Colors{"#92C5F9", "#12C5F9"}),
	)

	for _, dataVariation := range benchmarkDataVariations {
		for _, encoding := range speedEncodings {
			if dataVariation.generator.GetName() == "hostandcollector-otelmetrics.zst" &&
				encoding.Name() == "ARROW" {
				// Skip due to bug in Arrow encoding
				continue
			}
			batch := dataVariation.generator.Generate()
			inmem, err := encoding.FromOTLP(batch)
			require.NoError(b, err)
			bodyBytes, err := encoding.Encode(inmem)
			if err != nil {
				log.Fatal(err)
			}

			b.Run(
				fmt.Sprintf("%s/deserialize", encoding.Name()),
				func(b *testing.B) {

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						_, err = encoding.ToOTLP(bodyBytes)
						if err != nil {
							log.Fatal(err)
						}
					}
					chart.Record(
						b,
						encoding.LongName(),
						"CPU time to deserialize one data point",
						float64(b.Elapsed().Nanoseconds())/float64(b.N*batch.DataPointCount()),
					)
				},
			)
			addZstdDecompressTime(b, encoding, bodyBytes, batch.DataPointCount())
		}
	}
	b.ReportAllocs()
}

/* Need to rewrite this to use STEF.ReadMany() API when it becomes available.
func BenchmarkReaderReadMany(b *testing.B) {
	generator := &generators.File{
		FilePath: "testdata/hipstershop-otelmetrics.zst",
	}

	encoding := stef.STEFEncoding{}
	batch := generator.Generate()
	inmem, err := encoding.FromOTLP(batch)
	require.NoError(b, err)

	bodyBytes, err := encoding.Encode(inmem)
	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()
	var records metrics.Records
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(bodyBytes)
		reader, err := metrics.NewReader(buf)
		if err != nil {
			log.Fatal(err)
		}

		err = reader.ReadMany(0, &records)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}
	b.ReportMetric(
		float64(b.Elapsed().Nanoseconds())/float64(b.N*batch.DataPointCount()),
		"ns/point",
	)
}
*/

func BenchmarkSTEFReaderRead(b *testing.B) {
	generator := &generators.File{
		FilePath: "testdata/hipstershop-otelmetrics.zst",
	}

	encoding := stef.STEFEncoding{}
	batch := generator.Generate()
	inmem, err := encoding.FromOTLP(batch)
	require.NoError(b, err)

	bodyBytes, err := encoding.Encode(inmem)
	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(bodyBytes)
		reader, err := oteltef.NewMetricsReader(buf)
		if err != nil {
			log.Fatal(err)
		}

		for {
			err := reader.Read(pkg.ReadOptions{})
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	b.ReportMetric(
		float64(b.Elapsed().Nanoseconds())/float64(b.N*batch.DataPointCount()),
		"ns/point",
	)
}

var multipartFiles = []string{"astronomy-otelmetrics"}

func BenchmarkSTEFSerializeMultipart(b *testing.B) {
	for _, file := range multipartFiles {
		parts, err := testutils.ReadMultipartOTLPFile("testdata/" + file + ".zst")
		require.NoError(b, err)
		b.Run(
			file,
			func(b *testing.B) {
				pointCount := 0
				for i := 0; i < b.N; i++ {
					outputBuf := &pkg.MemChunkWriter{}
					writer, err := oteltef.NewMetricsWriter(
						outputBuf, pkg.WriterOptions{Compression: pkg.CompressionZstd},
					)
					require.NoError(b, err)

					// Encode each part one after another and write to the same STEF stream.
					// This models more closely the operation of STEF exporter in Collector.

					for _, part := range parts {
						converter := metrics.OtlpToStefSorted{}
						err = converter.Convert(part, writer)
						require.NoError(b, err)

						err = writer.Flush()
						require.NoError(b, err)
					}
					pointCount = int(writer.RecordCount())
				}
				b.ReportMetric(
					float64(b.Elapsed().Nanoseconds())/float64(b.N*pointCount),
					"ns/point",
				)
			},
		)
	}
}

func BenchmarkSTEFDeserializeMultipart(b *testing.B) {
	for _, file := range multipartFiles {
		parts, err := testutils.ReadMultipartOTLPFile("testdata/" + file + ".zst")
		require.NoError(b, err)
		b.Run(
			file,
			func(b *testing.B) {
				outputBuf := &pkg.MemChunkWriter{}
				writer, err := oteltef.NewMetricsWriter(
					outputBuf, pkg.WriterOptions{Compression: pkg.CompressionZstd},
				)
				require.NoError(b, err)

				// Encode each part one after another and write to the same STEF stream.
				// This models more closely the operation of STEF exporter in Collector.
				pointCount := 0
				for _, part := range parts {
					converter := metrics.OtlpToStefSorted{}
					err = converter.Convert(part, writer)
					require.NoError(b, err)
				}
				err = writer.Flush()
				require.NoError(b, err)
				pointCount = int(writer.RecordCount())

				b.ResetTimer()
				encoding := stef.STEFEncoding{}
				for i := 0; i < b.N; i++ {
					_, err = encoding.Decode(outputBuf.Bytes())
					require.NoError(b, err)
				}
				b.ReportMetric(
					float64(b.Elapsed().Nanoseconds())/float64(b.N*pointCount),
					"ns/point",
				)
			},
		)
	}
}
