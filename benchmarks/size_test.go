package tests

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/benchmarks/encodings/otelarrow"
	"github.com/splunk/stef/benchmarks/encodings/otlp"
	parquetenc "github.com/splunk/stef/benchmarks/encodings/parquet"
	"github.com/splunk/stef/benchmarks/encodings/stef"
	"github.com/splunk/stef/benchmarks/generators"
	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/traces"
	"github.com/splunk/stef/go/pkg"
)

func replaceExt(fname string, ext string) string {
	idx := strings.Index(fname, ".")
	if idx > 0 {
		return fname[:idx] + ext
	}
	return fname
}

var metricsDataVariations = []struct {
	generator            generators.Generator
	firstUncompessedSize int
	firstZstdedSize      int
}{
	{
		generator: &generators.File{
			FilePath: "testdata/astronomy-otelmetrics.zst",
		},
	},
	{
		generator: &generators.File{
			FilePath: "testdata/hipstershop-otelmetrics.zst",
		},
	},
	{
		generator: &generators.File{
			FilePath:  "testdata/hipstershop-otelmetrics.zst",
			BatchSize: 1,
		},
	},
	{
		generator: &generators.File{
			FilePath: "testdata/hostandcollector-otelmetrics.zst",
		},
	},
	{
		generator: &generators.File{
			FilePath:  "testdata/hostandcollector-otelmetrics.zst",
			BatchSize: 1,
		},
	},
}

var sizeEncodings = []encodings.MetricEncoding{
	&otlp.OTLPEncoding{},
	&stef.STEFEncoding{Opts: pkg.WriterOptions{Compression: pkg.CompressionNone}},
	&stef.STEFSEncoding{Opts: pkg.WriterOptions{Compression: pkg.CompressionNone}},
	&stef.STEFUEncoding{Opts: pkg.WriterOptions{Compression: pkg.CompressionNone}},
	&parquetenc.Encoding{},
	&otelarrow.OtelArrowEncoding{},
}

func unitSize(totalSize int, unitCount int) float64 {
	if unitCount == 0 {
		return 0
	}
	return roundFloat(float64(totalSize)/float64(unitCount), 1)
}

func TestMetricsSize(t *testing.T) {

	fmt.Println("===== Encoded sizes")

	chart.BeginSection("Size Benchmarks - One Large Batch")

	for _, dataVariation := range metricsDataVariations {
		fmt.Printf("%-36s", dataVariation.generator.GetName())
		fmt.Println("    Uncompressed           Zstd Compressed")
		fmt.Printf("%-36s", "")
		fmt.Println(" Bytes Ratio By/pt        Bytes Ratio By/pt")

		wantChart := strings.HasSuffix(dataVariation.generator.GetName(), ".zst")

		if wantChart {
			chart.BeginChart("Dataset: "+dataVariation.generator.GetName(), t)
		}

		for _, encoding := range sizeEncodings {
			if (dataVariation.generator.GetName() == "astronomy-otelmetrics.zst") &&
				encoding.Name() == "Otel ARROW" {
				// Skip due to bug in Arrow encoding
				continue
			}

			batch := dataVariation.generator.Generate()
			pointCount := batch.DataPointCount()

			inmem, err := encoding.FromOTLP(batch)
			require.NoError(t, err)

			bodyBytes, err := encoding.Encode(inmem)

			if err != nil {
				log.Fatal(err)
			}

			if enc, ok := encoding.(*stef.STEFEncoding); ok {
				if enc.Opts.Compression == pkg.CompressionZstd {
					// Write STEF file if it does not exist
					fname := "testdata/" + replaceExt(
						dataVariation.generator.GetName(), "."+strings.ToLower(enc.Name()),
					)
					_, err = os.Stat(fname)
					if err != nil {
						os.WriteFile(fname, bodyBytes, 0644)
					}
				}
			}

			zstdedBytes := testutils.CompressZstd(bodyBytes)

			uncompressedSize := len(bodyBytes)
			zstdedSize := len(zstdedBytes)

			uncompressedRatioStr := "1.00"
			zstdedRatioStr := "1.00"

			if dataVariation.firstUncompessedSize == 0 {
				dataVariation.firstUncompessedSize = uncompressedSize
			} else {
				uncompressedRatioStr = fmt.Sprintf(
					"%1.2f", float64(dataVariation.firstUncompessedSize)/float64(uncompressedSize),
				)
			}

			if dataVariation.firstZstdedSize == 0 {
				dataVariation.firstZstdedSize = zstdedSize
			} else {
				zstdedRatioStr = fmt.Sprintf(
					"%1.2f", float64(dataVariation.firstZstdedSize)/float64(zstdedSize),
				)
			}

			fmt.Printf(
				"%-33v%9d %5s %5.1f      %7d %5s %5.1f\n",
				encoding.Name(),
				uncompressedSize,
				uncompressedRatioStr,
				float64(uncompressedSize)/float64(pointCount),
				zstdedSize,
				zstdedRatioStr,
				float64(zstdedSize)/float64(pointCount),
			)

			if wantChart {
				chart.Record(
					nil, encoding.LongName(),
					"Bytes/point (zstd)",
					unitSize(zstdedSize, pointCount),
				)
			}
		}

		if wantChart {
			chart.EndChart(
				"Bytes/point",
				charts.WithColorsOpts(opts.Colors{"#92C5F9"}),
			)
		}

		fmt.Println("")
	}
}

