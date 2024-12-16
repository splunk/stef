package parquetenc

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/parquet-go/parquet-go"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/tigrannajaryan/stef/benchmarks/encodings"
	"github.com/tigrannajaryan/stef/stef-otel/oteltef"
	otlpconvert "github.com/tigrannajaryan/stef/stef-pdata/metrics"
	"github.com/tigrannajaryan/stef/stef-pdata/metrics/sortedbymetric"
)

type EncodingZ struct {
}

type AttributeZ struct {
	Key   string `parquet:"key,zstd"`
	Value string `parquet:"value,optional,zstd"`
}

type ResourceZ struct {
	Attributes []AttributeZ `parquet:"attrs"`
	SchemaURL  string       `parquet:"schema_url,zstd"`
}

type ScopeZ struct {
	Attributes []AttributeZ `parquet:"attrs"`
	SchemaURL  string       `parquet:"schema_url,zstd"`
	Name       string       `parquet:"name,zstd"`
	Version    string       `parquet:"version,zstd"`
}

type DatumZ struct {
	MetricName             string       `parquet:"name,zstd"`
	Description            string       `parquet:"description,zstd"`
	Unit                   string       `parquet:"unit,zstd"`
	Type                   uint         `parquet:"type,zstd"`
	Flags                  uint         `parquet:"flags,zstd"`
	MetricAttrs            []AttributeZ `parquet:"metric_attrs"`
	Resource               ResourceZ    `parquet:"resource"`
	Scope                  ScopeZ       `parquet:"scope"`
	Attributes             []AttributeZ `parquet:"attrs"`
	StartTimestampUnixNano uint64       `parquet:"start_timestamp,zstd"`
	TimestampUnixNano      uint64       `parquet:"timestamp,zstd"`
	Int64Val               int64        `parquet:"valint,zstd"`
	Float64Val             float64      `parquet:"valfloat,zstd"`
}

type DatumZSlice []*DatumZ

func (r DatumZSlice) Len() int {
	return len(r)
}

func lessRecord(r1, r2 *DatumZ) bool {
	// Sort by OrgID first.
	//if r1.OrgId < r2.OrgId {
	//	return true
	//} else if r1.OrgId > r2.OrgId {
	//	return false
	//}

	// Then by metric name
	c := strings.Compare(r1.MetricName, r2.MetricName)
	if c < 0 {
		return true
	} else if c > 0 {
		return false
	}

	c = int(r1.Type) - int(r2.Type)
	if c < 0 {
		return true
	} else if c > 0 {
		return false
	}
	c = int(r1.Flags) - int(r2.Flags)
	if c < 0 {
		return true
	} else if c > 0 {
		return false
	}

	c = cmpAttrList(r1.MetricAttrs, r2.MetricAttrs)
	if c < 0 {
		return true
	} else if c > 0 {
		return false
	}

	c = cmpAttrList(r1.Attributes, r2.Attributes)
	if c < 0 {
		return true
	} else if c > 0 {
		return false
	}

	// Then by timestamp.
	return r1.TimestampUnixNano < r2.TimestampUnixNano
}

func cmpAttrList(left []AttributeZ, right []AttributeZ) int {
	l := min(len(left), len(right))
	i := 0
	for ; i < l; i++ {
		c := strings.Compare(left[i].Key, right[i].Key)
		if c != 0 {
			return c
		}
		c = strings.Compare(left[i].Value, right[i].Value)
		if c != 0 {
			return c
		}
	}
	if l < len(right) {
		return -1
	} else if l < len(left) {
		return 1
	}
	return 0
}

func (r DatumZSlice) Less(i, j int) bool {
	return lessRecord(r[i], r[j])
}

func (r DatumZSlice) Swap(i, j int) {
	rec := r[i]
	r[i] = r[j]
	r[j] = rec
}

//func (d *EncodingZ) Encode2(data *metricspb.MetricsData) ([]byte, error) {
//	converter := metrics.NewConverter()
//	records := converter.Otlp2records(data.ResourceMetrics)
//	sort.Sort(records)
//
//	buf := bytes.NewBuffer(nil)
//	writer := parquet.NewGenericWriter[DatumZ](buf)
//
//	var datums []DatumZ
//	for _, r := range records {
//		datums = append(
//			datums, DatumZ{
//				MetricName:        r.MetricName,
//				Unit:              r.Unit,
//				Description:       r.Description,
//				Type:              uint(r.Type),
//				Flags:             uint(r.Flags),
//				Attributes:        convertAttrsZ(r.Metadata),
//				TimestampUnixNano: r.TimestampUnixNano,
//				Int64Vals:         r.Int64Vals,
//				Float64Vals:       r.Float64Vals,
//			},
//		)
//	}
//
//	if _, err := writer.Write(datums); err != nil {
//		log.Fatal(err)
//	}
//
//	if err := writer.Close(); err != nil {
//		log.Fatal(err)
//	}
//	return buf.Bytes(), nil
//}

func (d *EncodingZ) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	converter := otlpconvert.NewOtlpToSortedTree()
	sorted, err := converter.FromOtlp(data.ResourceMetrics())
	if err != nil {
		return nil, err
	}

	var datums []DatumZ
	err = sorted.Iter(
		func(metric *oteltef.Metric, byMetric *sortedbymetric.ByMetric) error {
			err := byMetric.Iter(
				func(resource *oteltef.Resource, byResource *sortedbymetric.ByResource) error {
					err := byResource.Iter(
						func(scope *oteltef.Scope, byScope *sortedbymetric.ByScope) error {
							err := byScope.Iter(
								func(attrs *oteltef.Attributes, points *sortedbymetric.Points) error {
									for _, value := range *points {
										datums = append(
											datums, DatumZ{
												MetricName:  metric.Name(),
												Unit:        metric.Unit(),
												Description: metric.Description(),
												Type:        uint(metric.Type()),
												Flags:       uint(metric.AggregationTemporality()),
												Resource: ResourceZ{
													Attributes: convertAttrsZ(resource.Attributes()),
													SchemaURL:  resource.SchemaURL(),
												},
												Scope: ScopeZ{
													Attributes: convertAttrsZ(scope.Attributes()),
													SchemaURL:  scope.SchemaURL(),
													Name:       scope.Name(),
													Version:    scope.Version(),
												},
												Attributes:             convertAttrsZ(attrs),
												StartTimestampUnixNano: value.StartTimestamp(),
												TimestampUnixNano:      value.Timestamp(),
												Int64Val:               value.Value().Int64(),
												Float64Val:             value.Value().Float64(),
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

func (d *EncodingZ) Encode(data encodings.InMemoryData) ([]byte, error) {
	datums := data.([]DatumZ)

	buf := bytes.NewBuffer(nil)
	writer := parquet.NewGenericWriter[DatumZ](buf)

	if _, err := writer.Write(datums); err != nil {
		log.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes(), nil
}

func convertAttrsZ(attrs *oteltef.Attributes) (r []AttributeZ) {
	for i := 0; i < attrs.Len(); i++ {
		attr := attrs.At(i)
		r = append(
			r, AttributeZ{
				Key:   attr.Key(),
				Value: string(attr.Value().String()),
			},
		)
	}
	return r
}

func (d *EncodingZ) Decode(b []byte) (any, error) {
	reader := parquet.NewGenericReader[DatumZ](bytes.NewReader(b))
	data := make([]DatumZ, 1000)
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

func (*EncodingZ) ToOTLP(data []byte) (pmetric.Metrics, error) {
	return pmetric.NewMetrics(), nil
}

func (*EncodingZ) Name() string {
	return "ParquetZ"
}
