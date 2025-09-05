package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/benchmarks/cmd"
	"github.com/splunk/stef/benchmarks/encodings/stef"
)

var inputPath = ""
var compressionMethod = ""

func main() {
	flag.StringVar(&inputPath, "input", inputPath, "Input OTLP Protobuf file. Required.")
	flag.StringVar(
		&compressionMethod, "compression", "", "Compression to use (currently only `zstd` is supported)",
	)
	flag.Parse()

	if inputPath == "" {
		flag.Usage()
		return
	}

	newName := cmd.NameNoExt(path.Base(inputPath))
	odir := path.Dir(inputPath)

	convertOTLPtoSTEF(inputPath, path.Join(odir, newName), compressionMethod)
}

func convertOTLPtoSTEF(inputFilePath, outputFilePathNoExt, compressionMethod string) {
	all, err := testutils.ReadOTLPFile(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	encoder := stef.STEFEncoding{}
	inmem, err := encoder.FromOTLP(all)
	if err != nil {
		log.Fatal(err)
	}

	ext := ".stef"
	switch compressionMethod {
	case "":
	case "zstd":
		encoder.Opts.Compression = pkg.CompressionZstd
		ext += "z"
	default:
		log.Fatal("Unsupported compression method: " + compressionMethod)
	}

	outputBytes, err := encoder.Encode(inmem)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(outputFilePathNoExt+ext, outputBytes, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
