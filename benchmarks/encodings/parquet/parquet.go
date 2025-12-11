package parquetenc

import (
	"bytes"
	"io"
	"log"

	"github.com/parquet-go/parquet-go"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/benchmarks/encodings"
	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/metrics/sortedbymetric"
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
	ValueType              uint        `parquet:"value_type"`
	Int64Vals              int64       `parquet:"valint"`
	Float64Vals            float64     `parquet:"valfloat"`
}

func (d *Encoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	sorted, err := sortedbymetric.OtlpToSortedTree(data)

	var datums []Datum
	err = sorted.Iter(
		func(metric *otelstef.Metric, byMetric *sortedbymetric.ByMetric) error {
			err := byMetric.Iter(
				func(resource *otelstef.Resource, byResource *sortedbymetric.ByResource) error {
					err := byResource.Iter(
						func(scope *otelstef.Scope, byScope *sortedbymetric.ByScope) error {
							err := byScope.Iter(
								func(attrs *otelstef.Attributes, points *sortedbymetric.Points) error {
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
												ValueType:              uint(value.Value().Type()),
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

func convertAttrs(attrs *otelstef.Attributes) (r []Attribute) {
	for i := 0; i < attrs.Len(); i++ {
		r = append(
			r, Attribute{
				Key:   attrs.Key(i),
				Value: string(attrs.Value(i).String()),
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

func (d *Encoding) ToOTLP(data []byte) (dst pmetric.Metrics, err error) {
	reader := parquet.NewGenericReader[Datum](bytes.NewReader(data))
	records := make([]Datum, 1000)
	dst = pmetric.NewMetrics()
	for {
		n, err := reader.Read(records)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		for i := 0; i < n; i++ {
			record := records[i]

			rm := dst.ResourceMetrics().AppendEmpty()
			covertResource(rm.Resource(), record.Resource)

			sm := rm.ScopeMetrics().AppendEmpty()
			ms := sm.Metrics().AppendEmpty()

			ms.SetName(record.MetricName)
			ms.SetDescription(record.Description)
			ms.SetUnit(record.Unit)

			switch pmetric.MetricType(record.Type) {
			case pmetric.MetricTypeEmpty:

			case pmetric.MetricTypeGauge:
				gauge := ms.SetEmptyGauge()
				point := gauge.DataPoints().AppendEmpty()
				convertAttrsFrom(point.Attributes(), record.Attributes)
				point.SetTimestamp(pcommon.Timestamp(record.TimestampUnixNano))
				point.SetStartTimestamp(pcommon.Timestamp(record.StartTimestampUnixNano))
				switch pmetric.NumberDataPointValueType(record.ValueType) {
				case pmetric.NumberDataPointValueTypeInt:
					point.SetIntValue(record.Int64Vals)
				case pmetric.NumberDataPointValueTypeDouble:
					point.SetDoubleValue(record.Float64Vals)
				}

			default:
				log.Fatalf("Unsupported metric type: %v", record.Type)
			}
		}
	}
	return dst, nil
}

func covertResource(dest pcommon.Resource, src Resource) {
	convertAttrsFrom(dest.Attributes(), src.Attributes)
}

func convertAttrsFrom(dest pcommon.Map, src []Attribute) {
	dest.EnsureCapacity(len(src))
	for _, attr := range src {
		dest.PutStr(attr.Key, attr.Value)
	}
}

func (d *Encoding) Name() string {
	return "Parquet"
}
func (d *Encoding) LongName() string {
	return d.Name()
}
