package main

import (
	"encoding/binary"
	"flag"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/tigrannajaryan/stef/stef-go/pkg"

	"github.com/tigrannajaryan/stef/benchmarks/cmd"
	"github.com/tigrannajaryan/stef/benchmarks/encodings/stef"
)

var inputPath = ""
var compressionMethod = ""

func main() {
	flag.StringVar(&inputPath, "input", inputPath, "Input Parquet file or directory. Required.")
	flag.StringVar(&cmd.CpuProfileFileName, "cpuprofile", "", "Write cpu profile to `file`")
	flag.StringVar(
		&compressionMethod, "compression", "", "Compression to use (currently only `zstd` is supported)",
	)
	flag.Parse()

	doneFunc := cmd.SetupProfiling()
	defer doneFunc()

	if inputPath == "" {
		flag.Usage()
		return
	}

	var stats cmd.CmdStats
	stats.Start("OTLP -> STEF")
	newName := cmd.NameNoExt(path.Base(inputPath))
	odir := path.Dir(inputPath)

	convertOTLPtoSTEF(inputPath, path.Join(odir, newName), compressionMethod)
	runtime.GC()

	stats.Stop()

	stats.Print()
}

func convertOTLPtoSTEF(inputFilePath, outputFilePathNoExt, compressionMethod string) {
	all := pmetric.NewMetrics()
	inputBytes, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	for len(inputBytes) > 0 {
		var msgSize uint64
		var msgBytes []byte
		if strings.HasSuffix(inputFilePath, ".otlp") {
			msgSize = uint64(binary.BigEndian.Uint32(inputBytes))
			n := 4
			inputBytes = inputBytes[n:]
			msgBytes = inputBytes[:msgSize]
			inputBytes = inputBytes[msgSize:]
		} else {
			msgBytes = inputBytes
			inputBytes = nil
		}

		unmarshaler := pmetric.ProtoUnmarshaler{}
		msg, err := unmarshaler.UnmarshalMetrics(msgBytes)
		if err != nil {
			log.Fatal(err)
		}
		msg.ResourceMetrics().MoveAndAppendTo(all.ResourceMetrics())
	}

	encoder := stef.STEFEncoding{}
	inmem, err := encoder.FromOTLP(all)
	if err != nil {
		log.Fatal(err)
	}

	switch compressionMethod {
	case "":
	case "zstd":
		encoder.Opts.Compression = pkg.CompressionZstd
	default:
		log.Fatal("Unsupported compression method: " + compressionMethod)
	}

	outputBytes, err := encoder.Encode(inmem)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(outputFilePathNoExt+".stef", outputBytes, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
