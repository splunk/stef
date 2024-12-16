package generator

import (
	"encoding/json"
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
	wireSchema, err := g.schema.PrunedForRoot(str.Name)
	if err != nil {
		return err
	}

	fileName := strings.ToLower(str.Name) + "writer"

	wireJson, err := json.Marshal(wireSchema)
	if err != nil {
		return err
	}

	data := map[string]any{
		"Schema":     string(wireJson),
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
