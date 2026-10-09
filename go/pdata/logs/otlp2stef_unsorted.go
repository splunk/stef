package logs

import (
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
	"github.com/splunk/stef/go/pkg"
)

// OtlpToStefUnsorted converts OTLP logs to STEF. If Sorted is true, it first
// groups equal resources and scopes and orders log records by timestamp.
type OtlpToStefUnsorted struct {
	otlp2stef otlptools.Otlp2Stef
	Sorted    bool
}

var _ OtlpToStef = (*OtlpToStefUnsorted)(nil)

func (c *OtlpToStefUnsorted) Convert(src plog.Logs, writer *otelstef.LogsWriter) error {
	otlp2stef := &c.otlp2stef
	resourceLogs := src.ResourceLogs()

	if c.Sorted {
		resourceLogs.Sort(func(a, b plog.ResourceLogs) bool {
			return otlptools.CmpResourceLogs(a, b) < 0
		})
		for i := 0; i < resourceLogs.Len()-1; {
			if otlptools.CmpResourceLogs(resourceLogs.At(i), resourceLogs.At(i+1)) != 0 {
				i++
				continue
			}
			resourceLogs.At(i + 1).ScopeLogs().MoveAndAppendTo(resourceLogs.At(i).ScopeLogs())
			removeResourceLogsAt(resourceLogs, i+1)
		}
	}

	for i := 0; i < resourceLogs.Len(); i++ {
		rl := resourceLogs.At(i)
		if c.Sorted {
			otlp2stef.ResourceSorted(writer.Record.Resource(), rl.Resource(), rl.SchemaUrl())
		} else {
			otlp2stef.ResourceUnsorted(writer.Record.Resource(), rl.Resource(), rl.SchemaUrl())
		}
		scopeLogs := rl.ScopeLogs()

		if c.Sorted {
			scopeLogs.Sort(func(a, b plog.ScopeLogs) bool {
				return otlptools.CmpScopeLogs(a, b) < 0
			})
			for j := 0; j < scopeLogs.Len()-1; {
				if otlptools.CmpScopeLogs(scopeLogs.At(j), scopeLogs.At(j+1)) != 0 {
					j++
					continue
				}
				scopeLogs.At(j + 1).LogRecords().MoveAndAppendTo(scopeLogs.At(j).LogRecords())
				removeScopeLogsAt(scopeLogs, j+1)
			}
		}

		for j := 0; j < scopeLogs.Len(); j++ {
			sl := scopeLogs.At(j)
			if c.Sorted {
				otlp2stef.ScopeSorted(writer.Record.Scope(), sl.Scope(), sl.SchemaUrl())
				sortLogRecords(sl.LogRecords())
			} else {
				otlp2stef.ScopeUnsorted(writer.Record.Scope(), sl.Scope(), sl.SchemaUrl())
			}

			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecordToStef(sl.LogRecords().At(k), writer.Record.Log(), otlp2stef, c.Sorted)
				if err := writer.Write(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func removeResourceLogsAt(logs plog.ResourceLogsSlice, index int) {
	i := 0
	logs.RemoveIf(func(plog.ResourceLogs) bool {
		remove := i == index
		i++
		return remove
	})
}

func removeScopeLogsAt(logs plog.ScopeLogsSlice, index int) {
	i := 0
	logs.RemoveIf(func(plog.ScopeLogs) bool {
		remove := i == index
		i++
		return remove
	})
}

func sortLogRecords(logs plog.LogRecordSlice) {
	logs.Sort(func(a, b plog.LogRecord) bool {
		return otlptools.CmpLogRecords(a, b) < 0
	})
}

func logRecordToStef(src plog.LogRecord, dst *otelstef.LogRecord, converter *otlptools.Otlp2Stef, sorted bool) {
	dst.SetTimeUnixNano(uint64(src.Timestamp()))
	dst.SetObservedTimeUnixNano(uint64(src.ObservedTimestamp()))
	dst.SetSeverityNumber(otelstef.SeverityNumber(src.SeverityNumber()))
	dst.SetSeverityText(src.SeverityText())
	converter.Value(src.Body(), dst.Body())
	if sorted {
		converter.MapSorted(src.Attributes(), dst.Attributes())
	} else {
		converter.MapUnsorted(src.Attributes(), dst.Attributes())
	}
	dst.SetDroppedAttributesCount(uint64(src.DroppedAttributesCount()))
	dst.SetFlags(uint64(src.Flags()))
	traceID := src.TraceID()
	spanID := src.SpanID()
	dst.SetTraceID(pkg.Bytes(traceID[:]))
	dst.SetSpanID(pkg.Bytes(spanID[:]))
	dst.SetEventName(src.EventName())
}
