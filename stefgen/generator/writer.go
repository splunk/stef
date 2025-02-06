package generator

import (
	"bytes"
	"strings"
)

func (g *Generator) oWriters() error {
	for _, str := range g.compiledSchema.Structs {
		if str.IsRoot {
			if err := g.oWriter(str); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) oWriter(str *genStructDef) error {
	prunedSchema, err := g.schema.PrunedForRoot(str.Name)
	if err != nil {
		return err
	}
	wireSchema := prunedSchema.ToWire()

	fileName := strings.ToLower(str.Name) + "writer"

	var wireBin bytes.Buffer
	if err := wireSchema.Serialize(&wireBin); err != nil {
		return err
	}

	data := map[string]any{
		"Schema":     string(wireBin.Bytes()),
		"StructName": str.Name,
	}
	if err := g.oTemplate("writer.go.tmpl", fileName+".go", data); err != nil {
		return err
	}

	if err := g.oTemplate("writer_test.go.tmpl", fileName+"_test.go", data); err != nil {
		return err
	}
	return g.lastErr
}
