package metrics

import (
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/metrics/sortedbymetric"
)

// OtlpToStefSorted converts OTLP metrics to STEF format.
// Data is sorted before conversion, typically improving STEF compression ratio.
type OtlpToStefSorted struct {
}

var _ OtlpToStef = (*OtlpToStefSorted)(nil)

// WriteMetrics converts OTLP metrics and writes to the provided STEF metrics writer.
func (d *OtlpToStefSorted) WriteMetrics(src pmetric.Metrics, writer *oteltef.MetricsWriter) error {
	tree, err := sortedbymetric.OtlpToSourceTree(src)
	if err != nil {
		return err
	}
	return tree.ToStef(writer)
}
