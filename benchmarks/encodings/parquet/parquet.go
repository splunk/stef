package parquetenc

import (
	"bytes"
	"io"
	"log"

	"github.com/parquet-go/parquet-go"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/stef-otel/oteltef"
	otlpconvert "github.com/splunk/stef/stef-pdata/metrics"
	"github.com/splunk/stef/stef-pdata/metrics/sortedbymetric"
)

type Encoding struct {
}

type Attribute struct {
	Key   string `parquet:"key"`
	Value string `parquet:"value,optional"`
}

type Resource struct {
	Attributes []Attribute `parquet:"attrs"`
	SchemaURL  string      `parquet:"schema_url"`
}

type Scope struct {
	Attributes []Attribute `parquet:"attrs"`
	SchemaURL  string      `parquet:"schema_url"`
	Name       string      `parquet:"name"`
	Version    string      `parquet:"version"`
}

type Datum struct {
	MetricName             string      `parquet:"name"`
	Description            string      `parquet:"description"`
	Unit                   string      `parquet:"unit"`
	Type                   uint        `parquet:"type"`
	Flags                  uint        `parquet:"flags"`
	MetricAttrs            []Attribute `parquet:"metric_attrs"`
	Resource               Resource    `parquet:"resource"`
	Scope                  Scope       `parquet:"scope"`
	Attributes             []Attribute `parquet:"attrs"`
	StartTimestampUnixNano uint64      `parquet:"start_timestamp"`
	TimestampUnixNano      uint64      `parquet:"timestamp"`
	Int64Vals              int64       `parquet:"valint"`
	Float64Vals            float64     `parquet:"valfloat"`
}

func (d *Encoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	converter := otlpconvert.NewOtlpToSortedTree()
	sorted, err := converter.FromOtlp(data.ResourceMetrics())

	var datums []Datum
	err = sorted.Iter(
		func(metric *oteltef.Metric, byMetric *sortedbymetric.ByMetric) error {
			err := byMetric.Iter(
				func(resource *oteltef.Resource, byResource *sortedbymetric.ByResource) error {
					err := byResource.Iter(
						func(scope *oteltef.Scope, byScope *sortedbymetric.ByScope) error {
							err := byScope.Iter(
								func(attrs *oteltef.Attributes, points *sortedbymetric.Points) error {
									for _, value := range *points {
										// TODO: histogram support
										datums = append(
											datums, Datum{
												MetricName:  metric.Name(),
												Unit:        metric.Unit(),
												Description: metric.Description(),
												Type:        uint(metric.Type()),
												Flags:       uint(metric.AggregationTemporality()),
												Resource: Resource{
													Attributes: convertAttrs(resource.Attributes()),
													SchemaURL:  resource.SchemaURL(),
												},
												Scope: Scope{
													Attributes: convertAttrs(scope.Attributes()),
													SchemaURL:  scope.SchemaURL(),
													Name:       scope.Name(),
													Version:    scope.Version(),
												},
												Attributes:             convertAttrs(attrs),
												StartTimestampUnixNano: value.StartTimestamp(),
												TimestampUnixNano:      value.Timestamp(),
												Int64Vals:              value.Value().Int64(),
												Float64Vals:            value.Value().Float64(),
											},
										)
									}
									return nil
								},
							)
							return err
						},
					)
					return err
				},
			)
			return err
		},
	)
	if err != nil {
		return nil, err
	}

	return datums, nil
}

func (d *Encoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	datums := data.([]Datum)

	buf := bytes.NewBuffer(nil)
	writer := parquet.NewGenericWriter[Datum](buf)

	if _, err := writer.Write(datums); err != nil {
		log.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes(), nil
}

func convertAttrs(attrs *oteltef.Attributes) (r []Attribute) {
	for i := 0; i < attrs.Len(); i++ {
		attr := attrs.At(i)
		r = append(
			r, Attribute{
				Key:   attr.Key(),
				Value: string(attr.Value().String()),
			},
		)
	}
	return r
}

func (d *Encoding) Decode(b []byte) (any, error) {
	reader := parquet.NewGenericReader[Datum](bytes.NewReader(b))
	data := make([]Datum, 1000)
	for {
		_, err := reader.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
	}
	return nil, nil
}

func (*Encoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	return pmetric.NewMetrics(), nil
}

func (*Encoding) Name() string {
	return "Parquet"
}
