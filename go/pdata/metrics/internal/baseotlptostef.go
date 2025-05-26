package internal

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
)

type BaseOTLPToSTEF struct {
	TempAttrs oteltef.Attributes
	Otlp2tef  otlptools.Otlp2Stef
}

func (c *BaseOTLPToSTEF) ConvertNumDatapoint(dst *oteltef.Point, src pmetric.NumberDataPoint) {
	dst.SetTimestamp(uint64(src.Timestamp()))
	dst.SetStartTimestamp(uint64(src.StartTimestamp()))

	switch src.ValueType() {
	case pmetric.NumberDataPointValueTypeInt:
		dst.Value().SetInt64(src.IntValue())
	case pmetric.NumberDataPointValueTypeDouble:
		dst.Value().SetFloat64(src.DoubleValue())
	default:
		panic("Unsupported number datapoint value type")
	}
}

func (c *BaseOTLPToSTEF) ConvertExemplars(dst *oteltef.ExemplarArray, src pmetric.ExemplarSlice) {
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

func (c *BaseOTLPToSTEF) ConvertHistogram(dst *oteltef.Point, src pmetric.HistogramDataPoint) error {
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

func (c *BaseOTLPToSTEF) ConvertExpHistogram(dst *oteltef.Point, src pmetric.ExponentialHistogramDataPoint) error {
	dst.SetTimestamp(uint64(src.Timestamp()))
	dst.SetStartTimestamp(uint64(src.StartTimestamp()))

	dstVal := dst.Value()
	dstVal.SetType(oteltef.PointValueTypeExpHistogram)
	dstHistogram := dstVal.ExpHistogram()
	dstHistogram.SetCount(src.Count())

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

	dstHistogram.SetScale(int64(src.Scale()))
	dstHistogram.SetZeroCount(src.ZeroCount())
	dstHistogram.SetZeroThreshold(src.ZeroThreshold())

	expBucketsToStef(dstHistogram.PositiveBuckets(), src.Positive())
	expBucketsToStef(dstHistogram.NegativeBuckets(), src.Negative())

	return nil
}

func expBucketsToStef(
	dst *oteltef.ExpHistogramBuckets, src pmetric.ExponentialHistogramDataPointBuckets,
) {
	dst.SetOffset(int64(src.Offset()))

	srcCounts := src.BucketCounts().AsRaw()
	counts := make([]uint64, len(srcCounts))
	for j := 0; j < len(srcCounts); j++ {
		counts[j] = srcCounts[j]
	}
	dst.BucketCounts().CopyFromSlice(counts)
}