func TestMetricsMultipart(t *testing.T) {
	datasets := []struct {
		name string
	}{
		{
			name: "hostandcollector-otelmetrics",
		},
		{
			name: "astronomy-otelmetrics",
		},
	}

	testMultipartEncodings := []encodings.MetricMultipartEncoding{
		&otlp.OTLPEncoding{},
		&stef.STEFEncoding{},
		&stef.STEFSEncoding{},
		&stef.STEFUEncoding{},
		&otelarrow.OtelArrowEncoding{},
	}

	compressions := []string{"none", "zstd"}

	chart.BeginSection("Size - Many Batches, Multipart Metrics")

	for _, compression := range compressions {
		for _, dataset := range datasets {
			chart.BeginChart("Dataset: "+dataset.name, t)

			fmt.Printf("%-30s %4v %9v %4v\n", dataset.name, "Comp", "Bytes", "Ratio")

			firstSize := 0
			for _, encoding := range testMultipartEncodings {
				stream, err := encoding.StartMultipart(compression)
				require.NoError(t, err)

				// Encode each part one after another and write to the same STEF stream.
				// This models more closely the operation of STEF exporter in Collector.

				parts, err := testutils.ReadMultipartOTLPFile("testdata/" + dataset.name + ".zst")
				require.NoError(t, err)

				pointCount := 0
				for _, part := range parts {
					pointCount += part.DataPointCount()
					err := stream.AppendPart(part)
					require.NoError(t, err)
				}
				byts, err := stream.FinishStream()
				require.NoError(t, err)

				curSize := len(byts)
				var delta string
				if firstSize == 0 {
					delta = "x 1.00"
				} else {
					delta = fmt.Sprintf("x %.2f", float64(firstSize)/float64(curSize))
				}
				fmt.Printf(
					"%-30s %4v %9v %s\n", encoding.Name(), compression, curSize, delta,
				)
				if firstSize == 0 {
					firstSize = curSize
				}

				chart.Record(
					nil, encoding.LongName(), "Bytes/point, compression="+compression,
					unitSize(curSize, pointCount),
				)
			}

			chart.EndChart(
				"Bytes/point",
				charts.WithColorsOpts(opts.Colors{"#87BB62"}),
			)
		}
	}
}

