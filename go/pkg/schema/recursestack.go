package schema

type recursable interface {
	SetRecursive()
}

type recurseStack struct {
	// Names of types in the path of the type tree currently being searched
	// from top to bottom of the tree. Root struct is the first element.
	asStack []string

	// Same types, but as a map for fast search by name.
	asMap map[string]bool

	// Fields of types through which the types in asStack are linked.
	// fields[i] is a field in asStack[i] type.
	fields []recursable
}
