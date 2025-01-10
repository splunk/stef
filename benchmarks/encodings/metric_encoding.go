package encodings

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricEncoding interface {
	Name() string
	FromOTLP(batch pmetric.Metrics) (InMemoryData, error)
	Encode(data InMemoryData) ([]byte, error)
	Decode([]byte) (any, error)
	ToOTLP(data []byte) (pmetric.Metrics, error)
}

type InMemoryData any

type MetricMultipartEncoding interface {
	Name() string
	StartMultipart(compression string) (MetricMultipartStream, error)
}

type MetricMultipartStream interface {
	AppendPart(part pmetric.Metrics) error
	FinishStream() ([]byte, error)
}
