package schema

import (
	"math/rand/v2"
	"sort"
)

// ShrinkRandomly will deterministically, pseudo-randomly shrink the provided
// schema by removing the last field from one of the structs or oneofs.
// This function is used for testing schema changes.
// Returns true if the schema was changed, false if there was nothing to shrink,
// i.e. the schema has no fields.
func ShrinkRandomly(r *rand.Rand, schem *Schema) bool {
	totalFieldCount := 0
	var structNames []string
	for structName := range schem.Structs {
		structNames = append(structNames, structName)
		totalFieldCount += len(schem.Structs[structName].Fields)
	}
	if totalFieldCount == 0 {
		// Nothing to shrink
		return false
	}

	sort.Strings(structNames)

	for {
		structName := structNames[r.IntN(len(structNames))]
		str := schem.Structs[structName]
		if shrinkStruct(r, schem, str) {
			return true
		}
	}
}

func shrinkStruct(r *rand.Rand, schem *Schema, str *Struct) bool {
	if r.IntN(10) == 0 && len(str.Fields) > 0 {
		str.Fields = str.Fields[0 : len(str.Fields)-1]
		return true
	}

	for _, field := range str.Fields {
		if field.Struct != "" {
			if r.IntN(3) == 0 {
				childStruct := schem.Structs[field.Struct]
				changed := shrinkStruct(r, schem, childStruct)
				if changed {
					return true
				}
			}
		}
	}

	return false
}
