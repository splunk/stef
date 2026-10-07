package otlp

import (
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/benchmarks/testutils"
)

type MetricsEncoding struct {
}

func (d *MetricsEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return data, nil
}

func (d *MetricsEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	marshaler := pmetric.ProtoMarshaler{}
	return marshaler.MarshalMetrics(data.(pmetric.Metrics))
}

func (d *MetricsEncoding) Decode(b []byte) (any, error) {
	return d.ToOTLP(b)
}

func (*MetricsEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	marshaler := pmetric.ProtoUnmarshaler{}
	return marshaler.UnmarshalMetrics(data)
}

func (*MetricsEncoding) Name() string {
	return "OTLP"
}
func (*MetricsEncoding) LongName() string {
	return "Protobuf OTLP"
}

type metricsMultipart struct {
	compression string
	bytes       []byte
}

func (o *metricsMultipart) AppendPart(part pmetric.Metrics) error {
	marshaler := pmetric.ProtoMarshaler{}
	b, err := marshaler.MarshalMetrics(part)
	if err != nil {
		return err
	}

	if o.compression == "zstd" {
		b = testutils.CompressZstd(b)
	}

	o.bytes = append(o.bytes, b...)
	return nil
}

func (o *metricsMultipart) FinishStream() ([]byte, error) {
	return o.bytes, nil
}

func (d *MetricsEncoding) StartMultipart(compression string) (encodings.MetricMultipartStream, error) {
	return &metricsMultipart{compression: compression}, nil
}
