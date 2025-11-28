package generator

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg/idl"
	"github.com/splunk/stef/go/pkg/schema"
)

func testSchema(t *testing.T, schemaContent []byte, schemaFileName string, failOnTest bool) {
	// Parse the schema
	lexer := idl.NewLexer(bytes.NewBuffer(schemaContent))
	parser := idl.NewParser(lexer, path.Base(schemaFileName))
	err := parser.Parse()
	require.NoError(t, err)

	parsedSchema := parser.Schema()

	// Clean Go directory
	goDir := path.Join("testdata", "out", path.Base(schemaFileName))
	err = os.RemoveAll(goDir)
	require.NoError(t, err)

	// Generate the Go code
	genGo := Generator{
		SchemaContent: schemaContent,
		OutputDir:     goDir,
		Lang:          LangGo,
		genTools:      true, // Generate testing tools
	}

	err = genGo.GenFile(parsedSchema)
	require.NoError(t, err)

	fmt.Printf("Testing generated code in %s\n", genGo.OutputDir)

	// Run tests in the generated code
	cmd := exec.Command("go", "test", "-v", genGo.OutputDir+"/...")
	cmd.Dir = genGo.OutputDir
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", stdoutStderr)
		if failOnTest {
			t.Fatal(err)
		} else {
			t.Skipf("Warning: go test failed: %v\n", err)
			return
		}
	}

	// Clean Java directory
	javaDir := path.Join("../../java/src/test/java")
	packageDir := path.Join(javaDir, strings.Join(parsedSchema.PackageName, "/"))
	err = os.RemoveAll(packageDir)
	require.NoError(t, err)

	// Generate the Java code
	genJava := Generator{
		SchemaContent: schemaContent,
		OutputDir:     javaDir,
		TestOutputDir: javaDir,
		Lang:          LangJava,
		genTools:      true, // Generate testing tools
	}

	err = genJava.GenFile(parsedSchema)
	require.NoError(t, err)
}

func TestGenerate(t *testing.T) {
	// Get the list of files in "testdata" directory
	files, err := filepath.Glob("testdata/*.stef")
	require.NoError(t, err)

	for _, file := range files {
		t.Run(
			file, func(t *testing.T) {
				// Read the schema file
				schemaContent, err := os.ReadFile(file)
				require.NoError(t, err)
				testSchema(t, schemaContent, file, true)
			},
		)
	}
}

type schemaGenerator struct {
	multimapNames []string
	structNames   []string
	enumNames     []string
	random        *rand.Rand
}

func (g *schemaGenerator) generate() *schema.Schema {
	sch := &schema.Schema{}
	sch.PackageName = []string{"com", "example", "gentest", "randomized"}

	// Generate multimap names
	multiMapCount := g.random.IntN(5)
	sch.Multimaps = map[string]*schema.Multimap{}
	for i := 0; i < multiMapCount; i++ {
		g.multimapNames = append(g.multimapNames, fmt.Sprintf("MultiMap%d", i+1))
	}

	// Generate struct names
	structCount := g.random.IntN(10) + 1
	sch.Structs = map[string]*schema.Struct{}
	for i := 0; i < structCount; i++ {
		g.structNames = append(g.structNames, fmt.Sprintf("Struct%d", i+1))
	}

	// Generate enum names
	enumCount := g.random.IntN(5)
	sch.Enums = map[string]*schema.Enum{}
	for i := 0; i < enumCount; i++ {
		g.enumNames = append(g.enumNames, fmt.Sprintf("Enum%d", i+1))
	}

	// Generate multimaps
	for _, mmName := range g.multimapNames {
		mm := &schema.Multimap{
			Name:  mmName,
			Key:   schema.MultimapField{Type: g.genRandomType(true)},
			Value: schema.MultimapField{Type: g.genRandomType(true)},
		}
		sch.Multimaps[mmName] = mm
	}

	// Generate structs
	hasRoot := false
	for _, structName := range g.structNames {
		fieldCount := g.random.IntN(5) + 1
		fields := make([]*schema.StructField, fieldCount)
		for i := 0; i < fieldCount; i++ {
			typ := g.genRandomType(true)
			field := &schema.StructField{
				Name:      fmt.Sprintf("Field%d", i+1),
				FieldType: typ,
				Optional:  g.random.IntN(2) == 0,
			}
			if typ.Struct != "" {
				// TODO: detect and avoid cycles. For now mark it optional since cycles are
				// allowed with optional fields.
				field.Optional = true
			}
			fields[i] = field
		}
		str := &schema.Struct{
			Name:   structName,
			Fields: fields,
		}
		if g.random.IntN(structCount) == 0 {
			str.IsRoot = true
			hasRoot = true
		}
		sch.Structs[structName] = str
	}
	if !hasRoot {
		// Ensure at least one root struct
		for _, str := range sch.Structs {
			str.IsRoot = true
			break
		}
	}

	// Generate enums
	for _, enumName := range g.enumNames {
		fieldCount := g.random.IntN(5) + 1
		fields := make([]schema.EnumField, fieldCount)
		for i := 0; i < fieldCount; i++ {
			fields[i] = schema.EnumField{
				Name:  fmt.Sprintf("Item%d", i+1),
				Value: uint64(i),
			}
		}
		sch.Enums[enumName] = &schema.Enum{
			Name:   enumName,
			Fields: fields,
		}
	}

	return sch
}

