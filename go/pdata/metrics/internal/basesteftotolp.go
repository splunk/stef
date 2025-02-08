package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
)

type BaseSTEFToOTLP struct {
}

func (s *BaseSTEFToOTLP) convertNumberPoint(src *oteltef.Point, dst pmetric.NumberDataPoint) error {
	dst.SetStartTimestamp(pcommon.Timestamp(src.StartTimestamp()))
	dst.SetTimestamp(pcommon.Timestamp(src.Timestamp()))

	switch src.Value().Type() {
	case oteltef.PointValueTypeInt64:
		dst.SetIntValue(src.Value().Int64())
	case oteltef.PointValueTypeFloat64:
		dst.SetDoubleValue(src.Value().Float64())
	default:
		panic("unexpected type")
	}

	for i := range src.Exemplars().Len() {
		exemplar := src.Exemplars().At(i)
		if err := s.ConvertExemplar(exemplar, dst.Exemplars().AppendEmpty()); err != nil {
			return err
		}
	}
	return nil
}

func (c *BaseSTEFToOTLP) ConvertExemplar(src *oteltef.Exemplar, dst pmetric.Exemplar) error {
	dst.SetTimestamp(pcommon.Timestamp(src.Timestamp()))
	switch src.Value().Type() {
	case oteltef.ExemplarValueTypeInt64:
		dst.SetIntValue(src.Value().Int64())
	case oteltef.ExemplarValueTypeFloat64:
		dst.SetDoubleValue(src.Value().Float64())
	case oteltef.ExemplarValueTypeNone:
	default:
		panic("unknown exemplar value type")
	}
	dst.SetTraceID(pcommon.TraceID([]byte(src.TraceID())))
	dst.SetSpanID(pcommon.SpanID([]byte(src.SpanID())))

	otlpAttrs := pcommon.NewMap()
	err := otlptools.TefToOtlpMap(src.FilteredAttributes(), otlpAttrs)
	if err != nil {
		return err
	}
	otlpAttrs.MoveTo(dst.FilteredAttributes())
	return nil
}

func (s *BaseSTEFToOTLP) AppendOTLPPoint(
	srcMetric *oteltef.Metric, srcAttrs *oteltef.Attributes, srcPoint *oteltef.Point, dstMetric pmetric.Metric,
) error {
	otlpAttrs := pcommon.NewMap()
	err := otlptools.TefToOtlpMap(srcAttrs, otlpAttrs)
	if err != nil {
		return err
	}

	switch dstMetric.Type() {
	case pmetric.MetricTypeGauge:
		point := dstMetric.Gauge().DataPoints().AppendEmpty()
		if err := s.convertNumberPoint(srcPoint, point); err != nil {
			return err
		}
		otlpAttrs.MoveTo(point.Attributes())
	case pmetric.MetricTypeSum:
		point := dstMetric.Sum().DataPoints().AppendEmpty()
		if err := s.convertNumberPoint(srcPoint, point); err != nil {
			return err
		}
		otlpAttrs.MoveTo(point.Attributes())
		dstMetric.Sum().SetIsMonotonic(srcMetric.Monotonic())
		dstMetric.Sum().SetAggregationTemporality(aggregationTemporalityToOtlp(oteltef.MetricFlags(srcMetric.AggregationTemporality())))
	case pmetric.MetricTypeHistogram:
		point := dstMetric.Histogram().DataPoints().AppendEmpty()
		if err := s.convertHistogramPoint(srcMetric, srcPoint, point); err != nil {
			return err
		}
		otlpAttrs.MoveTo(point.Attributes())
		dstMetric.Histogram().SetAggregationTemporality(aggregationTemporalityToOtlp(oteltef.MetricFlags(srcMetric.AggregationTemporality())))
	default:
		panic("not implemented")
	}

	return nil
}

func aggregationTemporalityToOtlp(flags oteltef.MetricFlags) pmetric.AggregationTemporality {
	switch flags & oteltef.MetricTemporalityMask {
	case oteltef.MetricTemporalityDelta:
		return pmetric.AggregationTemporalityDelta
	case oteltef.MetricTemporalityCumulative:
		return pmetric.AggregationTemporalityCumulative
	case oteltef.MetricTemporalityUnspecified:
		return pmetric.AggregationTemporalityUnspecified
	default:
		panic("unexpected metric flags")
	}
}

func (c *BaseSTEFToOTLP) convertHistogramPoint(
	srcMetric *oteltef.Metric, srcPoint *oteltef.Point, dstPoint pmetric.HistogramDataPoint,
) error {
	dstPoint.SetStartTimestamp(pcommon.Timestamp(srcPoint.StartTimestamp()))
	dstPoint.SetTimestamp(pcommon.Timestamp(srcPoint.Timestamp()))
	dstPoint.SetCount(uint64(srcPoint.Value().Histogram().Count()))

	dstPoint.BucketCounts().EnsureCapacity(srcPoint.Value().Histogram().BucketCounts().Len())
	for i := 0; i < srcPoint.Value().Histogram().BucketCounts().Len(); i++ {
		dstPoint.BucketCounts().Append(uint64(srcPoint.Value().Histogram().BucketCounts().At(i)))
	}

	dstPoint.ExplicitBounds().EnsureCapacity(srcMetric.HistogramBounds().Len())
	for i := range srcMetric.HistogramBounds().Len() {
		dstPoint.ExplicitBounds().Append(srcMetric.HistogramBounds().At(i))
	}

	if srcPoint.Value().Histogram().HasSum() {
		dstPoint.SetSum(srcPoint.Value().Histogram().Sum())
	}

	if srcPoint.Value().Histogram().HasMin() {
		dstPoint.SetMin(srcPoint.Value().Histogram().Min())
	}

	if srcPoint.Value().Histogram().HasMax() {
		dstPoint.SetMax(srcPoint.Value().Histogram().Max())
	}

	for i := range srcPoint.Exemplars().Len() {
		exemplar := srcPoint.Exemplars().At(i)
		if err := c.ConvertExemplar(exemplar, dstPoint.Exemplars().AppendEmpty()); err != nil {
			return err
		}
	}
	return nil
}
