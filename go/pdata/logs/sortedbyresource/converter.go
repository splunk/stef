package sortedbyresource

import (
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
	"github.com/splunk/stef/go/pkg"
)

func OtlpToSortedTree(data plog.Logs) *SortedTree {
	tree := NewSortedLogs()
	converter := otlptools.Otlp2Stef{}
	resourceLogs := data.ResourceLogs()

	for i := 0; i < resourceLogs.Len(); i++ {
		srcResource := resourceLogs.At(i)
		resource := otelstef.NewResource()
		converter.ResourceSorted(resource, srcResource.Resource(), srcResource.SchemaUrl())
		resource.Freeze()
		byResource := tree.ByResource(resource)

		scopeLogs := srcResource.ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			srcScope := scopeLogs.At(j)
			scope := otelstef.NewScope()
			converter.ScopeSorted(scope, srcScope.Scope(), srcScope.SchemaUrl())
			scope.Freeze()
			byScope := byResource.ByScope(scope)

			logRecords := srcScope.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				log := otelstef.NewLogRecord()
				logRecordToStef(logRecords.At(k), log, &converter)
				byScope.Append(log)
			}
		}
	}

	tree.SortValues()
	return tree
}

func logRecordToStef(src plog.LogRecord, dst *otelstef.LogRecord, converter *otlptools.Otlp2Stef) {
	dst.SetTimeUnixNano(uint64(src.Timestamp()))
	dst.SetObservedTimeUnixNano(uint64(src.ObservedTimestamp()))
	dst.SetSeverityNumber(otelstef.SeverityNumber(src.SeverityNumber()))
	dst.SetSeverityText(src.SeverityText())
	converter.Value(src.Body(), dst.Body())
	converter.MapSorted(src.Attributes(), dst.Attributes())
	dst.SetDroppedAttributesCount(uint64(src.DroppedAttributesCount()))
	dst.SetFlags(uint64(src.Flags()))
	traceID := src.TraceID()
	spanID := src.SpanID()
	dst.SetTraceID(pkg.Bytes(traceID[:]))
	dst.SetSpanID(pkg.Bytes(spanID[:]))
	dst.SetEventName(src.EventName())
}
