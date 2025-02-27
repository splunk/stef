package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/splunk/stef/go/pkg/idl"
	"github.com/splunk/stef/stefgen/generator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: stefgen <path-to-schema-file>")
		os.Exit(-1)
	}

	fileName := os.Args[1]
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

	fmt.Println("Generating code...")
	g := generator.Generator{OutputDir: wireSchema.PackageName}
	err = g.GenFile(wireSchema)
	if err != nil {
		log.Fatalln(err)
	}
}
