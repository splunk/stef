package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/benchmarks/cmd"
	"github.com/splunk/stef/go/otel/oteltef"
)

var InputPath = ""

func main() {
	flag.StringVar(&InputPath, "input", InputPath, "Input TEF `file` path. Required.")
	flag.StringVar(&cmd.CpuProfileFileName, "cpuprofile", "", "Write cpu profile to `file`.")
	// Parse the flag
	flag.Parse()

	if InputPath == "" {
		flag.Usage()
		return
	}

	var cmdStats cmd.CmdStats
	cmdStats.Start("STEF Bench, " + InputPath)
	cmdStats.NoOutput = true

	stefBench(InputPath)
}

func openStef(inputFilePath string) (*oteltef.MetricsReader, func()) {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Cannot read file %s: %v", inputFilePath, err)
	}
	reader, err := oteltef.NewMetricsReader(bufio.NewReaderSize(inputFile, 64*1024))
	if err != nil {
		log.Fatalf("Error reading file %s: %v", inputFilePath, err)
	}
	return reader, func() { inputFile.Close() }
}

func stefBench(inputFilePath string) {
	fmt.Printf("Benchmarking %s...\n", inputFilePath)
	readBench(inputFilePath)
	writeBench(inputFilePath)
}

func readBench(inputFilePath string) {
	reader, closer := openStef(inputFilePath)
	defer closer()

	recordCount := 0
	start := time.Now()
	for {
		record, err := reader.Read()
		record = record
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading from %s: %v", inputFilePath, err)
		}
		recordCount++
		if recordCount%1000000 == 0 {
			fmt.Printf("Read %d points\t\t\r", recordCount)
		}
	}
	dur := time.Since(start)
	fmt.Printf(
		"Read %d points in %.3f sec, %.0f points/sec\n", recordCount, dur.Seconds(), float64(recordCount)/dur.Seconds(),
	)
}

func writeBench(inputFilePath string) {
	reader, closer := openStef(inputFilePath)
	defer closer()

	outputFile, err := os.CreateTemp(os.TempDir(), "stefbench.stef")
	if err != nil {
		log.Fatalf("Cannot create file %s: %v", outputFile.Name(), err)
	}
	defer outputFile.Close()

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := oteltef.NewMetricsWriter(
		outputBuf, pkg.WriterOptions{
			Compression:         reader.Header().Compression,
			TimestampMultiplier: reader.Header().TimestampMultiplier,
		},
	)
	if err != nil {
		log.Fatalf("Error writting to buffer: %v", err)
	}

	fmt.Printf("Copying from %s to temp file %s...\n", inputFilePath, outputFile.Name())

	doneFunc := cmd.SetupProfiling()
	defer doneFunc()

	recordCount := 0
	start := time.Now()
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading from %s: %v", inputFilePath, err)
		}

		if record.IsEnvelopeModified() {
			writer.Record.Envelope().CopyFrom(record.Envelope())
		}
		if record.IsMetricModified() {
			writer.Record.Metric().CopyFrom(record.Metric())
		}
		if record.IsResourceModified() {
			writer.Record.Resource().CopyFrom(record.Resource())
		}
		if record.IsScopeModified() {
			writer.Record.Scope().CopyFrom(record.Scope())
		}
		if record.IsAttributesModified() {
			writer.Record.Attributes().CopyFrom(record.Attributes())
		}
		if record.IsPointModified() {
			writer.Record.Point().CopyFrom(record.Point())
		}
		err = writer.Write()
		if err != nil {
			panic("Cannot write:" + err.Error())
		}

		recordCount++
		if recordCount%1000000 == 0 {
			fmt.Printf("Copied %d points\t\t\r", recordCount)
		}
	}
	err = writer.Flush()
	if err != nil {
		panic("Cannot write:" + err.Error())
	}
	dur := time.Since(start)
	fmt.Printf(
		"Copied %d points in %.3f sec, %.0f points/sec\n", recordCount, dur.Seconds(),
		float64(recordCount)/dur.Seconds(),
	)

	_, err = outputFile.Write(outputBuf.Bytes())
	if err != nil {
		log.Fatalf("Error writting to %s: %v", outputFile.Name(), err)
	}
	/*
		fmt.Printf("Verifying copied data... ")

		reader, closer = openStef(inputFilePath)
		defer closer()

		reader2, closer2 := openStef(outputFile.Name())
		defer closer2()

		var err2 error
		for {
			var record *oteltef.Metrics
			var record2 *oteltef.Metrics
			record, err = reader.Read()
			record2, err2 = reader2.Read()
			if err == io.EOF {
				break
			}
			if err2 != nil {
				log.Fatalf("Error reading second stream: %v", err2)
			}
			if record.Changed != record2.Changed {
				log.Fatalf("Changed don't match")
			}
			if CmpAttrs(record.Envelope, record2.Envelope) != 0 {
				log.Fatalf("Envelopes don't match")
			}
			if types.CmpMetric(record.Metric, record2.Metric) != 0 {
				log.Fatalf("Metrics don't match")
			}
			if types.CmpResource(record.Resource, record2.Resource) != 0 {
				log.Fatalf("Metrics don't match")
			}
			if types.CmpScope(record.Scope, record2.Scope) != 0 {
				log.Fatalf("Metrics don't match")
			}
			if types.CmpAttrs(record.Attrs, record2.Attrs) != 0 {
				log.Fatalf("Metrics don't match")
			}
			if !equalValues(record.Point, record2.Point) {
				log.Fatalf("Values don't match")
			}
		}
		if err2 != io.EOF {
			log.Fatalf("Expected EOF")
		}
		fmt.Printf("All data matches.\n")
	*/
}

func equalSlices[T comparable](s1, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func equalFloatSlices(s1, s2 []float64) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			if !math.IsNaN(s1[i]) || !math.IsNaN(s2[i]) {
				return false
			}
		}
	}
	return true
}

/*
func equalValues(v1, v2 *types.TimedPoint) bool {
	if v1.StartTimestampUnixNano != v2.StartTimestampUnixNano {
		return false
	}
	if v1.TimestampUnixNano != v2.TimestampUnixNano {
		return false
	}
	if !reflect.DeepEqual(v1.Exemplars, v2.Exemplars) {
		return false
	}
	if !equalSlices(v1.Int64Vals, v2.Int64Vals) {
		return false
	}
	if !equalFloatSlices(v1.Float64Vals, v2.Float64Vals) {
		return false
	}
	return true
}
*/
