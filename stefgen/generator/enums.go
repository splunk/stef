package generator

import "strings"

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

	templateName := "enum.go.tmpl"

	if err := g.oTemplate(templateName, strings.ToLower(enum.Name), data); err != nil {
		return err
	}

	return g.lastErr
}
