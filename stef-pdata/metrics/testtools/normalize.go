package testtools

import (
	"errors"
	"slices"
	"strings"

	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/tigrannajaryan/stef/stef-pdata/internal/otlptools"
)

func NormalizeMetrics(data pmetric.Metrics) {
	resourceMetrics := data.ResourceMetrics()

	for i := 0; i < resourceMetrics.Len(); i++ {
		sortAttrs(resourceMetrics.At(i).Resource().Attributes())
	}

	resourceMetrics.Sort(
		func(a, b pmetric.ResourceMetrics) bool {
			return otlptools.CmpResourceMetrics(a, b) < 0
		},
	)

	for i := 0; i < resourceMetrics.Len()-1; {
		if otlptools.CmpResourceMetrics(resourceMetrics.At(i), resourceMetrics.At(i+1)) == 0 {
			resourceMetrics.At(i + 1).ScopeMetrics().MoveAndAppendTo(resourceMetrics.At(i).ScopeMetrics())
			j := 0
			resourceMetrics.RemoveIf(
				func(metrics pmetric.ResourceMetrics) bool {
					j++
					return j == i+2
				},
			)
		} else {
			i++
		}
	}

	for i := 0; i < resourceMetrics.Len(); i++ {
		normalizeScopeMetrics(resourceMetrics.At(i).ScopeMetrics())
	}
}

func normalizeScopeMetrics(scopeMetrics pmetric.ScopeMetricsSlice) {
	for i := 0; i < scopeMetrics.Len(); i++ {
		sortAttrs(scopeMetrics.At(i).Scope().Attributes())
	}

	scopeMetrics.Sort(
		func(a, b pmetric.ScopeMetrics) bool {
			return otlptools.CmpScopeMetrics(a, b) < 0
		},
	)

	for i := 0; i < scopeMetrics.Len()-1; {
		if otlptools.CmpScopeMetrics(scopeMetrics.At(i), scopeMetrics.At(i+1)) == 0 {
			scopeMetrics.At(i + 1).Metrics().MoveAndAppendTo(scopeMetrics.At(i).Metrics())
			j := 0
			scopeMetrics.RemoveIf(
				func(s pmetric.ScopeMetrics) bool {
					j++
					return j == i+2
				},
			)
		} else {
			i++
		}
	}

	for i := 0; i < scopeMetrics.Len(); i++ {
		normalizeMetrics(scopeMetrics.At(i).Metrics())
	}
}

func normalizeMetrics(metrics pmetric.MetricSlice) {
	for i := 0; i < metrics.Len(); i++ {
		sortAttrs(metrics.At(i).Metadata())
	}

	metrics.Sort(
		func(a, b pmetric.Metric) bool {
			return cmpMetric(a, b) < 0
		},
	)

	for i := 0; i < metrics.Len()-1; {
		if cmpMetric(metrics.At(i), metrics.At(i+1)) == 0 {
			appendMetricData(metrics.At(i), metrics.At(i+1))
			j := 0
			metrics.RemoveIf(
				func(s pmetric.Metric) bool {
					j++
					return j == i+2
				},
			)
		} else {
			i++
		}
	}

	for i := 0; i < metrics.Len(); i++ {
		normalizeMetricData(metrics.At(i))
	}
}

func normalizeMetricData(metric pmetric.Metric) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		normalizeNumberDatapointAttrs(metric.Gauge().DataPoints())
		metric.Gauge().DataPoints().Sort(
			func(a, b pmetric.NumberDataPoint) bool {
				return cmpNumberDataPoint(a, b) < 0
			},
		)
	case pmetric.MetricTypeSum:
		normalizeNumberDatapointAttrs(metric.Sum().DataPoints())
		metric.Sum().DataPoints().Sort(
			func(a, b pmetric.NumberDataPoint) bool {
				return cmpNumberDataPoint(a, b) < 0
			},
		)
	case pmetric.MetricTypeHistogram:
		normalizeHistogramDatapointAttrs(metric.Histogram().DataPoints())
		metric.Histogram().DataPoints().Sort(
			func(a, b pmetric.HistogramDataPoint) bool {
				return cmpHistogramDataPoint(a, b) < 0
			},
		)
	default:
		panic("not implemented")
	}

}

func normalizeNumberDatapointAttrs(points pmetric.NumberDataPointSlice) {
	points.RemoveIf(
		func(point pmetric.NumberDataPoint) bool {
			return point.ValueType() == pmetric.NumberDataPointValueTypeEmpty
		},
	)

	for i := 0; i < points.Len(); i++ {
		point := points.At(i)
		sortAttrs(point.Attributes())
		normalizeExemplars(point.Exemplars())
	}
}

