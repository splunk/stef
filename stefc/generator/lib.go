package generator

import "sort"

func (g *Generator) oRustLib() error {
	if g.Lang != LangRust {
		return nil
	}

	modules := map[string]struct{}{}
	add := func(module string) {
		if module == "" {
			return
		}
		modules[module] = struct{}{}
	}

	add(g.stefSymbol2FileName("ModifiedFields"))

	for _, struc := range g.compiledSchema.Structs {
		add(g.stefSymbol2FileName(struc.Name))
		if struc.IsRoot {
			base := g.stefSymbol2FileName(struc.Name + "Writer")
			add(base)
			add(base + "_test")
			add(g.stefSymbol2FileName(struc.Name + "Reader"))
		}
	}

	for _, mm := range g.compiledSchema.Multimaps {
		add(g.stefSymbol2FileName(mm.Name))
	}

	for _, struc := range g.compiledSchema.Structs {
		for _, field := range struc.Fields {
			if arr, ok := field.Type.(*genArrayTypeRef); ok {
				add(g.stefSymbol2FileName(arr.IDLMangledName()))
			}
		}
	}
	for _, mm := range g.compiledSchema.Multimaps {
		if arr, ok := mm.Key.Type.(*genArrayTypeRef); ok {
			add(g.stefSymbol2FileName(arr.IDLMangledName()))
		}
		if arr, ok := mm.Value.Type.(*genArrayTypeRef); ok {
			add(g.stefSymbol2FileName(arr.IDLMangledName()))
		}
	}

	add(g.stefSymbol2FileName("ReaderState"))
	add(g.stefSymbol2FileName("WriterState"))

	for _, enum := range g.compiledSchema.Enums {
		add(g.stefSymbol2FileName(enum.Name))
	}

	if g.GenTools {
		add(g.stefSymbol2FileName("Tools") + "_test")
	}

	add(g.stefSymbol2FileName("Common"))

	moduleList := make([]string, 0, len(modules))
	for module := range modules {
		moduleList = append(moduleList, module)
	}
	sort.Strings(moduleList)

	data := map[string]any{
		"Modules": moduleList,
	}

	return g.oTemplates("lib", "lib", data)
}
