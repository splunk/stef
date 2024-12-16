package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type CmdStats struct {
	jobName string

	outputFilePath string

	WorkerCount int
	NoOutput    bool

	startTime, endTime time.Time
	totalDur           time.Duration

	InputDatapointCount   uint64
	OutputDatapointCount  uint64
	DroppedDatapointCount uint64

	InputFileCount  int
	OutputFileCount int

	InputBytes  uint64
	OutputBytes uint64
}

func (c *CmdStats) Start(jobName string) {
	c.jobName = jobName
	c.startTime = time.Now()
	c.WorkerCount = 1

	curDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get current directory: %v", err)
	}
	c.outputFilePath = path.Join(curDir, "results.txt")
}

func (c *CmdStats) Stop() {
	c.endTime = time.Now()
	c.totalDur = c.endTime.Sub(c.startTime)
}

func (c *CmdStats) Print() {
	str := ""
	str += fmt.Sprintf("Job: %s\n\n", c.jobName)
	str += fmt.Sprintf("Duration: %.1f sec\n", c.totalDur.Seconds())

	if c.WorkerCount > 1 {
		str += fmt.Sprintf("Workers: %2d\n", c.WorkerCount)
	}

	str += fmt.Sprintf("Input:\t%14s datapoints\n", NumToStr(c.InputDatapointCount))

	if !c.NoOutput {
		str += fmt.Sprintf("Output:\t%14s datapoints\n", NumToStr(c.OutputDatapointCount))
		str += fmt.Sprintf("Dropped:%14s datapoints\n", NumToStr(c.DroppedDatapointCount))
		if c.InputDatapointCount-c.DroppedDatapointCount != c.OutputDatapointCount {
			str += fmt.Sprintf(
				"Missing:%14s datapoints!!!\n",
				NumToStr(c.InputDatapointCount-c.DroppedDatapointCount-c.OutputDatapointCount),
			)
		}
	}

	str += fmt.Sprintf("Input: %4d files\n", c.InputFileCount)

	if !c.NoOutput {
		str += fmt.Sprintf("Output:%4d files\n", c.OutputFileCount)
	}

	str += fmt.Sprintf("Input:\t%14s Bytes, %5.1f MiB/sec\n", NumToStr(c.InputBytes), c.mibPerSec(c.InputBytes))
	if !c.NoOutput {
		str += fmt.Sprintf("Output:\t%14s Bytes, %5.1f MiB/sec\n", NumToStr(c.OutputBytes), c.mibPerSec(c.OutputBytes))
	}

	str += fmt.Sprintf("Datapoint in:\t%8.1f avg bytes/point\n", float64(c.InputBytes)/float64(c.InputDatapointCount))
	if !c.NoOutput {
		str += fmt.Sprintf(
			"Datapoint out:\t%8.1f avg bytes/point\n", float64(c.OutputBytes)/float64(c.OutputDatapointCount),
		)
	}

	totalDPS := float64(c.InputDatapointCount) / c.totalDur.Seconds()
	str += fmt.Sprintf("Input rate:\t%8s DPS (%8s DPM)\n", NumToStr(uint64(totalDPS)), NumToStr(uint64(totalDPS*60)))

	perWorkerDPS := totalDPS / float64(c.WorkerCount)

	const totalTargetDPM = 3 * 1e9
	const utilization = 0.8 // 80%
	cpuCoresNeeded := totalTargetDPM / (perWorkerDPS * 60) / utilization
	str += fmt.Sprintf(
		"\nCPU Core needed for %.1f billion DPM at %.0f%% utilization: %.1f\n",
		totalTargetDPM/1e9, // billions
		utilization*100,
		cpuCoresNeeded,
	)

	fmt.Print("\n")
	fmt.Print(str)

	// Get current git commit hash
	var commitHash string
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Cannot get git commit hash: %v", err)
	} else {
		commitHash = strings.TrimSpace(string(output))
	}
	str = strings.Repeat("-", 50) + "\n" +
		"Time:   " + time.Now().Local().String() + "\n" +
		"Commit: " + commitHash + "\n" +
		str + "\n\n"

	f, err := os.OpenFile(c.outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Cannot append results to %s: %v", c.outputFilePath, err)
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(str)
	if err != nil {
		log.Fatalf("Cannot write results to %s: %v", c.outputFilePath, err)
	}
}

func (c *CmdStats) mibPerSec(n uint64) float64 {
	return float64(n) / 1024 / 1024 / c.totalDur.Seconds()
}

func NumToStr(n uint64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%v", number.Decimal(n))
}
