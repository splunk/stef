package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/splunk/stef/go/pkg/idl"
	"github.com/splunk/stef/stefgen/generator"
)

func main() {
	lang := flag.String("lang", "", "Target language for code generation. Currently only go is supported.")
	outDir := flag.String("outdir", "", "Output directory.")
	testOutDir := flag.String(
		"testoutdir", "",
		"Output directory for test files. If unspecified, it defaults to outdir. Can be used with --lang=java only.",
	)
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: stefgen [flags] <path-to-schema-file>")
		os.Exit(-1)
	}

	switch generator.Lang(*lang) {
	case generator.LangGo:
	case generator.LangJava:
	default:
		fmt.Printf("Error: Unsupported language %s. Currently only go is supported.\n", *lang)
		os.Exit(-1)
	}

	fileName := flag.Arg(0)
	schemaContent, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	lexer := idl.NewLexer(bytes.NewBuffer(schemaContent))
	parser := idl.NewParser(lexer, path.Base(fileName))
	err = parser.Parse()
	if err != nil {
		log.Fatalln(err)
	}
	wireSchema := parser.Schema()

	fmt.Printf("Generating %s code to %s\n", *lang, *outDir)

	g := generator.Generator{
		SchemaContent: schemaContent,
		OutputDir:     *outDir,
		TestOutputDir: *testOutDir,
		Lang:          generator.Lang(*lang),
	}
	err = g.GenFile(wireSchema)
	if err != nil {
		log.Fatalln(err)
	}
}