func (g *schemaGenerator) genRandomType(allowArray bool) schema.FieldType {
	// Type categories: 0=primitive, 1=multimap, 2=enum, 3=array, 4=struct
	choices := []int{0} // Always allow primitive
	if allowArray {
		choices = append(choices, 3)
	}
	if len(g.multimapNames) > 0 {
		choices = append(choices, 1)
	}
	if len(g.enumNames) > 0 {
		choices = append(choices, 2)
	}
	if len(g.structNames) > 0 {
		choices = append(choices, 4)
	}
	cat := choices[g.random.IntN(len(choices))]

	switch cat {
	case 0: // Primitive
		prims := []schema.PrimitiveFieldType{
			schema.PrimitiveTypeInt64,
			schema.PrimitiveTypeUint64,
			schema.PrimitiveTypeFloat64,
			schema.PrimitiveTypeBool,
			schema.PrimitiveTypeString,
			schema.PrimitiveTypeBytes,
		}
		prim := prims[g.random.IntN(len(prims))]
		return schema.FieldType{Primitive: &schema.PrimitiveType{Type: prim}}
	case 1: // Multimap
		name := g.multimapNames[g.random.IntN(len(g.multimapNames))]
		return schema.FieldType{MultiMap: name}
	case 2: // Enum
		name := g.enumNames[g.random.IntN(len(g.enumNames))]
		return schema.FieldType{Enum: name}
	case 3: // Array
		// Limit array nesting to avoid deep recursion
		if g.random.IntN(4) == 0 { // 25% chance to stop at primitive
			prims := []schema.PrimitiveFieldType{
				schema.PrimitiveTypeInt64,
				schema.PrimitiveTypeUint64,
				schema.PrimitiveTypeFloat64,
				schema.PrimitiveTypeBool,
				schema.PrimitiveTypeString,
				schema.PrimitiveTypeBytes,
			}
			prim := prims[g.random.IntN(len(prims))]
			return schema.FieldType{Primitive: &schema.PrimitiveType{Type: prim}}
		}
		return schema.FieldType{Array: &schema.ArrayType{ElemType: g.genRandomType(false)}}
	case 4: // Struct
		name := g.structNames[g.random.IntN(len(g.structNames))]
		return schema.FieldType{Struct: name}
	}
	// Fallback to primitive
	return schema.FieldType{Primitive: &schema.PrimitiveType{Type: schema.PrimitiveTypeInt64}}
}

func TestRandomizedSchema(t *testing.T) {
	seed1 := uint64(time.Now().UnixNano())
	random := rand.New(rand.NewPCG(seed1, 0))

	g := schemaGenerator{}
	g.random = random
	sch := g.generate()

	succeeded := false
	defer func() {
		if !succeeded {
			fmt.Printf("Test failed with seed %v\n", seed1)
			schemaContent := sch.PrettyPrint()
			fmt.Printf("Schema:\n%s\n", schemaContent)
		}
	}()

	schemaContent := sch.PrettyPrint()

	// Test the schema. Don't fail if generated code tests fail, just report the seed for now.
	testSchema(t, []byte(schemaContent), "randomized.stef", true)

	succeeded = true
}
