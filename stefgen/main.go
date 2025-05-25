package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/splunk/stef/go/pkg/idl"
	"github.com/splunk/stef/stefgen/generator"
)

func main() {
	lang := flag.String("lang", "", "Target language for code generation. Currently only go is supported.")
	outDir := flag.String("outdir", "", "Output directory.")
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

	packageComponents := strings.Split(wireSchema.PackageName, ".")
	packageName := packageComponents[len(packageComponents)-1]

	destDir := path.Join(*outDir, packageName)
	fmt.Printf("Generating %s code to %s\n", *lang, destDir)

	g := generator.Generator{
		OutputDir: destDir,
		Lang:      generator.Lang(*lang),
	}
	err = g.GenFile(wireSchema)
	if err != nil {
		log.Fatalln(err)
	}
}
