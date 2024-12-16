package otelarrow

import (
	v1 "github.com/open-telemetry/otel-arrow/api/experimental/arrow/v1"
	"github.com/open-telemetry/otel-arrow/pkg/config"
	"github.com/open-telemetry/otel-arrow/pkg/otel/arrow_record"
	"github.com/open-telemetry/otel-arrow/pkg/otel/metrics/otlp"
	"github.com/open-telemetry/otel-arrow/pkg/werror"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"google.golang.org/protobuf/proto"

	"github.com/tigrannajaryan/stef/benchmarks/encodings"
)

type OtelArrowEncoding struct {
}

func (d *OtelArrowEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return data, nil
}

func (d *OtelArrowEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	arrowProducer := arrow_record.NewProducerWithOptions(config.WithNoZstd())
	records, err := arrowProducer.BatchArrowRecordsFromMetrics(data.(pmetric.Metrics))
	if err != nil {
		return nil, err
	}

	b, err := proto.Marshal(records)
	return b, err
}

func (d *OtelArrowEncoding) Decode(b []byte) (any, error) {
	var bar v1.BatchArrowRecords
	err := proto.Unmarshal(b, &bar)

	arrowConsumer := arrow_record.NewConsumer()

	records, err := arrowConsumer.Consume(&bar)
	if err != nil {
		return nil, err
	}

	result := make([]pmetric.Metrics, 0, len(records))

	// builds the related entities (i.e. Attributes, Summaries, Histograms, ...)
	// from the records and returns the main record.
	// This only does half the job that is necessary to access and work with
	// the metric data.
	_, metricsRecord, err := otlp.RelatedDataFrom(records)
	if err != nil {
		return nil, werror.Wrap(err)
	}

	// For fairness we also need to iterate over the actual datapoints
	// but there is no simple way to do it here. It's a TODO for later.
	// Process the main record with the related entities.
	if metricsRecord != nil {
		// Decode OTLP metrics from the combination of the main record and the
		// related records.
	}

	return result, err
}

func (*OtelArrowEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	var bar v1.BatchArrowRecords
	err := proto.Unmarshal(data, &bar)

	arrowConsumer := arrow_record.NewConsumer()

	records, err := arrowConsumer.Consume(&bar)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	result := pmetric.NewMetrics()

	// builds the related entities (i.e. Attributes, Summaries, Histograms, ...)
	// from the records and returns the main record.
	relatedData, metricsRecord, err := otlp.RelatedDataFrom(records)
	if err != nil {
		return pmetric.NewMetrics(), werror.Wrap(err)
	}

	if metricsRecord != nil {
		// Decode OTLP metrics from the combination of the main record and the
		// related records.
		metrics, err := otlp.MetricsFrom(metricsRecord.Record(), relatedData)
		if err != nil {
			return pmetric.NewMetrics(), werror.Wrap(err)
		}
		metrics.ResourceMetrics().MoveAndAppendTo(result.ResourceMetrics())
	}

	return result, err
}

func (*OtelArrowEncoding) Name() string {
	return "ARROW"
}

func (e *OtelArrowEncoding) StartMultipart(compression string) (encodings.MetricMultipartStream, error) {
	opts := []config.Option{}
	if compression == "zstd" {
		opts = append(opts, config.WithZstd())
	} else {
		opts = append(opts, config.WithNoZstd())
	}
	arrowProducer := arrow_record.NewProducerWithOptions(opts...)
	return &multipart{producer: arrowProducer}, nil
}

type multipart struct {
	producer *arrow_record.Producer
	bytes    []byte
}

func (m *multipart) AppendPart(part pmetric.Metrics) error {
	bar, err := m.producer.BatchArrowRecordsFromMetrics(part)
	if err != nil {
		panic(err)
	}
	bytes, err := proto.Marshal(bar)
	if err != nil {
		return err
	}
	m.bytes = append(m.bytes, bytes...)
	return nil
}

func (m *multipart) FinishStream() ([]byte, error) {
	return m.bytes, nil
}
