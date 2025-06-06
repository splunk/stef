package metrics

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/oteltef"

	"github.com/splunk/stef/go/pdata/metrics/internal"
	"github.com/splunk/stef/go/pdata/metrics/sortedbymetric"
)

type OtlpToSortedTree struct {
	internal.BaseOTLPToSTEF
	recordCount         int
	emptyDataPointCount int
	//encoder             anyvalue.Encoder
}

func NewOtlpToSortedTree() *OtlpToSortedTree {
	return &OtlpToSortedTree{}
}

func (c *OtlpToSortedTree) FromOtlp(rms pmetric.ResourceMetricsSlice) (*sortedbymetric.SortedTree, error) {
	sm := sortedbymetric.NewSortedMetrics()

	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sms := rm.ScopeMetrics().At(j)
			for k := 0; k < sms.Metrics().Len(); k++ {
				metric := sms.Metrics().At(k)
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					c.covertNumberDataPoints(sm, rm, sms, metric, metric.Gauge().DataPoints(), 0)
				case pmetric.MetricTypeSum:
					c.covertNumberDataPoints(
						sm, rm, sms, metric, metric.Sum().DataPoints(),
						calcMetricFlags(metric.Sum().IsMonotonic(), metric.Sum().AggregationTemporality()),
					)
				case pmetric.MetricTypeHistogram:
					err := c.covertHistogramDataPoints(sm, rm, sms, metric, metric.Histogram())
					if err != nil {
						return nil, err
					}
				case pmetric.MetricTypeExponentialHistogram:
					err := c.covertExponentialHistogramDataPoints(sm, rm, sms, metric, metric.ExponentialHistogram())
					if err != nil {
						return nil, err
					}
				default:
					panic(fmt.Sprintf("Unsupported metric type: %v (metric name=%s)", metric.Type(), metric.Name()))
				}
			}
		}
	}

	sm.SortValues()

	return sm, nil
}

func calcMetricFlags(monotonic bool, temporality pmetric.AggregationTemporality) internal.MetricFlags {
	var flags internal.MetricFlags
	if monotonic {
		flags |= internal.MetricMonotonic
	}

	switch temporality {
	case pmetric.AggregationTemporalityDelta:
		flags |= internal.MetricTemporalityDelta
	case pmetric.AggregationTemporalityCumulative:
		flags |= internal.MetricTemporalityCumulative
	case pmetric.AggregationTemporalityUnspecified:
		flags |= internal.MetricTemporalityUnspecified
	default:
		panic("Unknown temporality value")
	}
	return flags
}

func (c *OtlpToSortedTree) covertNumberDataPoints(
	sm *sortedbymetric.SortedTree,
	rm pmetric.ResourceMetrics,
	sms pmetric.ScopeMetrics,
	metric pmetric.Metric,
	srcPoints pmetric.NumberDataPointSlice,
	flags internal.MetricFlags,
) {
	var metricType oteltef.MetricType
	var byMetric *sortedbymetric.ByMetric
	var byScope *sortedbymetric.ByScope

	dstPointSlice := make([]oteltef.Point, srcPoints.Len())

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		if srcPoint.ValueType() == pmetric.NumberDataPointValueTypeEmpty {
			c.emptyDataPointCount++
			continue
		}

		c.recordCount++

		mt := calcNumericMetricType(metric)
		if mt != metricType || byMetric == nil {
			metricType = mt
			byMetric = sm.ByMetric(metric, metricType, flags, nil)
			byResource := byMetric.ByResource(rm.Resource(), rm.SchemaUrl())
			byScope = byResource.ByScope(sms.Scope(), sms.SchemaUrl())
		}

		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)

		dstPoint := &dstPointSlice[l]
		dstPoint.Init()

		*dstPoints = append(*dstPoints, dstPoint)
		dstPoint.SetTimestamp(uint64(srcPoint.Timestamp()))
		dstPoint.SetStartTimestamp(uint64(srcPoint.StartTimestamp()))
		c.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars())

		c.ConvertNumDatapoint(dstPoint, srcPoint)
	}
}

func calcNumericMetricType(metric pmetric.Metric) oteltef.MetricType {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return oteltef.MetricTypeGauge
	case pmetric.MetricTypeSum:
		return oteltef.MetricTypeSum
	default:
		panic("Unsupported metric type")
	}
	return 0
}

func (c *OtlpToSortedTree) covertHistogramDataPoints(
	sm *sortedbymetric.SortedTree,
	rm pmetric.ResourceMetrics,
	sms pmetric.ScopeMetrics,
	metric pmetric.Metric,
	hist pmetric.Histogram,
) error {
	var byMetric *sortedbymetric.ByMetric
	var byScope *sortedbymetric.ByScope
	flags := calcMetricFlags(false, hist.AggregationTemporality())
	srcPoints := hist.DataPoints()

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		c.recordCount++

		byMetric = sm.ByMetric(metric, oteltef.MetricTypeHistogram, flags, srcPoint.ExplicitBounds().AsRaw())
		byResource := byMetric.ByResource(rm.Resource(), rm.SchemaUrl())
		byScope = byResource.ByScope(sms.Scope(), sms.SchemaUrl())
		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)

		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)
		dstPoint := oteltef.NewPoint()
		*dstPoints = append(*dstPoints, dstPoint)

		c.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars())

		err := c.ConvertHistogram(dstPoint, srcPoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *OtlpToSortedTree) covertExponentialHistogramDataPoints(
	sm *sortedbymetric.SortedTree,
	rm pmetric.ResourceMetrics,
	sms pmetric.ScopeMetrics,
	metric pmetric.Metric,
	hist pmetric.ExponentialHistogram,
) error {
	var byMetric *sortedbymetric.ByMetric
	var byScope *sortedbymetric.ByScope
	flags := calcMetricFlags(false, hist.AggregationTemporality())
	srcPoints := hist.DataPoints()

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		c.recordCount++

		byMetric = sm.ByMetric(metric, oteltef.MetricTypeExpHistogram, flags, nil)
		byResource := byMetric.ByResource(rm.Resource(), rm.SchemaUrl())
		byScope = byResource.ByScope(sms.Scope(), sms.SchemaUrl())
		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)

		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)
		dstPoint := oteltef.NewPoint()
		*dstPoints = append(*dstPoints, dstPoint)

		c.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars())

		err := c.ConvertExpHistogram(dstPoint, srcPoint)
		if err != nil {
			return err
		}
	}
	return nil
}
