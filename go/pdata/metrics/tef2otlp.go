package metrics

import (
	"errors"
	"io"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
	"github.com/splunk/stef/go/pdata/metrics/internal"
)

type TEFToOTLPUnsorted struct {
	internal.BaseTEFToOTLP
}

func (c *TEFToOTLPUnsorted) Convert(reader *oteltef.MetricsReader) (pmetric.Metrics, error) {
	var resourceMetrics pmetric.ResourceMetrics
	var scopeMetrics pmetric.ScopeMetrics
	var metric pmetric.Metric

	metrics := pmetric.NewMetrics()
	modified := true
	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return metrics, err
		}

		if modified || record.IsResourceModified() {
			modified = true
			resourceMetrics = metrics.ResourceMetrics().AppendEmpty()
			if err := otlptools.ResourceToOtlp(record.Resource(), resourceMetrics); err != nil {
				return metrics, err
			}
		}

		if modified || record.IsScopeModified() {
			modified = true
			scopeMetrics = resourceMetrics.ScopeMetrics().AppendEmpty()
			if err := otlptools.ScopeToOtlp(record.Scope(), scopeMetrics); err != nil {
				return metrics, err
			}
		}

		if modified || record.IsMetricModified() {
			metric = scopeMetrics.Metrics().AppendEmpty()
			if err := otlptools.MetricToOtlp(record.Metric(), metric); err != nil {
				return metrics, err
			}
		}

		err = c.AppendOTLPPoint(record.Metric(), record.Attributes(), record.Point(), metric)
		if err != nil {
			return metrics, err
		}

		modified = false
	}

	return metrics, nil
}
