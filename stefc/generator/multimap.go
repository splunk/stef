package generator

func (g *Generator) oMultimaps() error {
	var multimaps []*genStructFieldDef
	for _, struc := range g.compiledSchema.Structs {
		multimaps = append(multimaps, g.getMultiMaps(struc)...)
	}

	for _, mm := range g.compiledSchema.Multimaps {
		if err := g.oMultimap(mm); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) getMultiMaps(struc *genStructDef) (ret []*genStructFieldDef) {
	for _, field := range struc.Fields {
		if _, ok := field.Type.(*genMultimapTypeRef); ok {
			ret = append(ret, field)
		}
	}
	return ret
}

type MultimapTemplateModel struct {
	PackageName     string
	MultimapName    string
	Key             genFieldDef
	Value           genFieldDef
	KeyStoreByPtr   bool
	ValueStoreByPtr bool
}

func (g *Generator) oMultimap(multimap *genMapDef) error {
	mapType := g.compiledSchema.Multimaps[multimap.Name]

	data := map[string]any{
		"PackageName":     g.compiledSchema.PackageNameStr,
		"MultimapName":    mapType.Name,
		"Key":             mapType.Key,
		"Value":           mapType.Value,
		"KeyStoreByPtr":   !mapType.Key.Type.IsPrimitive() && mapType.Key.Type.DictName() != "",
		"ValueStoreByPtr": !mapType.Value.Type.IsPrimitive() && mapType.Value.Type.DictName() != "",
	}

	return g.oTemplates("multimap", g.stefSymbol2FileName(multimap.Name), data)
}
