package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/splunk/stef/go/pkg/idl"
	"github.com/splunk/stef/stefc/generator"
)

func main() {
	lang := flag.String("lang", "", "Target language for code generation. Supported languages: go, java, rust.")
	outDir := flag.String("outdir", "", "Output directory.")
	testOutDir := flag.String(
		"testoutdir", "",
		"Output directory for test files. If unspecified, it defaults to outdir.",
	)
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: stefc [flags] <path-to-schema-file>")
		os.Exit(-1)
	}

	switch generator.Lang(*lang) {
	case generator.LangGo:
	case generator.LangJava:
	case generator.LangRust:
	default:
		fmt.Printf("Error: Unsupported language %s. Supported languages: go, java, rust.\n", *lang)
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

	for _, msg := range parser.Messages() {
		fmt.Println(msg.String())
	}

	wireSchema := parser.Schema()

	fmt.Printf("Generating %s code to %s\n", *lang, *outDir)

	g := generator.Generator{
		SchemaContent: schemaContent,
		OutputDir:     *outDir,
		TestOutputDir: *testOutDir,
		Lang:          generator.Lang(*lang),
		GenTools:      *testOutDir != "",
	}
	err = g.GenFile(wireSchema)
	if err != nil {
		log.Fatalln(err)
	}
}
