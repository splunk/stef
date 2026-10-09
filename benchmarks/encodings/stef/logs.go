package stef

import (
	"bytes"
	"io"

	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/go/otel/otelstef"
	logconvert "github.com/splunk/stef/go/pdata/logs"
	"github.com/splunk/stef/go/pdata/logs/sortedbyresource"
	"github.com/splunk/stef/go/pkg"
)

// LogsSTEFEncoding is resource-sorted STEF format for logs.
type LogsSTEFEncoding struct {
	Opts pkg.WriterOptions
}

var _ encodings.LogEncoding = (*LogsSTEFEncoding)(nil)

func (*LogsSTEFEncoding) FromOTLP(data plog.Logs) (encodings.InMemoryData, error) {
	return sortedbyresource.OtlpToSortedTree(data), nil
}

func (d *LogsSTEFEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	tree := data.(*sortedbyresource.SortedTree)
	output := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewLogsWriter(output, d.Opts)
	if err != nil {
		return nil, err
	}
	if err := tree.ToStef(writer); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (*LogsSTEFEncoding) Decode(data []byte) (any, error) {
	reader, err := otelstef.NewLogsReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for {
		if err := reader.Read(pkg.ReadOptions{}); err != nil {
			if err == io.EOF {
				return nil, nil
			}
			return nil, err
		}
	}
}

func (*LogsSTEFEncoding) ToOTLP(data []byte) (plog.Logs, error) {
	reader, err := otelstef.NewLogsReader(bytes.NewReader(data))
	if err != nil {
		return plog.NewLogs(), err
	}
	converter := logconvert.StefToOtlpUnsorted{}
	return converter.Convert(reader, true)
}

func (e *LogsSTEFEncoding) Name() string {
	name := "STEF"
	if e.Opts.Compression != pkg.CompressionNone {
		name += "Z"
	}
	return name
}

func (*LogsSTEFEncoding) LongName() string {
	return "STEF"
}
