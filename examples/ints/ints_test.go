package main

import (
	"bytes"
	"log"
	"testing"

	"github.com/splunk/stef/examples/ints/internal/ints"
	"github.com/splunk/stef/go/pkg"
)

func writeInts(recCount int) *pkg.MemChunkWriter {
	buf := &pkg.MemChunkWriter{}
	writer, err := ints.NewRecordWriter(buf, pkg.WriterOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < recCount; i++ {
		writer.Record.SetUint64(uint64(i % 8))
		err = writer.Write()
		if err != nil {
			log.Fatal(err)
		}
	}
	err = writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
	return buf
}

func BenchmarkSerialize(b *testing.B) {
	const recCount = 10000
	for i := 0; i < b.N; i++ {
		writeInts(recCount)
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/rec")
}

func BenchmarkDeserialize(b *testing.B) {
	const recCount = 10000
	writeBuf := writeInts(recCount)
	testData := writeBuf.Bytes()
	var readBuf bytes.Reader
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readBuf.Reset(testData)
		reader, err := ints.NewRecordReader(&readBuf)
		if err != nil {
			log.Fatal(err)
		}
		for j := 0; j < recCount; j++ {
			err := reader.Read(pkg.ReadOptions{})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/rec")
}
