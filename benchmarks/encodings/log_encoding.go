package encodings

import "go.opentelemetry.io/collector/pdata/plog"

type LogEncoding interface {
	Name() string
	LongName() string
	FromOTLP(batch plog.Logs) (InMemoryData, error)
	Encode(data InMemoryData) ([]byte, error)
	Decode([]byte) (any, error)
	ToOTLP(data []byte) (plog.Logs, error)
}
