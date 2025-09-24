package generator

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
	data := map[string]any{
		"StructName": str.Name,
	}
	fileName := g.stefSymbol2FileName(str.Name + "Reader")
	if err := g.oTemplates("reader", fileName, data); err != nil {
		return err
	}
	return g.lastErr
}
