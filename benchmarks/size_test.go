package tests

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/benchmarks/encodings/otelarrow"
	"github.com/splunk/stef/benchmarks/encodings/otlp"
	"github.com/splunk/stef/benchmarks/encodings/stef"
	"github.com/splunk/stef/benchmarks/generators"
	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/stef-go/pkg"
	"github.com/splunk/stef/stef-otel/oteltef"
	traces2 "github.com/splunk/stef/stef-pdata/traces"
)

func TestTracesMultipart(t *testing.T) {
	u := ptrace.ProtoUnmarshaler{}

	fileNames := []string{"testdata/astronomy-traces.pb.zst", "testdata/hipstershop_traces.pb.zst"}

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

func TestMetricsSize(t *testing.T) {

	dataVariations := []struct {
		generator            generators.Generator
		firstUncompessedSize int
		firstZstdedSize      int
	}{
		{
			generator: &generators.File{
				FilePath:      "testdata/oteldemo-with-histogram.otlp.zst",
				MultipartFile: true,
			},
		},
		{
			generator: &generators.File{
				FilePath:      "testdata/astronomyshop.pb.zst",
				MultipartFile: true,
			},
		},
		{
			generator: &generators.File{
				FilePath: "testdata/hipstershop.pb.zst",
			},
		},
		{
			generator: &generators.File{
				FilePath:  "testdata/hipstershop.pb.zst",
				BatchSize: 1,
			},
		},
		{
			generator: &generators.File{
				FilePath:      "testdata/hostandcollectormetrics.pb.zst",
				MultipartFile: true,
			},
		},
		{
			generator: &generators.File{
				FilePath:      "testdata/hostandcollectormetrics.pb.zst",
				MultipartFile: true,
				BatchSize:     1,
			},
		},
		{
			generator: &generators.Random{
				Name:                    "Int64/1DP/1000TS",
				TimeseriesCount:         1000,
				DatapointsPerTimeseries: 1,
				IncludeInt64:            true,
			},
		},
		{
			generator: &generators.Random{
				Name:                    "Int64/10DP/100TS",
				TimeseriesCount:         100,
				DatapointsPerTimeseries: 10,
				IncludeInt64:            true,
			},
		},
		{
			generator: &generators.Random{
				Name:                    "Int64/100DP/10TS",
				TimeseriesCount:         10,
				DatapointsPerTimeseries: 100,
				IncludeInt64:            true,
			},
		},
	}

	fmt.Println("===== Encoded sizes")

	for _, dataVariation := range dataVariations {
		fmt.Printf("%-36s", dataVariation.generator.GetName())
		fmt.Println("    Uncompressed           Zstd Compressed")
		fmt.Printf("%-36s", "")
		fmt.Println(" Bytes Ratio By/pt        Bytes Ratio By/pt")
		for _, encoding := range testEncodings {
			if (dataVariation.generator.GetName() == "astronomyshop.pb.zst") &&
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

			//if enc, ok := encoding.(*stef.STEFEncoding); ok {
			//	fname := "testdata/" + dataVariation.generator.GetName() + "." + strings.ToLower(enc.Name())
			//	os.WriteFile(fname, bodyBytes, 0644)
			//}

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
		}
		fmt.Println("")
	}
}

func TestMetricsMultipart(t *testing.T) {
	datasets := []struct {
		name string
	}{
		{
			name: "oteldemo-with-histogram.otlp",
		},
		{
			name: "hostandcollectormetrics.pb",
		},
		{
			name: "astronomyshop.pb",
		},
	}

	testMultipartEncodings := []encodings.MetricMultipartEncoding{
		&otlp.OTLPEncoding{},
		&stef.STEFEncoding{},
		//&stef.STEFUEncoding{},
		&otelarrow.OtelArrowEncoding{},
		//&stef.STEFSEncoding{},
	}

	compressions := []string{"none", "zstd"}

	for _, compression := range compressions {
		for _, dataset := range datasets {
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
			}
		}
	}
}

func TestTEFVeryShortFrames(t *testing.T) {
	input, err := testutils.ReadOTLPFile("testdata/oteldemo-with-histogram.otlp.zst", true)
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
			readRecord, err := tefReader.Read()
			if err == io.EOF {
				break
			}

			tefWriter.Record.CopyFrom(readRecord)

			err = tefWriter.Write()
			if err != nil {
				require.NoError(t, err)
			}
			err = tefWriter.Flush()
			if err != nil {
				require.NoError(t, err)
			}
		}
		fmt.Printf("Compression:        %v\n", tefCompression2str(compression))
		fmt.Printf("Long frames total:  %7d bytes\n", len(tefBytes))
		fmt.Printf("Short frames total: %7d bytes\n", len(shortFrameBuf.Bytes()))
		fmt.Printf("Increase ratio:     %.2fx\n", float64(len(shortFrameBuf.Bytes()))/float64(len(tefBytes)))
	}
}

func tefCompression2str(compression pkg.Compression) any {
	switch compression {
	case pkg.CompressionNone:
		return "none"
	case pkg.CompressionZstd:
		return "zstd"
	}
	panic("unknown compression")
}
