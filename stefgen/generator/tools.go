package generator

func (g *Generator) oTools() error {
	if !g.genTools {
		return nil
	}

	structs := []string{}

	for _, struc := range g.compiledSchema.Structs {
		if struc.IsRoot {
			structs = append(structs, struc.Name)
		}
	}

	data := map[string]any{
		"Structs": structs,
	}
	return g.oTemplates("tools", g.stefSymbol2FileName("Tools"), data)
}
