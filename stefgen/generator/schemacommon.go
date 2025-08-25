package generator

func (g *Generator) oCommon() error {
	data := map[string]any{
		"Schema":  g.SchemaContent,
		"Structs": g.compiledSchema.Structs,
	}
	return g.oTemplates("common", g.stefSymbol2FileName("Common"), data)
}
