package generator

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
)

func TestGenerate(t *testing.T) {
	// Get the list of files in "testdata" directory
	files, err := filepath.Glob("testdata/*.stef")
	require.NoError(t, err)

	for _, file := range files {
		t.Run(
			file, func(t *testing.T) {
				// Create a new generator instance
				g := Generator{
					OutputDir: path.Join("testdata", "out", path.Base(file)),
					Lang:      LangGo,
				}

				// Read the schema file
				schemaContent, err := os.ReadFile(file)
				require.NoError(t, err)

				// Parse the schema
				lexer := idl.NewLexer(bytes.NewBuffer(schemaContent))
				parser := idl.NewParser(lexer, path.Base(file))
				err = parser.Parse()
				require.NoError(t, err)

				wireSchema := parser.Schema()

				// Generate the code
				err = g.GenFile(wireSchema)
				require.NoError(t, err)

				fmt.Printf("Testing generated code in %s\n", g.OutputDir)

				// Run tests in the generated code
				cmd := exec.Command("go", "test", "-v", g.OutputDir+"/...")
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("%s\n", stdoutStderr)
					t.Fatal(err)
				}
			},
		)
	}
}