func normalizeExemplars(exemplars pmetric.ExemplarSlice) {
	for j := 0; j < exemplars.Len(); j++ {
		sortAttrs(exemplars.At(j).FilteredAttributes())
	}
}

func normalizeHistogramDatapointAttrs(points pmetric.HistogramDataPointSlice) {
	for i := 0; i < points.Len(); i++ {
		point := points.At(i)
		sortAttrs(point.Attributes())
		normalizeExemplars(point.Exemplars())
	}
}

func cmpHistogramDataPoint(left pmetric.HistogramDataPoint, right pmetric.HistogramDataPoint) int {
	c := otlptools.CmpAttrs(left.Attributes(), right.Attributes())
	if c != 0 {
		return c
	}

	if left.Timestamp() < right.Timestamp() {
		return -1
	}
	if left.Timestamp() > right.Timestamp() {
		return 1
	}
	return 0
}

func cmpNumberDataPoint(left pmetric.NumberDataPoint, right pmetric.NumberDataPoint) int {
	c := otlptools.CmpAttrs(left.Attributes(), right.Attributes())
	if c != 0 {
		return c
	}

	if left.Timestamp() < right.Timestamp() {
		return -1
	}
	if left.Timestamp() > right.Timestamp() {
		return 1
	}
	return 0
}

func appendMetricData(left, right pmetric.Metric) {
	switch left.Type() {
	case pmetric.MetricTypeGauge:
		right.Gauge().DataPoints().MoveAndAppendTo(left.Gauge().DataPoints())
	case pmetric.MetricTypeSum:
		right.Sum().DataPoints().MoveAndAppendTo(left.Sum().DataPoints())
	case pmetric.MetricTypeHistogram:
		right.Histogram().DataPoints().MoveAndAppendTo(left.Histogram().DataPoints())
	default:
		panic("not implemented")
	}
}

func sortAttrs(attributes pcommon.Map) {
	attrs := otlptools.Map2attrs(attributes)
	slices.SortFunc(
		attrs, func(a, b otlptools.AttrAccessible) int {
			return cmpAttr(a, b)
		},
	)
	attributes.Clear()
	attributes.EnsureCapacity(len(attrs))
	for _, a := range attrs {
		v := attributes.PutEmpty(a.Key)
		a.Value.CopyTo(v)
	}
}

func cmpMetric(left, right pmetric.Metric) int {
	c := strings.Compare(left.Name(), right.Name())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.Unit(), right.Unit())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.Description(), right.Description())
	if c != 0 {
		return c
	}
	c = otlptools.CmpAttrs(left.Metadata(), right.Metadata())
	if c != 0 {
		return c
	}
	return metricTypeIndex(left) - metricTypeIndex(right)
}

func metricTypeIndex(metric pmetric.Metric) int {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return 0
	case pmetric.MetricTypeSum:
		return 1
	case pmetric.MetricTypeHistogram:
		return 2
	case pmetric.MetricTypeSummary:
		return 3
	case pmetric.MetricTypeExponentialHistogram:
		return 4
	default:
		panic("unknown metric type")
	}
}

func cmpAttr(left otlptools.AttrAccessible, right otlptools.AttrAccessible) int {
	c := strings.Compare(left.Key, right.Key)
	if c != 0 {
		return c
	}
	return otlptools.CmpVal(left.Value, right.Value)
}

func DiffMetrics(left, right pmetric.Metrics) error {
	if left.ResourceMetrics().Len() != right.ResourceMetrics().Len() {
		return errors.New("ResourceMetrics count mismatch")
	}

	for i := 0; i < left.ResourceMetrics().Len(); i++ {
		if err := diffResourceMetrics(left.ResourceMetrics().At(i), right.ResourceMetrics().At(i)); err != nil {
			return err
		}
	}
	return nil
}

func diffResourceMetrics(left pmetric.ResourceMetrics, right pmetric.ResourceMetrics) error {
	c := strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return errors.New(cmp.Diff(left.SchemaUrl(), right.SchemaUrl()))
	}
	err := diffAttrs(left.Resource().Attributes(), right.Resource().Attributes())
	if err != nil {
		return err
	}

	if left.ScopeMetrics().Len() != right.ScopeMetrics().Len() {
		return errors.New("ScopeMetrics count mismatch")
	}

	for i := 0; i < left.ScopeMetrics().Len(); i++ {
		if err := diffScopeMetrics(left.ScopeMetrics().At(i), right.ScopeMetrics().At(i)); err != nil {
			return err
		}
	}
	return nil
}

