package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/examples/jsonl/internal/jsonpb"
	"github.com/splunk/stef/examples/jsonl/internal/jsonstef"
	"github.com/splunk/stef/go/pkg"
)

func TestConvertToJsonValue(t *testing.T) {
	dir := "./testdata"
	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err, "failed to read testdata dir")
	fmt.Printf(
		"%-30s | %-6s | %-14s | %-14s | %-6s | %-6s | %-6s\n",
		"File --> Size in bytes", "JSON", "Protobuf", "STEF", "JSONZ", "ProtoZ", "stefz",
	)

	enc, err := zstd.NewWriter(nil)
	require.NoError(t, err, "failed to create zstd writer")
	defer enc.Close()
	dec, err := zstd.NewReader(nil)
	require.NoError(t, err, "failed to create zstd reader")
	defer dec.Close()
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".zst" {
			continue
		}
		path := filepath.Join(dir, file.Name())
		compressedData, err := ioutil.ReadFile(path)
		require.NoErrorf(t, err, "failed to read file %s", path)
		jsonData, err := dec.DecodeAll(compressedData, nil)
		require.NoErrorf(t, err, "failed to decompress file %s", path)
		stefData, err := convertJSONLToSTEF(jsonData, pkg.WriterOptions{})
		require.NoErrorf(t, err, "failed to create writer for %s", path)
		protoData, err := convertJSONLToProto(jsonData)
		require.NoErrorf(t, err, "failed to convert to proto for %s", path)

		// Compress with zstd
		jsonCompressed := enc.EncodeAll(jsonData, nil)
		protoCompressed := enc.EncodeAll(protoData, nil)
		stefCompressed := enc.EncodeAll(stefData, nil)

		var improvementProto, improvementStef float64
		if len(protoData) > 0 {
			improvementProto = float64(len(protoData)) / float64(len(jsonData))
		}
		if len(stefData) > 0 {
			improvementStef = float64(len(stefData)) / float64(len(jsonData))
		}
		fmt.Printf(
			"%-30s | %6d | %6d (%4.2fx) | %6d (%4.2fx) | %6d | %6d | %6d\n",
			file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))], len(jsonData), len(protoData),
			improvementProto, len(stefData), improvementStef, len(jsonCompressed), len(protoCompressed),
			len(stefCompressed),
		)
	}
}

func BenchmarkDeserialization(b *testing.B) {
	dir := "./testdata"
	files, err := ioutil.ReadDir(dir)
	require.NoError(b, err, "failed to read testdata dir")
	dec, err := zstd.NewReader(nil)
	require.NoError(b, err, "failed to create zstd reader")
	defer dec.Close()
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".zst" {
			continue
		}
		path := filepath.Join(dir, file.Name())
		compressedData, err := ioutil.ReadFile(path)
		require.NoErrorf(b, err, "failed to read file %s", path)
		jsonData, err := dec.DecodeAll(compressedData, nil)
		require.NoErrorf(b, err, "failed to decompress file %s", path)
		stefData, err := convertJSONLToSTEF(jsonData, pkg.WriterOptions{})
		require.NoErrorf(b, err, "failed to create writer for %s", path)
		protoData, err := convertJSONLToProto(jsonData)
		require.NoErrorf(b, err, "failed to convert to proto for %s", path)

		fileBase := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]

		recordCount := 0

		b.Run(
			"file="+fileBase+"/format=json", func(b *testing.B) {
				// Count records
				for i := 0; i < b.N; i++ {
					recordCount = 0
					r := bytes.NewReader(jsonData)
					scanner := bufio.NewScanner(r)
					for scanner.Scan() {
						recordCount++
						var v interface{}
						_ = json.Unmarshal(scanner.Bytes(), &v)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recordCount), "ns/record")
			},
		)

		b.Run(
			"file="+fileBase+"/format=stef", func(b *testing.B) {
				// Count records
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					recordCount = 0
					r := bytes.NewReader(stefData)
					reader, err := jsonstef.NewRecordReader(r)
					if err != nil {
						b.Fatalf("failed to create STEF reader: %v", err)
					}
					for {
						if err := reader.Read(pkg.ReadOptions{}); err != nil {
							break
						}
						recordCount++
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recordCount), "ns/record")
			},
		)

		b.Run(
			"file="+fileBase+"/format=protobuf", func(b *testing.B) {
				// Count records
				rec := &jsonpb.Record{}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					recordCount = 0
					rec.Reset()
					var offset int
					for offset < len(protoData) {
						msgLen, n := proto.DecodeVarint(protoData[offset:])
						if n <= 0 || offset+n+int(msgLen) > len(protoData) {
							break
						}
						offset += n
						err := proto.Unmarshal(protoData[offset:offset+int(msgLen)], rec)
						if err != nil {
							b.Fatalf("failed to unmarshal proto: %v", err)
						}
						offset += int(msgLen)
						recordCount++
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recordCount), "ns/record")
			},
		)
	}
}
