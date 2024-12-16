package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tigrannajaryan/stef/stef-go/schema"

	"github.com/tigrannajaryan/stef/stefgen/generator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: stefgen <path-to-schema-file>")
		os.Exit(-1)
	}

	wireJson, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var wireSchema schema.Schema
	err = json.Unmarshal(wireJson, &wireSchema)
	if err != nil {
		panic(err)
	}

	//schemaAst := schema.compileSchema(&wireSchema)
	//schemaAst.resolveRefs()

	fmt.Println("Generating code...")
	g := generator.Generator{OutputDir: wireSchema.PackageName}
	err = g.GenFile(&wireSchema)
	if err != nil {
		log.Fatalln(err)
	}
}
