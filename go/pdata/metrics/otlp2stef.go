package metrics

import (
	"github.com/splunk/stef/go/otel/oteltef"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

// OtlpToStef defines a converter from OTLP metrics to STEF format.
type OtlpToStef interface {
	// WriteMetrics converts OTLP metrics to STEF format and writes them to the provided writer.
	// Will not call Flush() on the writer at the end.
	WriteMetrics(src pmetric.Metrics, writer *oteltef.MetricsWriter) error
}
