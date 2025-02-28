package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"time"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/benchmarks/cmd"
)

var InputPath = "./"

type MetricsStats struct {
	FileCount      int
	TotalByteSize  uint64
	DatapointCount uint64
	MinTimestamp   uint64
	MaxTimestamp   uint64

	UniqueMetricNames map[string]bool
	UniqueTraceIds    map[string]bool
	UniqueSpanIds     map[string]bool

	TimeseriesDurSum uint64
	TimeseriesDurMin uint64
	TimeseriesDurMax uint64
	TimeseriesCount  uint64

	ExemplarCount uint64
}

func NewStats() *MetricsStats {
	return &MetricsStats{
		TimeseriesDurMin:  math.MaxUint64,
		UniqueMetricNames: map[string]bool{},
		UniqueTraceIds:    map[string]bool{},
		UniqueSpanIds:     map[string]bool{},
	}
}

func main() {
	flag.StringVar(&InputPath, "input", InputPath, "Input STEF file name or directory")
	flag.StringVar(&cmd.CpuProfileFileName, "cpuprofile", "", "Write cpu profile to `file`")
	// Parse the flag
	flag.Parse()

	doneFunc := cmd.SetupProfiling()
	defer doneFunc()

	var cmdStats cmd.CmdStats
	cmdStats.Start("STEF Stats, " + InputPath)
	cmdStats.NoOutput = true

	fileStats := NewStats()
	inputFiles := cmd.ListInputFiles(InputPath, &cmdStats)
	for i := 0; i < len(inputFiles); i++ {
		ext := path.Ext(inputFiles[i].FullPath)
		if ext == ".stef" {
			collectSTEFFileStats(inputFiles[i].FullPath, fileStats)
		}
		cmdStats.InputDatapointCount += fileStats.DatapointCount
	}
	fileStats.FileCount = len(inputFiles)

	cmdStats.Stop()
	cmdStats.Print()

	fmt.Printf("\n")
	fmt.Printf("Unique Metric names: %4d\n", len(fileStats.UniqueMetricNames))
	fmt.Printf("\n")
	fmt.Printf("Timeseries\n")
	//fmt.Printf("Unique Count:  %10s\n", NumToStr(uint64(len(fileStats.UniqueOrgTsids))))
	fmt.Printf("Total Count:   %10s\n", cmd.NumToStr(fileStats.TimeseriesCount))
	fmt.Printf("Avg datapoints: %.1f\n", float64(fileStats.DatapointCount)/float64(fileStats.TimeseriesCount))
	fmt.Printf(
		"Duration: min=%s, avg=%s, max=%s\n",
		time.Duration(fileStats.TimeseriesDurMin).String(),
		time.Duration(fileStats.TimeseriesDurSum/fileStats.TimeseriesCount).String(),
		time.Duration(fileStats.TimeseriesDurMax).String(),
	)
	fmt.Printf("\n")
	fmt.Printf("Min timestamp: %s\n", time.Unix(0, int64(fileStats.MinTimestamp)).String())
	fmt.Printf("Max timestamp: %s\n", time.Unix(0, int64(fileStats.MaxTimestamp)).String())
	//fmt.Printf("\n")
	//fmt.Printf("Blocks\n")
	//fmt.Printf("Metrics:    %d\n", fileStats.BlockStats.Metrics)
	//fmt.Printf("Metadatas:  %d\n", fileStats.BlockStats.Metadatas)
	//fmt.Printf("Resources:  %d\n", fileStats.BlockStats.Resources)
	//fmt.Printf("Scopes:     %d\n", fileStats.BlockStats.Scopes)
	//fmt.Printf("Attrs:      %d\n", fileStats.BlockStats.Attrs)
	//fmt.Printf("Datapoints: %d\n", fileStats.BlockStats.Datapoints)

	if fileStats.ExemplarCount > 0 {
		fmt.Printf("\n")
		fmt.Printf("Exemplars:      %d\n", fileStats.ExemplarCount)
		fmt.Printf("Unique traceid: %d\n", len(fileStats.UniqueTraceIds))
		fmt.Printf("Unique spanid:  %d\n", len(fileStats.UniqueSpanIds))
	}
}

