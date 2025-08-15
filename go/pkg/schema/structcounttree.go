package schema

type structCountTree struct {
	structName   string // for debugging, delete.
	fieldCount   uint
	structFields []structCountTree
}

func schemaToStructCountTree(src *FieldType, dst *structCountTree, stack *recurseStack) {
	switch {
	case src.Primitive != nil: // nothing to do

	case src.StructDef != nil:
		if stack.asMap[src.StructDef.Name] {
			return
		}

		dst.structName = src.StructDef.Name
		dst.fieldCount = uint(len(src.StructDef.Fields))
		//dst.structFields = make([]structCountTree, len(src.StructDef.Fields))

		stack.asStack = append(stack.asStack, src.StructDef.Name)
		stack.asMap[src.StructDef.Name] = true

		for _, field := range src.StructDef.Fields {
			subDst := structCountTree{}
			schemaToStructCountTree(&field.FieldType, &subDst, stack)
			if subDst.fieldCount != 0 {
				dst.structFields = append(dst.structFields, subDst)
			}
		}

		stack.asStack = stack.asStack[:len(stack.asStack)-1]
		delete(stack.asMap, src.StructDef.Name)

	case src.Array != nil:
		schemaToStructCountTree(&src.Array.ElemType, dst, stack)

	case src.MultimapDef != nil:
		if stack.asMap[src.MultimapDef.Name] {
			return
		}

		stack.asStack = append(stack.asStack, src.MultimapDef.Name)
		stack.asMap[src.MultimapDef.Name] = true

		schemaToStructCountTree(&src.MultimapDef.Key.Type, dst, stack)
		schemaToStructCountTree(&src.MultimapDef.Value.Type, dst, stack)

		stack.asStack = stack.asStack[:len(stack.asStack)-1]
		delete(stack.asMap, src.MultimapDef.Name)

	default:
		panic("unknown FieldType")
	}
}