func diffScopeMetrics(left pmetric.ScopeMetrics, right pmetric.ScopeMetrics) error {
	c := strings.Compare(left.Scope().Name(), right.Scope().Name())
	if c != 0 {
		return errors.New(cmp.Diff(left.Scope().Name(), right.Scope().Name()))
	}
	c = strings.Compare(left.Scope().Version(), right.Scope().Version())
	if c != 0 {
		return errors.New(cmp.Diff(left.Scope().Version(), right.Scope().Version()))
	}
	c = strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return errors.New(cmp.Diff(left.SchemaUrl(), right.SchemaUrl()))
	}

	err := diffAttrs(left.Scope().Attributes(), right.Scope().Attributes())
	if err != nil {
		return err
	}

	if left.Metrics().Len() != right.Metrics().Len() {
		return errors.New("Metrics count mismatch")
	}

	for i := 0; i < left.Metrics().Len(); i++ {
		if err := diffMetric(left.Metrics().At(i), right.Metrics().At(i)); err != nil {
			return err
		}
	}

	return nil
}

func diffMetric(left pmetric.Metric, right pmetric.Metric) error {
	c := strings.Compare(left.Name(), right.Name())
	if c != 0 {
		return errors.New(cmp.Diff(left.Name(), right.Name()))
	}
	c = strings.Compare(left.Unit(), right.Unit())
	if c != 0 {
		return errors.New(cmp.Diff(left.Unit(), right.Unit()))
	}
	c = strings.Compare(left.Description(), right.Description())
	if c != 0 {
		return errors.New(cmp.Diff(left.Description(), right.Description()))
	}
	err := diffAttrs(left.Metadata(), right.Metadata())
	if err != nil {
		return err
	}
	if metricTypeIndex(left) != metricTypeIndex(right) {
		return errors.New("metric types are different")
	}

	switch left.Type() {
	case pmetric.MetricTypeGauge:
		err = diffNumberDataPoints(left.Gauge().DataPoints(), right.Gauge().DataPoints())
	case pmetric.MetricTypeSum:
		if str := cmp.Diff(left.Sum().AggregationTemporality(), right.Sum().AggregationTemporality()); str != "" {
			return errors.New("temporarily is different: " + str)
		}
		if str := cmp.Diff(left.Sum().IsMonotonic(), right.Sum().IsMonotonic()); str != "" {
			return errors.New("monotonicity is different: " + str)
		}
		err = diffNumberDataPoints(left.Sum().DataPoints(), right.Sum().DataPoints())
	case pmetric.MetricTypeHistogram:
		err = diffHistogram(left.Histogram(), right.Histogram())
	default:
		panic("unknown metric data")
	}

	return err
}

func diffHistogram(left, right pmetric.Histogram) error {
	if str := cmp.Diff(left.AggregationTemporality(), right.AggregationTemporality()); str != "" {
		return errors.New(str)
	}

	if left.DataPoints().Len() != right.DataPoints().Len() {
		return errors.New("DataPoints count mismatch")
	}

	for i := 0; i < left.DataPoints().Len(); i++ {
		if err := diffHistogramDataPoint(left.DataPoints().At(i), right.DataPoints().At(i)); err != nil {
			return err
		}
	}
	return nil
}

func diffNumberDataPoints(left, right pmetric.NumberDataPointSlice) error {
	if left.Len() != right.Len() {
		return errors.New("DataPoints count mismatch")
	}

	for i := 0; i < left.Len(); i++ {
		if err := diffNumberDataPoint(left.At(i), right.At(i)); err != nil {
			return err
		}
	}

	return nil
}

func diffNumberDataPoint(left, right pmetric.NumberDataPoint) error {
	if str := cmp.Diff(left.StartTimestamp(), right.StartTimestamp()); str != "" {
		return errors.New(str)
	}
	if str := cmp.Diff(left.Timestamp(), right.Timestamp()); str != "" {
		return errors.New(str)
	}
	if str := cmp.Diff(left.Flags(), right.Flags()); str != "" {
		return errors.New(str)
	}
	if str := cmp.Diff(left.ValueType(), right.ValueType()); str != "" {
		return errors.New("Values are different: " + str)
	}
	switch left.ValueType() {
	case pmetric.NumberDataPointValueTypeInt:
		if str := cmp.Diff(left.IntValue(), right.IntValue()); str != "" {
			return errors.New("Values are different: " + str)
		}
	case pmetric.NumberDataPointValueTypeDouble:
		if str := cmp.Diff(left.DoubleValue(), right.DoubleValue()); str != "" {
			return errors.New("Values are different: " + str)
		}
	}
	err := diffAttrs(left.Attributes(), right.Attributes())
	if err != nil {
		return err
	}
	return diffExemplars(left.Exemplars(), right.Exemplars())
}

