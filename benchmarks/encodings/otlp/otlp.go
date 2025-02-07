package otlp

import (
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/benchmarks/testutils"
)

type OTLPEncoding struct {
}

func (d *OTLPEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return data, nil
}

func (d *OTLPEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	marshaler := pmetric.ProtoMarshaler{}
	return marshaler.MarshalMetrics(data.(pmetric.Metrics))
}

func (d *OTLPEncoding) Decode(b []byte) (any, error) {
	return d.ToOTLP(b)
}

func (*OTLPEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	marshaler := pmetric.ProtoUnmarshaler{}
	return marshaler.UnmarshalMetrics(data)
}

func (*OTLPEncoding) Name() string {
	return "OTLP"
}
func (*OTLPEncoding) LongName() string {
	return "Protobuf OTLP"
}

type otlpMultipart struct {
	compression string
	bytes       []byte
}

func (o *otlpMultipart) AppendPart(part pmetric.Metrics) error {
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

func (o *otlpMultipart) FinishStream() ([]byte, error) {
	return o.bytes, nil
}

func (d *OTLPEncoding) StartMultipart(compression string) (encodings.MetricMultipartStream, error) {
	return &otlpMultipart{compression: compression}, nil
}
