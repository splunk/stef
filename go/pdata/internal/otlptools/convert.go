package otlptools

import (
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/otelstef"
)

func ResourceToOtlp(resource *otelstef.Resource, out pmetric.ResourceMetrics) error {
	if resource == nil {
		return nil
	}
	out.SetSchemaUrl(resource.SchemaURL())
	var err error
	err = TefToOtlpMap(resource.Attributes(), out.Resource().Attributes())
	if err != nil {
		return err
	}
	out.Resource().SetDroppedAttributesCount(uint32(resource.DroppedAttributesCount()))
	return nil
}

func ScopeToOtlp(scope *otelstef.Scope, out pmetric.ScopeMetrics) error {
	if scope == nil {
		return nil
	}
	out.SetSchemaUrl(scope.SchemaURL())
	out.Scope().SetName(scope.Name())
	out.Scope().SetVersion(scope.Version())
	var err error
	err = TefToOtlpMap(scope.Attributes(), out.Scope().Attributes())
	if err != nil {
		return err
	}
	out.Scope().SetDroppedAttributesCount(uint32(scope.DroppedAttributesCount()))
	return nil
}

func MetricToOtlp(metric *otelstef.Metric, out pmetric.Metric) error {
	if metric == nil {
		return nil
	}
	out.SetName(metric.Name())
	out.SetUnit(metric.Unit())
	out.SetDescription(metric.Description())

	var err error
	err = TefToOtlpMap(metric.Metadata(), out.Metadata())
	if err != nil {
		return err
	}

	switch otelstef.MetricType(metric.Type()) {
	case otelstef.MetricTypeGauge:
		out.SetEmptyGauge()
	case otelstef.MetricTypeSum:
		out.SetEmptySum()
	case otelstef.MetricTypeHistogram:
		out.SetEmptyHistogram()
	case otelstef.MetricTypeExpHistogram:
		out.SetEmptyExponentialHistogram()
	case otelstef.MetricTypeSummary:
		out.SetEmptySummary()
	default:
		panic("not implemented")
	}

	return nil
}
