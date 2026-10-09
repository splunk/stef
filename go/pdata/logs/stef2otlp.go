package logs

import (
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/otel/otelstef"
)

// StefToOtlp defines a converter from STEF format to OTLP logs.
type StefToOtlp interface {
	// Convert reads STEF records and returns the corresponding OTLP logs.
	// When untilEOF is false, it stops at the end of the current frame.
	Convert(reader *otelstef.LogsReader, untilEOF bool) (plog.Logs, error)
}
