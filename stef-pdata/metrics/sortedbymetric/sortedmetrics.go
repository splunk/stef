package sortedbymetric

import (
	"io"
	"slices"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"modernc.org/b/v2"

	"github.com/splunk/stef/stef-go/pkg"

	"github.com/splunk/stef/stef-otel/oteltef"

	"github.com/splunk/stef/stef-pdata/internal/otlptools"
)

type SortedTree struct {
	//encoder  anyvalue.Encoder
	otlp2tef otlptools.Otlp2Tef
	byMetric *b.Tree[*oteltef.Metric, *ByMetric]
}

type ByMetric struct {
	//encoder    *anyvalue.Encoder
	otlp2tef   *otlptools.Otlp2Tef
	byResource *b.Tree[*oteltef.Resource, *ByResource]
}

type ByResource struct {
	//encoder  *anyvalue.Encoder
	otlp2tef *otlptools.Otlp2Tef
	byScope  *b.Tree[*oteltef.Scope, *ByScope]
}

type Points []*oteltef.Point

func (p Points) SortValues() {
	slices.SortFunc(
		p, func(a, b *oteltef.Point) int {
			return pkg.Uint64Compare(a.Timestamp(), b.Timestamp())
		},
	)
}

type ByScope struct {
	byAttrs *b.Tree[*oteltef.Attributes, *Points]
}

func NewSortedMetrics() *SortedTree {
	return &SortedTree{byMetric: b.TreeNew[*oteltef.Metric, *ByMetric](oteltef.CmpMetric)}
}

func (s *SortedTree) ToTef(writer *oteltef.MetricsWriter) error {
	i := 0
	err := s.Iter(
		func(metric *oteltef.Metric, byMetric *ByMetric) error {
			writer.Record.Metric().CopyFrom(metric)
			err := byMetric.Iter(
				func(resource *oteltef.Resource, byResource *ByResource) error {
					writer.Record.Resource().CopyFrom(resource)
					err := byResource.Iter(
						func(scope *oteltef.Scope, byScope *ByScope) error {
							writer.Record.Scope().CopyFrom(scope)
							err := byScope.Iter(
								func(attrs *oteltef.Attributes, points *Points) error {
									writer.Record.Attributes().CopyFrom(attrs)
									for _, value := range *points {
										writer.Record.Point().CopyFrom(value)
										if err := writer.Write(); err != nil {
											return err
										}
										i++
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
	return err
}

func (s *SortedTree) ByMetric(
	metric pmetric.Metric, metricType oteltef.MetricType, flags oteltef.MetricFlags,
	histogramBounds []float64,
) *ByMetric {
	metr := metric2metric(metric, metricType, flags, histogramBounds, &s.otlp2tef)
	elem, exists := s.byMetric.Get(metr)
	if !exists {
		elem = &ByMetric{
			otlp2tef: &s.otlp2tef, byResource: b.TreeNew[*oteltef.Resource, *ByResource](oteltef.CmpResource),
		}
		s.byMetric.Set(metr, elem)
	}
	return elem
}

func metric2metric(
	metric pmetric.Metric, metricType oteltef.MetricType, flags oteltef.MetricFlags, histogramBounds []float64,
	otlp2tef *otlptools.Otlp2Tef,
) *oteltef.Metric {

	var dst oteltef.Metric
	otlp2tef.MapSorted(metric.Metadata(), dst.Metadata())
	dst.SetName(metric.Name())
	dst.SetDescription(metric.Description())
	dst.SetUnit(metric.Unit())
	dst.SetType(uint64(metricType))
	//dst.SetFlags(uint64(flags))
	dst.HistogramBounds().CopyFromSlice(histogramBounds)
	dst.SetMonotonic(flags&oteltef.MetricMonotonic != 0)
	dst.SetAggregationTemporality(uint64(flags & oteltef.MetricTemporalityMask))

	return &dst
}

func (s *SortedTree) SortValues() {
	iter, err := s.byMetric.SeekFirst()
	if err != nil {
		return
	}
	for {
		_, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		v.SortValues()
	}
}

func (s *SortedTree) Iter(f func(metric *oteltef.Metric, byMetric *ByMetric) error) error {
	iter, err := s.byMetric.SeekFirst()
	if err != nil {
		return nil
	}
	for {
		k, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (m *ByMetric) ByResource(resource pcommon.Resource, schemaUrl string) *ByResource {
	var res oteltef.Resource
	m.otlp2tef.ResourceSorted(&res, resource, schemaUrl)
	elem, exists := m.byResource.Get(&res)
	if !exists {
		elem = &ByResource{otlp2tef: m.otlp2tef, byScope: b.TreeNew[*oteltef.Scope, *ByScope](oteltef.CmpScope)}
		m.byResource.Set(&res, elem)
	}
	return elem
}

func (m *ByMetric) SortValues() {
	iter, err := m.byResource.SeekFirst()
	if err != nil {
		return
	}
	for {
		_, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		v.SortValues()
	}
}

func (m *ByMetric) Iter(f func(resource *oteltef.Resource, byResource *ByResource) error) error {
	iter, err := m.byResource.SeekFirst()
	if err != nil {
		return nil
	}
	for {
		k, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (m *ByResource) ByScope(scope pcommon.InstrumentationScope, schemaUrl string) *ByScope {
	var dst oteltef.Scope
	m.otlp2tef.ScopeSorted(&dst, scope, schemaUrl)
	elem, exists := m.byScope.Get(&dst)
	if !exists {
		elem = &ByScope{byAttrs: b.TreeNew[*oteltef.Attributes, *Points](oteltef.CmpAttributes)}
		m.byScope.Set(&dst, elem)
	}
	return elem
}

func (m *ByResource) Iter(f func(scope *oteltef.Scope, byScope *ByScope) error) error {
	iter, err := m.byScope.SeekFirst()
	if err != nil {
		return nil
	}
	for {
		k, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (m *ByResource) SortValues() {
	iter, err := m.byScope.SeekFirst()
	if err != nil {
		return
	}
	for {
		_, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		v.SortValues()
	}
}

func (m *ByScope) ByAttrs(attrs *oteltef.Attributes) *Points {
	elem, exists := m.byAttrs.Get(attrs)
	if !exists {
		elem = new(Points)
		var attrsClone oteltef.Attributes
		attrsClone.CopyFrom(attrs)
		m.byAttrs.Set(&attrsClone, elem)
	}
	return elem
}

func (m *ByScope) SortValues() {
	iter, err := m.byAttrs.SeekFirst()
	if err != nil {
		return
	}
	for {
		_, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		v.SortValues()
	}
}

func (m *ByScope) Iter(f func(attrs *oteltef.Attributes, points *Points) error) error {
	iter, err := m.byAttrs.SeekFirst()
	if err != nil {
		return nil
	}
	for {
		k, v, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}
