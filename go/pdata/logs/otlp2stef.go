package logs

import (
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/otel/otelstef"
)

// OtlpToStef defines a converter from OTLP logs to STEF format.
type OtlpToStef interface {
	// Convert writes OTLP logs to the provided STEF writer.
	// It does not call Flush on the writer.
	Convert(src plog.Logs, writer *otelstef.LogsWriter) error
}

// NewOtlpToStef returns an OTLP logs converter. When sorted is true, the
// converter groups equal resources and scopes and sorts log records by time.
func NewOtlpToStef(sorted bool) OtlpToStef {
	return &OtlpToStefUnsorted{Sorted: sorted}
}
