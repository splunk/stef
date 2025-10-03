package generator

func (g *Generator) oReaderWriterState() error {
	data := map[string]any{
		"Dicts":    g.getDicts(),
		"Encoders": g.getEncoders(),
	}
	if err := g.oTemplates("allreaderstate", g.stefSymbol2FileName("ReaderState"), data); err != nil {
		return err
	}
	if err := g.oTemplates("allwriterstate", g.stefSymbol2FileName("WriterState"), data); err != nil {
		return err
	}
	return g.lastErr
}

func (g *Generator) getDicts() (ret map[string]genFieldTypeRef) {
	ret = map[string]genFieldTypeRef{}

	for _, struc := range g.compiledSchema.Structs {
		if struc.Dict != "" {
			ret[struc.Dict] = &genStructTypeRef{
				Name: struc.Name,
				Def:  struc,
				Lang: g.Lang,
			}
		}

		for _, field := range struc.Fields {
			dictName := field.Type.DictName()
			if dictName != "" {
				ret[dictName] = field.Type
			}
		}
	}
	for _, m := range g.compiledSchema.Multimaps {
		if m.Key.Type.DictName() != "" {
			ret[m.Key.Type.DictName()] = m.Key.Type
		}
		if m.Value.Type.DictName() != "" {
			ret[m.Value.Type.DictName()] = m.Value.Type
		}
	}
	return ret
}

func (g *Generator) getEncoders() (ret map[string]bool) {
	ret = map[string]bool{}

	for _, struc := range g.compiledSchema.Structs {
		ret[struc.Name] = true
		for _, field := range struc.Fields {
			if !field.Type.IsPrimitive() {
				ret[field.Type.EncoderType()] = true
			}
		}
	}
	for _, m := range g.compiledSchema.Multimaps {
		if !m.Key.Type.IsPrimitive() {
			ret[m.Key.Type.EncoderType()] = true
		}
		if !m.Value.Type.IsPrimitive() {
			ret[m.Value.Type.EncoderType()] = true
		}
	}
	return ret
}
