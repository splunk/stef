package otelstef

import (
	"bytes"
	"testing"

	"github.com/splunk/stef/go/pkg"
)

func TestLogsCloneTracksExistingAttributeValueModifications(t *testing.T) {
	var source Logs
	source.Init()
	source.Log().Attributes().EnsureLen(1)
	source.Log().Attributes().SetKey(0, "key")
	source.Log().Attributes().Value(0).SetInt64(1)

	var allocators Allocators
	var copy Logs
	source.CloneTo(&copy, &allocators)

	if copy.IsLogModified() || copy.Log().IsAttributesModified() {
		t.Fatal("clone is unexpectedly marked modified")
	}

	copy.Log().Attributes().Value(0).SetInt64(2)

	if !copy.IsLogModified() || !copy.Log().IsAttributesModified() {
		t.Fatal("attribute value modification did not propagate through clone")
	}
}

func TestLogsCloneModificationSurvivesEncoding(t *testing.T) {
	var buf pkg.MemChunkWriter
	writer, err := NewLogsWriter(&buf, pkg.WriterOptions{})
	if err != nil {
		t.Fatal(err)
	}

	writer.Record.Log().Attributes().EnsureLen(1)
	writer.Record.Log().Attributes().SetKey(0, "key")
	writer.Record.Log().Attributes().Value(0).SetInt64(1)
	if err := writer.Write(); err != nil {
		t.Fatal(err)
	}

	var allocators Allocators
	var cloned Logs
	writer.Record.CloneTo(&cloned, &allocators)
	cloned.Log().Attributes().Value(0).SetInt64(2)
	writer.Record.CopyFrom(&cloned)
	if err := writer.Write(); err != nil {
		t.Fatal(err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	reader, err := NewLogsReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Read(pkg.ReadOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := reader.Record.Log().Attributes().Value(0).Int64(); got != 1 {
		t.Fatalf("first attribute value = %d, want 1", got)
	}
	if err := reader.Read(pkg.ReadOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := reader.Record.Log().Attributes().Value(0).Int64(); got != 2 {
		t.Fatalf("cloned attribute value after encoding = %d, want 2", got)
	}
}
