package generator

func (g *Generator) oEnums() error {
	for _, enum := range g.compiledSchema.Enums {
		err := g.oEnum(enum)
		if err != nil {
			return err
		}
	}

	return g.lastErr
}

func (g *Generator) oEnum(enum *genEnumDef) error {
	fields := []any{}

	for _, field := range enum.Fields {
		fieldData := map[string]any{
			"Name":  field.Name,
			"Value": field.Value,
		}

		fields = append(fields, fieldData)
	}

	data := map[string]any{
		"EnumName": enum.Name,
		"Fields":   fields,
	}

	if err := g.oTemplates("enum", g.stefSymbol2FileName(enum.Name), data); err != nil {
		return err
	}

	return g.lastErr
}
