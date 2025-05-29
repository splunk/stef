package generator

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
	fileName := g.stefSymbol2FileName(str.Name + "Writer")

	data := map[string]any{
		"StructName": str.Name,
	}
	if err := g.oTemplates("writer", fileName, data); err != nil {
		return err
	}
	return g.lastErr
}
