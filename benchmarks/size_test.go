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
	"github.com/splunk/stef/go/otel/oteltef"
	traces2 "github.com/splunk/stef/go/pdata/traces"
	"github.com/splunk/stef/go/pkg"
)

func TestTracesMultipart(t *testing.T) {
	u := ptrace.ProtoUnmarshaler{}

	fileNames := []string{"testdata/astronomy-oteltraces.zst", "testdata/hipstershop-oteltraces.zst"}

	for _, fileName := range fileNames {
		fmt.Println("======= " + fileName)

		var otlpParts [][]byte
		traces, err := testutils.ReadMultipartOTLPFileGeneric(
			fileName, func(data []byte) (any, error) {
				otlpParts = append(otlpParts, data)
				return u.UnmarshalTraces(data)
			},
		)
		require.NoError(t, err)

		compressions := []pkg.Compression{pkg.CompressionNone, pkg.CompressionZstd}
		sorteds := []bool{false, true}

		for _, sorted := range sorteds {
			if sorted {
				fmt.Println("Sorted")
			} else {
				fmt.Println("Unsorted")
			}

			for _, compression := range compressions {

				outputBuf := &pkg.MemChunkWriter{}
				writer, err := oteltef.NewSpansWriter(outputBuf, pkg.WriterOptions{Compression: compression})
				require.NoError(t, err)

				converter := traces2.PdataToSTEFTraces{Sorted: sorted}

				otlpSize := 0
				for i := 0; i < len(traces); i++ {
					err = converter.WriteTraces(traces[i].(ptrace.Traces), writer)
					require.NoError(t, err)

					err = writer.Flush()
					require.NoError(t, err)

					if compression == pkg.CompressionNone {
						otlpSize += len(otlpParts[i])
					} else {
						otlpZstd := testutils.CompressZstd(otlpParts[i])
						otlpSize += len(otlpZstd)
					}
				}

				stefSize := len(outputBuf.Bytes())

				if compression == pkg.CompressionZstd {
					fmt.Println("zstd")
				} else {
					fmt.Println("none")
				}
				fmt.Printf("Traces OTLP: %8d\n", otlpSize)
				fmt.Printf("Traces STEF: %8d\n", stefSize)
				fmt.Printf(
					"Ratio:       %8.2f\n", float64(otlpSize)/float64(stefSize),
				)
			}
		}
	}
}

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
	&stef.STEFUEncoding{Opts: pkg.WriterOptions{Compression: pkg.CompressionNone}},
	&parquetenc.Encoding{},
	&otelarrow.OtelArrowEncoding{},
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
					"Compressed size in bytes (zstd)",
					float64(zstdedSize),
				)
			}
		}

		if wantChart {
			chart.EndChart(
				"Bytes",
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
		&stef.STEFUEncoding{},
		&otelarrow.OtelArrowEncoding{},
	}

	compressions := []string{"none", "zstd"}

	chart.BeginSection("Size Benchmarks - Many Batches, Multipart")

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

				for _, part := range parts {
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
					nil, encoding.LongName(), "Size in bytes, compression="+compression,
					float64(curSize),
				)
			}

			chart.EndChart(
				"Bytes",
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

		tefReader, err := oteltef.NewMetricsReader(bytes.NewBuffer(tefBytes))
		require.NoError(t, err)

		shortFrameBuf := pkg.MemChunkWriter{}
		tefWriter, err := oteltef.NewMetricsWriter(&shortFrameBuf, pkg.WriterOptions{Compression: compression})
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