func collectSTEFFileStats(inputFilePath string, fileStats *MetricsStats) {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Cannot read file %s: %v", inputFilePath, err)
	}
	defer func() { _ = inputFile.Close() }()
	reader, err := oteltef.NewMetricsReader(bufio.NewReaderSize(inputFile, 64*1024))
	if err != nil {
		log.Fatalf("Error reading file %s: %v", inputFilePath, err)
	}

	fmt.Printf("Calculating stats on %s...\r", inputFilePath)

	if err := CollectRecordStats(reader, fileStats); err != nil {
		log.Fatalf("Error reading file %s: %v", inputFilePath, err)
	}
}

func CollectRecordStats(reader *oteltef.MetricsReader, fileStats *MetricsStats) error {
	tsMinTimestamp := uint64(math.MaxUint64)
	tsMaxTimestamp := uint64(0)
	var timeseriesDPCount uint64

	finishTimeseries := func() {
		if timeseriesDPCount > 0 {
			// Calc stats for previous timeseries.
			timeseriesDur := tsMaxTimestamp - tsMinTimestamp
			fileStats.TimeseriesDurSum += timeseriesDur
			fileStats.TimeseriesDurMin = min(timeseriesDur, fileStats.TimeseriesDurMin)
			fileStats.TimeseriesDurMax = max(timeseriesDur, fileStats.TimeseriesDurMax)
			fileStats.TimeseriesCount++
		}
	}

	for {
		err := reader.Read(pkg.ReadOptions{})
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		record := &reader.Record

		fileStats.DatapointCount++
		fileStats.UniqueMetricNames[string(record.Metric().Name())] = true

		if record.Point().Timestamp() > fileStats.MaxTimestamp {
			fileStats.MaxTimestamp = record.Point().Timestamp()
		}
		if record.Point().Timestamp() < fileStats.MinTimestamp || fileStats.MinTimestamp == 0 {
			fileStats.MinTimestamp = record.Point().Timestamp()
		}

		if record.IsAttributesModified() {
			finishTimeseries()

			tsMinTimestamp = math.MaxUint64
			tsMaxTimestamp = 0
			timeseriesDPCount = 0
		}
		tsMinTimestamp = min(tsMinTimestamp, record.Point().Timestamp())
		tsMaxTimestamp = max(tsMaxTimestamp, record.Point().Timestamp())
		timeseriesDPCount++

		for i := 0; i < record.Point().Exemplars().Len(); i++ {
			exemplar := record.Point().Exemplars().At(i)
			fileStats.ExemplarCount++
			fileStats.UniqueTraceIds[string(exemplar.TraceID())] = true
			fileStats.UniqueSpanIds[string(exemplar.SpanID())] = true
		}
	}
	finishTimeseries()

	//fileStats.MinTimestamp *= reader.Header().TimestampMultiplier
	//fileStats.MaxTimestamp *= reader.Header().TimestampMultiplier
	//fileStats.TimeseriesDurSum *= reader.Header().TimestampMultiplier
	//fileStats.TimeseriesDurMin *= reader.Header().TimestampMultiplier
	//fileStats.TimeseriesDurMax *= reader.Header().TimestampMultiplier

	//blockStats := reader.Stats()
	//fileStats.BlockStats.Metrics += blockStats.Metrics
	//fileStats.BlockStats.Metadatas += blockStats.Metadatas
	//fileStats.BlockStats.Resources += blockStats.Resources
	//fileStats.BlockStats.Scopes += blockStats.Scopes
	//fileStats.BlockStats.Attrs += blockStats.Attrs
	//fileStats.BlockStats.Datapoints += blockStats.Datapoints

	return nil
}
