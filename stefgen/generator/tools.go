package generator

func (g *Generator) oTools() error {
	if !g.genTools {
		return nil
	}

	rootStructs := []string{}

	for _, struc := range g.compiledSchema.Structs {
		if struc.IsRoot {
			rootStructs = append(rootStructs, struc.Name)
		}
	}

	data := map[string]any{
		"RootStructs": rootStructs,
	}
	return g.oTemplates("tools", g.stefSymbol2FileName("Tools"), data)
}
