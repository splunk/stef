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
	"github.com/klauspost/compress/zstd"
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

	fmt.Println("Uncompressed and zstd-compressed bytes sizes")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s | %-6s\n", "File", "pprof", "stef", "pprofz", "stefz")

	// Create zstd encoder
	encoder, err := zstd.NewWriter(nil)
	require.NoError(t, err)
	defer encoder.Close()

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
		stefData := buf.Bytes()

		// Compress both formats with zstd
		pprofCompressed := encoder.EncodeAll(pprofData, nil)
		stefCompressed := encoder.EncodeAll(stefData, nil)

		improv := float64(len(stefData)) / float64(len(pprofData))
		compImprov := float64(len(stefCompressed)) / float64(len(pprofCompressed))
		fmt.Printf(
			"%-20s | %6d | %6d | %6d | %6d (%5.3fx/%5.3fx)\n",
			fileName, len(pprofData), len(stefData), len(pprofCompressed), len(stefCompressed), improv, compImprov,
		)
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
					_, err := pprof.ParseUncompressed(pprofData)
					if err != nil {
						b.Fatalf("failed to parse pprof data: %v", err)
					}
				}
				recCount := len(prof.Sample)
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/sample")
			},
		)

		b.Run(
			"file="+file.Name()+"/format=stef", func(b *testing.B) {
				// Count records
				b.ResetTimer()
				recCount := 0
				for i := 0; i < b.N; i++ {
					r := bytes.NewReader(stefData)
					reader, err := stefprofile.NewSampleReader(r)
					if err != nil {
						b.Fatalf("failed to create STEF reader: %v", err)
					}
					recCount = 0
					for {
						if err := reader.Read(pkg.ReadOptions{}); err != nil {
							break
						}
						recCount++
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/sample")
			},
		)
	}
}

func BenchmarkSerialization(b *testing.B) {
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
				var w bytes.Buffer
				for i := 0; i < b.N; i++ {
					w.Reset()
					err = prof.WriteUncompressed(&w)
					if err != nil {
						b.Fatalf("failed to write pprof data: %v", err)
					}
				}
				recCount := len(prof.Sample)
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/sample")
			},
		)

		b.Run(
			"file="+file.Name()+"/format=stef", func(b *testing.B) {
				// Count records
				r := bytes.NewReader(stefData)
				reader, err := stefprofile.NewSampleReader(r)
				if err != nil {
					b.Fatalf("failed to create STEF reader: %v", err)
				}
				recCount := 0
				records := make([]stefprofile.Sample, len(prof.Sample))
				for {
					if err := reader.Read(pkg.ReadOptions{}); err != nil {
						break
					}
					records[recCount].CopyFrom(&reader.Record)
					recCount++
				}
				b.ResetTimer()
				var dst bytes.Buffer
				for i := 0; i < b.N; i++ {
					dst.Reset()
					err = convertPprofToStef(prof, &dst)
					//chunkWriter := pkg.NewWrapChunkWriter(&dst)
					//
					//// Create sample writer
					//writer, err := stefprofile.NewSampleWriter(chunkWriter, pkg.WriterOptions{})
					//if err != nil {
					//	b.Fatalf("failed to create srcSample writer: %s", err)
					//}
					//
					//for j := 0; j < recCount; j++ {
					//	writer.Record.CopyFrom(&records[j])
					//	if err := writer.Write(); err != nil {
					//		b.Fatalf("failed to write srcSample: %v", err)
					//	}
					//}
					//err = writer.Flush()
					if err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/sample")
			},
		)
	}
}
