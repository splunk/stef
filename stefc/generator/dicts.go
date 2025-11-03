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

// getEncoders collects all encoder types used in the schema.
func (g *Generator) getEncoders() (ret map[string]bool) {
	ret = map[string]bool{}
	visited := map[string]bool{}

	var collectEncodersFromType func(t genFieldTypeRef)
	collectEncodersFromType = func(t genFieldTypeRef) {
		if t == nil {
			return
		}
		id := t.IDLMangledName()
		if visited[id] {
			return
		}
		visited[id] = true
		if !t.IsPrimitive() {
			ret[t.EncoderType()] = true
		}
		switch tt := t.(type) {
		case *genStructTypeRef:
			if tt.Def != nil {
				for _, f := range tt.Def.Fields {
					collectEncodersFromType(f.Type)
				}
			}
		case *genArrayTypeRef:
			collectEncodersFromType(tt.ElemType)
		case *genMultimapTypeRef:
			if tt.Def != nil {
				collectEncodersFromType(tt.Def.Key.Type)
				collectEncodersFromType(tt.Def.Value.Type)
			}
		}
	}

	for _, struc := range g.compiledSchema.Structs {
		collectEncodersFromType(
			&genStructTypeRef{
				Name: struc.Name,
				Def:  struc,
				Lang: g.Lang,
			},
		)
	}
	for _, m := range g.compiledSchema.Multimaps {
		collectEncodersFromType(
			&genMultimapTypeRef{
				Name: m.Name,
				Def:  m,
				Lang: g.Lang,
			},
		)
	}
	return ret
}
