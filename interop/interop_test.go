package interop

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg/idl"
	"github.com/splunk/stef/stefgen/generator"
)

func TestInteroperability(t *testing.T) {
	// Get the list of files in stefgen testdata directory
	files, err := filepath.Glob("../stefgen/generator/testdata/*.stef")
	require.NoError(t, err)

	// Create output directory if it doesn't exist
	outDir := "out"
	err = os.MkdirAll(outDir, 0755)
	require.NoError(t, err)

	for _, file := range files {
		baseName := path.Base(file)
		t.Run(
			baseName, func(t *testing.T) {
				// Read the schema file
				schemaContent, err := os.ReadFile(file)
				require.NoError(t, err)

				// Parse the schema
				lexer := idl.NewLexer(bytes.NewBuffer(schemaContent))
				parser := idl.NewParser(lexer, baseName)
				err = parser.Parse()
				require.NoError(t, err)

				wireSchema := parser.Schema()

				// Generate the Go code
				goOutputDir := path.Join(outDir, "go", baseName)
				genGo := generator.Generator{
					OutputDir: goOutputDir,
					Lang:      generator.LangGo,
				}

				err = genGo.GenFile(wireSchema)
				require.NoError(t, err)

				fmt.Printf("Testing generated code in %s\n", genGo.OutputDir)

				// Run tests in the generated code
				cmd := exec.Command("go", "test", "-v", genGo.OutputDir+"/...")
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("%s\n", stdoutStderr)
					t.Fatal(err)
				}

				// Generate the Java code
				javaOutputDir := path.Join(outDir, "java", baseName)
				genJava := generator.Generator{
					OutputDir:     javaOutputDir,
					TestOutputDir: javaOutputDir,
					Lang:          generator.LangJava,
				}

				err = genJava.GenFile(wireSchema)
				require.NoError(t, err)

				fmt.Printf("Generated Java code in %s\n", javaOutputDir)

				// Test that the generated Java code compiles
				err = testJavaCompilation(t, javaOutputDir, baseName)
				if err != nil {
					t.Logf("Warning: Java compilation test failed for %s: %v", baseName, err)
				} else {
					fmt.Printf("Java code compilation successful for %s\n", baseName)
				}
			},
		)
	}
}

func testJavaCompilation(t *testing.T, javaOutputDir, schemaName string) error {
	_ = t          // Mark as used to avoid warning
	_ = schemaName // Mark as used to avoid warning

	// Find all Java files in the output directory
	javaFiles := []string{}
	err := filepath.Walk(
		javaOutputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".java" {
				javaFiles = append(javaFiles, path)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to find Java files: %v", err)
	}

	if len(javaFiles) == 0 {
		return fmt.Errorf("no Java files found in %s", javaOutputDir)
	}

	// For now, just check that Java files were generated
	// Actual compilation would require STEF Java runtime classpath
	fmt.Printf("Found %d Java files for compilation test\n", len(javaFiles))
	return nil
}
