package sortedbymetric

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/otelstef"

	"github.com/splunk/stef/go/pdata/metrics/internal"
)

type converter struct {
	internal.BaseOtlpToStef
}

func OtlpToSortedTree(data pmetric.Metrics) (*SortedTree, error) {
	rms := data.ResourceMetrics()
	sm := NewSortedMetrics()
	c := converter{}

	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		resource := otelstef.NewResource()
		c.Otlp2tef.ResourceSorted(resource, rm.Resource(), rm.SchemaUrl())
		resource.Freeze()

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sms := rm.ScopeMetrics().At(j)
			scope := otelstef.NewScope()
			c.Otlp2tef.ScopeSorted(scope, sms.Scope(), sms.SchemaUrl())
			scope.Freeze()

			for k := 0; k < sms.Metrics().Len(); k++ {
				metric := sms.Metrics().At(k)
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					if err := c.covertNumberDataPoints(
						sm, resource, scope, metric, metric.Gauge().DataPoints(), internal.MetricFlags{},
					); err != nil {
						return nil, err
					}
				case pmetric.MetricTypeSum:
					flags, err := calcMetricFlags(metric.Sum().IsMonotonic(), metric.Sum().AggregationTemporality())
					if err != nil {
						return nil, err
					}
					if err := c.covertNumberDataPoints(
						sm, resource, scope, metric, metric.Sum().DataPoints(), flags,
					); err != nil {
						return nil, err
					}
				case pmetric.MetricTypeHistogram:
					err := c.covertHistogramDataPoints(sm, resource, scope, metric, metric.Histogram())
					if err != nil {
						return nil, err
					}
				case pmetric.MetricTypeExponentialHistogram:
					err := c.covertExponentialHistogramDataPoints(
						sm, resource, scope, metric, metric.ExponentialHistogram(),
					)
					if err != nil {
						return nil, err
					}
				case pmetric.MetricTypeSummary:
					err := c.covertSummaryDataPoints(sm, resource, scope, metric, metric.Summary())
					if err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("unsupported metric type: %v (metric name=%s)", metric.Type(), metric.Name())
				}
			}
		}
	}

	sm.SortValues()

	return sm, nil
}

func calcMetricFlags(monotonic bool, temporality pmetric.AggregationTemporality) (internal.MetricFlags, error) {
	at, err := internal.AggregationTemporalityToStef(temporality)
	if err != nil {
		return internal.MetricFlags{}, err
	}
	return internal.MetricFlags{
		Monotonic:   monotonic,
		Temporality: at,
	}, nil
}

func calcNumericMetricType(metric pmetric.Metric) (otelstef.MetricType, error) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return otelstef.MetricTypeGauge, nil
	case pmetric.MetricTypeSum:
		return otelstef.MetricTypeSum, nil
	default:
		return 0, fmt.Errorf("unsupported numeric metric type: %v", metric.Type())
	}
}

func (c *converter) covertNumberDataPoints(
	sm *SortedTree,
	rm *otelstef.Resource,
	sms *otelstef.Scope,
	metric pmetric.Metric,
	srcPoints pmetric.NumberDataPointSlice,
	flags internal.MetricFlags,
) error {
	var metricType otelstef.MetricType
	var byMetric *ByMetric
	var byScope *ByScope

	dstPointSlice := make([]otelstef.Point, srcPoints.Len())

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		if srcPoint.ValueType() == pmetric.NumberDataPointValueTypeEmpty {
			continue
		}

		mt, err := calcNumericMetricType(metric)
		if err != nil {
			return err
		}
		if mt != metricType || byMetric == nil {
			metricType = mt
			byMetric = sm.ByMetric(metric, metricType, flags, nil)
			byResource := byMetric.ByResource(rm)
			byScope = byResource.ByScope(sms)
		}

		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)

		dstPoint := &dstPointSlice[l]
		dstPoint.Init()

		*dstPoints = append(*dstPoints, dstPoint)
		dstPoint.SetTimestamp(uint64(srcPoint.Timestamp()))
		dstPoint.SetStartTimestamp(uint64(srcPoint.StartTimestamp()))
		if err := c.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars()); err != nil {
			return err
		}

		if err := c.ConvertNumDatapoint(dstPoint, srcPoint); err != nil {
			return err
		}
	}
	return nil
}

func (c *converter) covertHistogramDataPoints(
	sm *SortedTree,
	rm *otelstef.Resource,
	sms *otelstef.Scope,
	metric pmetric.Metric,
	hist pmetric.Histogram,
) error {
	var byMetric *ByMetric
	var byScope *ByScope
	flags, err := calcMetricFlags(false, hist.AggregationTemporality())
	if err != nil {
		return err
	}
	srcPoints := hist.DataPoints()

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		byMetric = sm.ByMetric(metric, otelstef.MetricTypeHistogram, flags, srcPoint.ExplicitBounds().AsRaw())
		byResource := byMetric.ByResource(rm)
		byScope = byResource.ByScope(sms)
		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)

		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)
		dstPoint := otelstef.NewPoint()
		*dstPoints = append(*dstPoints, dstPoint)

		if err := c.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars()); err != nil {
			return err
		}

		err := c.ConvertHistogram(dstPoint, srcPoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *converter) covertExponentialHistogramDataPoints(
	sm *SortedTree,
	rm *otelstef.Resource,
	sms *otelstef.Scope,
	metric pmetric.Metric,
	hist pmetric.ExponentialHistogram,
) error {
	var byMetric *ByMetric
	var byScope *ByScope
	flags, err := calcMetricFlags(false, hist.AggregationTemporality())
	if err != nil {
		return err
	}
	srcPoints := hist.DataPoints()

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		byMetric = sm.ByMetric(metric, otelstef.MetricTypeExpHistogram, flags, nil)
		byResource := byMetric.ByResource(rm)
		byScope = byResource.ByScope(sms)
		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)

		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)
		dstPoint := otelstef.NewPoint()
		*dstPoints = append(*dstPoints, dstPoint)

		if err := c.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars()); err != nil {
			return err
		}

		err := c.ConvertExpHistogram(dstPoint, srcPoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *converter) covertSummaryDataPoints(
	sm *SortedTree,
	rm *otelstef.Resource,
	sms *otelstef.Scope,
	metric pmetric.Metric,
	summary pmetric.Summary,
) error {
	var byMetric *ByMetric
	var byScope *ByScope
	flags := internal.MetricFlags{} // No monotonic/temporality for summary
	srcPoints := summary.DataPoints()

	for l := 0; l < srcPoints.Len(); l++ {
		srcPoint := srcPoints.At(l)

		byMetric = sm.ByMetric(metric, otelstef.MetricTypeSummary, flags, nil)
		byResource := byMetric.ByResource(rm)
		byScope = byResource.ByScope(sms)
		c.Otlp2tef.MapSorted(srcPoint.Attributes(), &c.TempAttrs)
		dstPoints := byScope.ByAttrs(&c.TempAttrs)
		dstPoint := otelstef.NewPoint()
		*dstPoints = append(*dstPoints, dstPoint)

		dstPoint.SetTimestamp(uint64(srcPoint.Timestamp()))
		dstPoint.SetStartTimestamp(uint64(srcPoint.StartTimestamp()))
		// No exemplars for summary

		err := c.ConvertSummary(dstPoint, srcPoint)
		if err != nil {
			return err
		}
	}
	return nil
}
