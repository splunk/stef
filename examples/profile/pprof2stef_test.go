package profile

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pprof "github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"

	stefprofile "github.com/splunk/stef/examples/profile/internal/profile"
	"github.com/splunk/stef/go/pkg"
)

func TestConvertToStef(t *testing.T) {
	// Get the testdata directory path
	testdataDir := "testdata"

	// Read all files in the testdata directory
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	// Filter for .prof files
	var profFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".prof") {
			profFiles = append(profFiles, entry.Name())
		}
	}

	fmt.Println("Uncompressed bytes sizes")
	fmt.Printf("%-20s | %-6s | %-6s\n", "File", "pprof", "stef")

	// Test each .prof file
	for _, fileName := range profFiles {
		filePath := filepath.Join(testdataDir, fileName)

		// Open the profile file
		file, err := os.Open(filePath)
		require.NoError(t, err)
		defer func() { file.Close() }()

		gz, err := gzip.NewReader(file)
		pprofData, err := io.ReadAll(gz)
		require.NoError(t, err)

		// Parse the profile using profile.Parse
		prof, err := pprof.ParseUncompressed(pprofData)
		require.NoError(t, err)
		require.NotNil(t, prof)

		buf := bytes.NewBuffer(nil)
		err = convertPprofToStef(prof, buf)
		require.NoError(t, err)

		improv := float64(buf.Len()) / float64(len(pprofData))
		fmt.Printf("%-20s | %6d | %6d (%5.3fx)\n", fileName, len(pprofData), buf.Len(), improv)
	}
}

func BenchmarkDeserialization(b *testing.B) {
	dir := "testdata"
	files, err := ioutil.ReadDir(dir)
	require.NoError(b, err, "failed to read testdata dir")
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".prof" {
			continue
		}
		path := filepath.Join(dir, file.Name())
		freader, err := os.Open(path)
		require.NoError(b, err)
		defer func() { freader.Close() }()

		gz, err := gzip.NewReader(freader)
		pprofData, err := io.ReadAll(gz)
		require.NoError(b, err)

		// Parse the profile using profile.Parse
		prof, err := pprof.ParseUncompressed(pprofData)
		require.NoError(b, err)
		require.NotNil(b, prof)

		buf := bytes.NewBuffer(nil)
		err = convertPprofToStef(prof, buf)
		require.NoError(b, err)
		stefData := buf.Bytes()

		b.Run(
			"file="+file.Name()+"/format=pprof", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					pprof.ParseUncompressed(pprofData)
				}
			},
		)

		b.Run(
			"file="+file.Name()+"/format=stef", func(b *testing.B) {
				// Count records
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r := bytes.NewReader(stefData)
					reader, err := stefprofile.NewSampleReader(r)
					if err != nil {
						b.Fatalf("failed to create STEF reader: %v", err)
					}
					for {
						if err := reader.Read(pkg.ReadOptions{}); err != nil {
							break
						}
					}
				}
			},
		)
	}
}
