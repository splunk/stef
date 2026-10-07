package otlp

import (
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/benchmarks/encodings"
)

type LogsEncoding struct{}

var _ encodings.LogEncoding = (*LogsEncoding)(nil)

func (*LogsEncoding) FromOTLP(data plog.Logs) (encodings.InMemoryData, error) {
	return data, nil
}

func (*LogsEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	marshaler := plog.ProtoMarshaler{}
	return marshaler.MarshalLogs(data.(plog.Logs))
}

func (d *LogsEncoding) Decode(data []byte) (any, error) {
	return d.ToOTLP(data)
}

func (*LogsEncoding) ToOTLP(data []byte) (plog.Logs, error) {
	unmarshaler := plog.ProtoUnmarshaler{}
	return unmarshaler.UnmarshalLogs(data)
}

func (*LogsEncoding) Name() string {
	return "OTLP"
}

func (*LogsEncoding) LongName() string {
	return "Protobuf OTLP"
}
