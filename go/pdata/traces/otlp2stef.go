package traces

import (
	"github.com/splunk/stef/go/otel/oteltef"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

// OtlpToStef defines a converter from OTLP traces to STEF format.
type OtlpToStef interface {
	// Convert OTLP traces to STEF format and write them to the provided writer.
	// Will not call Flush() on the writer at the end.
	Convert(src ptrace.Traces, writer *oteltef.SpansWriter) error
}
