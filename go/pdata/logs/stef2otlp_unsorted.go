package logs

import (
	"errors"
	"io"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
	"github.com/splunk/stef/go/pkg"
)

// StefToOtlpUnsorted reads STEF records and converts them to OTLP logs.
type StefToOtlpUnsorted struct{}

var _ StefToOtlp = (*StefToOtlpUnsorted)(nil)

// Convert reads until EOF when untilEOF is true. Otherwise, it reads at least
// one record and stops at the end of the current frame.
func (c *StefToOtlpUnsorted) Convert(reader *otelstef.LogsReader, untilEOF bool) (plog.Logs, error) {
	logs := plog.NewLogs()

	if err := reader.Read(pkg.ReadOptions{}); err != nil {
		return logs, err
	}

	var resourceLogs plog.ResourceLogs
	var scopeLogs plog.ScopeLogs
	record := &reader.Record
	modified := true

	for {
		if modified || record.IsResourceModified() {
			modified = true
			resourceLogs = logs.ResourceLogs().AppendEmpty()
			if err := resourceToOtlp(record.Resource(), resourceLogs); err != nil {
				return logs, err
			}
		}

		if modified || record.IsScopeModified() {
			modified = true
			scopeLogs = resourceLogs.ScopeLogs().AppendEmpty()
			if err := scopeToOtlp(record.Scope(), scopeLogs); err != nil {
				return logs, err
			}
		}

		if err := logRecordToOtlp(record.Log(), scopeLogs.LogRecords().AppendEmpty()); err != nil {
			return logs, err
		}
		modified = false

		var err error
		if untilEOF {
			err = reader.Read(pkg.ReadOptions{})
			if errors.Is(err, io.EOF) {
				break
			}
		} else {
			err = reader.Read(pkg.ReadOptions{TillEndOfFrame: true})
			if errors.Is(err, pkg.ErrEndOfFrame) {
				break
			}
		}
		if err != nil {
			return logs, err
		}
	}

	return logs, nil
}

func resourceToOtlp(src *otelstef.Resource, dst plog.ResourceLogs) error {
	if src == nil {
		return nil
	}
	dst.SetSchemaUrl(src.SchemaURL())
	dst.Resource().SetDroppedAttributesCount(uint32(src.DroppedAttributesCount()))
	return otlptools.TefToOtlpMap(src.Attributes(), dst.Resource().Attributes())
}

func scopeToOtlp(src *otelstef.Scope, dst plog.ScopeLogs) error {
	if src == nil {
		return nil
	}
	dst.SetSchemaUrl(src.SchemaURL())
	dst.Scope().SetName(src.Name())
	dst.Scope().SetVersion(src.Version())
	dst.Scope().SetDroppedAttributesCount(uint32(src.DroppedAttributesCount()))
	return otlptools.TefToOtlpMap(src.Attributes(), dst.Scope().Attributes())
}

func logRecordToOtlp(src *otelstef.LogRecord, dst plog.LogRecord) error {
	dst.SetTimestamp(pcommon.Timestamp(src.TimeUnixNano()))
	dst.SetObservedTimestamp(pcommon.Timestamp(src.ObservedTimeUnixNano()))
	dst.SetSeverityNumber(plog.SeverityNumber(src.SeverityNumber()))
	dst.SetSeverityText(src.SeverityText())
	if err := otlptools.TefAnyValueToOtlp(src.Body(), dst.Body()); err != nil {
		return err
	}
	if err := otlptools.TefToOtlpMap(src.Attributes(), dst.Attributes()); err != nil {
		return err
	}
	dst.SetDroppedAttributesCount(uint32(src.DroppedAttributesCount()))
	dst.SetFlags(plog.LogRecordFlags(src.Flags()))
	var traceID pcommon.TraceID
	copy(traceID[:], src.TraceID())
	dst.SetTraceID(traceID)
	var spanID pcommon.SpanID
	copy(spanID[:], src.SpanID())
	dst.SetSpanID(spanID)
	dst.SetEventName(src.EventName())
	return nil
}