func diffExemplars(left, right pmetric.ExemplarSlice) error {
	if left.Len() != right.Len() {
		return errors.New("Exemplar count mismatch")
	}

	for i := 0; i < left.Len(); i++ {
		if err := diffExemplar(left.At(i), right.At(i)); err != nil {
			return err
		}
	}

	return nil
}

func diffExemplar(left, right pmetric.Exemplar) error {
	if str := cmp.Diff(left.Timestamp(), right.Timestamp()); str != "" {
		return errors.New(str)
	}

	if str := cmp.Diff(left.ValueType(), right.ValueType()); str != "" {
		return errors.New(str)
	}

	switch left.ValueType() {
	case pmetric.ExemplarValueTypeInt:
		if str := cmp.Diff(left.IntValue(), right.IntValue()); str != "" {
			return errors.New(str)
		}
	case pmetric.ExemplarValueTypeDouble:
		if str := cmp.Diff(left.DoubleValue(), right.DoubleValue()); str != "" {
			return errors.New(str)
		}
	}

	if str := cmp.Diff(left.TraceID(), right.TraceID()); str != "" {
		return errors.New("Exemplar TraceID mismatch: " + str)
	}

	if str := cmp.Diff(left.SpanID(), right.SpanID()); str != "" {
		return errors.New("Exemplar SpanID mismatch: " + str)
	}

	return diffAttrs(left.FilteredAttributes(), right.FilteredAttributes())
}

func diffHistogramDataPoint(left, right pmetric.HistogramDataPoint) error {
	if str := cmp.Diff(left.StartTimestamp(), right.StartTimestamp()); str != "" {
		return errors.New(str)
	}
	if str := cmp.Diff(left.Timestamp(), right.Timestamp()); str != "" {
		return errors.New(str)
	}
	if str := cmp.Diff(left.Flags(), right.Flags()); str != "" {
		return errors.New(str)
	}
	err := diffExemplars(left.Exemplars(), right.Exemplars())
	if err != nil {
		return err
	}

	if str := cmp.Diff(left.HasSum(), right.HasSum()); str != "" {
		return errors.New("HasSum is different: " + str)
	}
	if str := cmp.Diff(left.Sum(), right.Sum()); str != "" {
		return errors.New("Sum is different: " + str)
	}

	if str := cmp.Diff(left.HasMin(), right.HasMin()); str != "" {
		return errors.New("HasMin is different: " + str)
	}
	if str := cmp.Diff(left.Min(), right.Min()); str != "" {
		return errors.New("Min is different: " + str)
	}

	if str := cmp.Diff(left.HasMax(), right.HasMax()); str != "" {
		return errors.New("HasMin is different: " + str)
	}
	if str := cmp.Diff(left.Max(), right.Max()); str != "" {
		return errors.New("Max is different: " + str)
	}

	if str := cmp.Diff(left.Count(), right.Count()); str != "" {
		return errors.New("Count is different: " + str)
	}

	if str := cmp.Diff(left.BucketCounts().AsRaw(), right.BucketCounts().AsRaw()); str != "" {
		return errors.New("BucketCounts are different" + str)
	}
	if str := cmp.Diff(left.ExplicitBounds().AsRaw(), right.ExplicitBounds().AsRaw()); str != "" {
		return errors.New(str)
	}
	return diffAttrs(left.Attributes(), right.Attributes())
}

func diffAttrs(a, b pcommon.Map) error {
	if a.Len() != b.Len() {
		return errors.New("attribute length not equal")
	}

	left := otlptools.Map2attrs(a)
	right := otlptools.Map2attrs(b)

	l := min(len(left), len(right))
	for i := 0; i < l; i++ {
		c := strings.Compare(left[i].Key, right[i].Key)
		if c != 0 {
			return errors.New("Keys are different: " + cmp.Diff(left[i].Key, right[i].Key))
		}
	}
	for i := 0; i < l; i++ {
		c := otlptools.CmpVal(left[i].Value, right[i].Value)
		if c != 0 {
			return errors.New("values are different")
		}
	}
	return nil
}
