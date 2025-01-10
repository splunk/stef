package generator

import "strings"

func (g *Generator) oReaders() error {
	for _, str := range g.compiledSchema.Structs {
		if str.IsRoot {
			if err := g.oReader(str); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) oReader(str *genStructDef) error {
	fileName := strings.ToLower(str.Name) + "reader.go"
	data := map[string]any{
		"StructName": str.Name,
	}
	if err := g.oTemplate("reader.go.tmpl", fileName, data); err != nil {
		return err
	}
	//if err := g.formatAndWriteToFile(); err != nil {
	//	return err
	//}
	return g.lastErr
}
