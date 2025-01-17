package internal

import (
	"fmt"
	"log"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
)

type BaseOTLPToTEF struct {
	TempAttrs oteltef.Attributes
	Otlp2tef  otlptools.Otlp2Tef
}

func (c *BaseOTLPToTEF) ConvertNumDatapoint(dst *oteltef.Point, src pmetric.NumberDataPoint) {
	dst.SetTimestamp(uint64(src.Timestamp()))
	dst.SetStartTimestamp(uint64(src.StartTimestamp()))

	switch src.ValueType() {
	case pmetric.NumberDataPointValueTypeInt:
		dst.Value().SetInt64(src.IntValue())
	case pmetric.NumberDataPointValueTypeDouble:
		dst.Value().SetFloat64(src.DoubleValue())
	default:
		log.Fatalf("Unsupported value type: %v", src)
	}
}

func (c *BaseOTLPToTEF) ConvertExemplars(dst *oteltef.ExemplarArray, src pmetric.ExemplarSlice) {
	dst.EnsureLen(src.Len())

	for i := 0; i < src.Len(); i++ {
		srcExemplar := src.At(i)
		c.Otlp2tef.MapSorted(srcExemplar.FilteredAttributes(), &c.TempAttrs)
		traceId := srcExemplar.TraceID()
		spanId := srcExemplar.SpanID()

		dstExemplar := dst.At(i)
		dstExemplar.SetTimestamp(uint64(srcExemplar.Timestamp()))

		dstExemplar.FilteredAttributes().CopyFrom(&c.TempAttrs)

		dstExemplar.SetTraceID(pkg.Bytes(traceId[:]))
		dstExemplar.SetSpanID(pkg.Bytes(spanId[:]))

		switch srcExemplar.ValueType() {
		case pmetric.ExemplarValueTypeInt:
			dstExemplar.Value().SetInt64(srcExemplar.IntValue())
		case pmetric.ExemplarValueTypeDouble:
			dstExemplar.Value().SetFloat64(srcExemplar.DoubleValue())
		case pmetric.ExemplarValueTypeEmpty:
			dstExemplar.Value().SetType(oteltef.ExemplarValueTypeNone)
		default:
			panic("unknown Exemplar value type")
		}
	}
}

func (c *BaseOTLPToTEF) ConvertHistogram(dst *oteltef.Point, src pmetric.HistogramDataPoint) error {
	dst.SetTimestamp(uint64(src.Timestamp()))
	dst.SetStartTimestamp(uint64(src.StartTimestamp()))

	dstVal := dst.Value()
	dstVal.SetType(oteltef.PointValueTypeHistogram)
	dstHistogram := dstVal.Histogram()
	dstHistogram.SetCount(int64(src.Count()))

	if src.HasSum() {
		dstHistogram.SetSum(src.Sum())
	} else {
		dstHistogram.UnsetSum()
	}
	if src.HasMin() {
		dstHistogram.SetMin(src.Min())
	} else {
		dstHistogram.UnsetMin()
	}
	if src.HasMax() {
		dstHistogram.SetMax(src.Max())
	} else {
		dstHistogram.UnsetMax()
	}

	if src.BucketCounts().Len() != src.ExplicitBounds().Len()+1 {
		return fmt.Errorf(
			"invalid histogram, bucket counts len %d, bounds len %d",
			src.BucketCounts().Len(), src.ExplicitBounds().Len(),
		)
	}

	srcCounts := src.BucketCounts().AsRaw()
	counts := make([]int64, len(srcCounts))
	for j := 0; j < len(srcCounts); j++ {
		counts[j] = int64(srcCounts[j])
	}
	dstHistogram.BucketCounts().CopyFromSlice(counts)

	return nil
}
