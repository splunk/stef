package hyper

import (
	"buf.build/go/hyperpb"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"google.golang.org/protobuf/proto"

	otlpmetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	"github.com/splunk/stef/benchmarks/encodings"
)

type HyperEncoding struct {
	msgType *hyperpb.MessageType
}

func NewHyperEncoding() *HyperEncoding {
	return &HyperEncoding{
		msgType: hyperpb.CompileMessageDescriptor(
			(*otlpmetrics.ExportMetricsServiceRequest)(nil).ProtoReflect().Descriptor(),
		),
	}
}

func (d *HyperEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return data, nil
}

func (d *HyperEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	marshaler := pmetric.ProtoMarshaler{}
	return marshaler.MarshalMetrics(data.(pmetric.Metrics))
}

func (d *HyperEncoding) Decode(b []byte) (any, error) {
	msg := hyperpb.NewMessage(d.msgType)
	if err := proto.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*HyperEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	marshaler := pmetric.ProtoUnmarshaler{}
	return marshaler.UnmarshalMetrics(data)
}

func (*HyperEncoding) Name() string {
	return "Hyper"
}
func (*HyperEncoding) LongName() string {
	return "Protobuf Hyperpb"
}