func TestSTEFVeryShortFrames(t *testing.T) {
	input, err := testutils.ReadOTLPFile("testdata/hipstershop-otelmetrics.zst")
	require.NoError(t, err)

	compressions := []pkg.Compression{pkg.CompressionNone, pkg.CompressionZstd}
	for _, compression := range compressions {
		tefEncoding := stef.STEFEncoding{Opts: pkg.WriterOptions{Compression: compression}}
		inmem, err := tefEncoding.FromOTLP(input)
		require.NoError(t, err)
		tefBytes, err := tefEncoding.Encode(inmem)

		tefReader, err := otelstef.NewMetricsReader(bytes.NewBuffer(tefBytes))
		require.NoError(t, err)

		shortFrameBuf := pkg.MemChunkWriter{}
		tefWriter, err := otelstef.NewMetricsWriter(&shortFrameBuf, pkg.WriterOptions{Compression: compression})
		require.NoError(t, err)

		for {
			err := tefReader.Read(pkg.ReadOptions{})
			if err == io.EOF {
				break
			}

			tefWriter.Record.CopyFrom(&tefReader.Record)

			err = tefWriter.Write()
			if err != nil {
				require.NoError(t, err)
			}
			err = tefWriter.Flush()
			if err != nil {
				require.NoError(t, err)
			}
		}
		fmt.Printf("Compression:        %v\n", stefCompression2str(compression))
		fmt.Printf("Long frames total:  %7d bytes\n", len(tefBytes))
		fmt.Printf("Short frames total: %7d bytes\n", len(shortFrameBuf.Bytes()))
		fmt.Printf("Increase ratio:     %.2fx\n", float64(len(shortFrameBuf.Bytes()))/float64(len(tefBytes)))
	}
}

func stefCompression2str(compression pkg.Compression) any {
	switch compression {
	case pkg.CompressionNone:
		return "none"
	case pkg.CompressionZstd:
		return "zstd"
	}
	panic("unknown compression")
}

func TestTracesMultipart(t *testing.T) {
	u := ptrace.ProtoUnmarshaler{}

	fileNames := []string{"astronomy-oteltraces", "hipstershop-oteltraces"}

	chart.BeginSection("Size - Many Batches, Multipart Traces")

	for _, fileName := range fileNames {
		fmt.Println("======= " + fileName)

		var otlpParts [][]byte
		traceData, err := testutils.ReadMultipartOTLPFileGeneric(
			"testdata/"+fileName+".zst", func(data []byte) (any, error) {
				otlpParts = append(otlpParts, data)
				return u.UnmarshalTraces(data)
			},
		)
		require.NoError(t, err)

		compressions := []pkg.Compression{pkg.CompressionNone, pkg.CompressionZstd}
		sorteds := []bool{false, true}

		for _, compression := range compressions {
			var compressionStr string
			if compression == pkg.CompressionZstd {
				compressionStr = "zstd"
			} else {
				compressionStr = "none"
			}
			fmt.Println(compressionStr)

			chart.BeginChart("Dataset: "+fileName, t)

			otlpSize := 0
			spanCount := 0
			for i := 0; i < len(traceData); i++ {
				if compression == pkg.CompressionNone {
					otlpSize += len(otlpParts[i])
				} else {
					otlpZstd := testutils.CompressZstd(otlpParts[i])
					otlpSize += len(otlpZstd)
				}
				spanCount += traceData[i].(ptrace.Traces).SpanCount()
			}

			chart.Record(
				nil, "OTLP", "Bytes/span, compression="+compressionStr,
				unitSize(otlpSize, spanCount),
			)

			for _, sorted := range sorteds {
				var sortedStr string
				if sorted {
					sortedStr = "Sorted"
				} else {
					sortedStr = "Unsorted"
				}
				fmt.Println(sortedStr)

				outputBuf := &pkg.MemChunkWriter{}
				writer, err := otelstef.NewSpansWriter(outputBuf, pkg.WriterOptions{Compression: compression})
				require.NoError(t, err)

				converter := traces.OtlpToStefUnsorted{Sorted: sorted}

				for i := 0; i < len(traceData); i++ {
					err = converter.Convert(traceData[i].(ptrace.Traces), writer)
					require.NoError(t, err)

					err = writer.Flush()
					require.NoError(t, err)
				}

				stefSize := len(outputBuf.Bytes())

				fmt.Printf(
					"Traces OTLP: %8d (%5.1f bytes/span)\n",
					otlpSize,
					float64(otlpSize)/float64(spanCount),
				)
				fmt.Printf(
					"Traces STEF: %8d (%5.1f bytes/span)\n",
					stefSize,
					float64(stefSize)/float64(spanCount),
				)
				fmt.Printf(
					"Ratio:       %8.2f\n", float64(otlpSize)/float64(stefSize),
				)

				chart.Record(
					nil, "STEF "+sortedStr, "Bytes/span, compression="+compressionStr,
					unitSize(stefSize, spanCount),
				)
			}
			chart.EndChart(
				"Bytes/span",
				charts.WithColorsOpts(opts.Colors{"#87BB62"}),
			)
		}
	}
}
