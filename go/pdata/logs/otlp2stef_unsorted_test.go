package logs

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pkg"
)

func TestConvertLogRecord(t *testing.T) {
	src := plog.NewLogs()
	rl := src.ResourceLogs().AppendEmpty()
	rl.SetSchemaUrl("resource-schema")
	rl.Resource().Attributes().PutStr("service.name", "checkout")
	rl.Resource().SetDroppedAttributesCount(1)

	sl := rl.ScopeLogs().AppendEmpty()
	sl.SetSchemaUrl("scope-schema")
	sl.Scope().SetName("logger")
	sl.Scope().SetVersion("1.2.3")
	sl.Scope().Attributes().PutBool("scope-attribute", true)
	sl.Scope().SetDroppedAttributesCount(2)

	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.SetTimestamp(pcommon.Timestamp(100))
	logRecord.SetObservedTimestamp(pcommon.Timestamp(200))
	logRecord.SetSeverityNumber(plog.SeverityNumberError)
	logRecord.SetSeverityText("ERROR")
	body := logRecord.Body().SetEmptyMap()
	body.PutStr("message", "failed")
	body.PutInt("attempt", 3)
	logRecord.Attributes().PutStr("log.attribute", "value")
	logRecord.SetDroppedAttributesCount(3)
	logRecord.SetFlags(plog.DefaultLogRecordFlags.WithIsSampled(true))
	traceID := pcommon.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := pcommon.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	logRecord.SetTraceID(traceID)
	logRecord.SetSpanID(spanID)
	logRecord.SetEventName("exception")

	buf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewLogsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)
	require.NoError(t, (&OtlpToStefUnsorted{}).Convert(src, writer))
	require.NoError(t, writer.Flush())
	require.EqualValues(t, 1, writer.RecordCount())

	reader, err := otelstef.NewLogsReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	require.NoError(t, reader.Read(pkg.ReadOptions{}))

	require.Equal(t, "resource-schema", reader.Record.Resource().SchemaURL())
	require.EqualValues(t, 1, reader.Record.Resource().DroppedAttributesCount())
	require.Equal(t, "logger", reader.Record.Scope().Name())
	require.Equal(t, "1.2.3", reader.Record.Scope().Version())
	require.Equal(t, "scope-schema", reader.Record.Scope().SchemaURL())
	require.EqualValues(t, 2, reader.Record.Scope().DroppedAttributesCount())

	got := reader.Record.Log()
	require.EqualValues(t, 100, got.TimeUnixNano())
	require.EqualValues(t, 200, got.ObservedTimeUnixNano())
	require.Equal(t, otelstef.SeverityNumberError, got.SeverityNumber())
	require.Equal(t, "ERROR", got.SeverityText())
	require.Equal(t, otelstef.AnyValueTypeKVList, got.Body().Type())
	require.Equal(t, 2, got.Body().KVList().Len())
	require.Equal(t, 1, got.Attributes().Len())
	require.EqualValues(t, 3, got.DroppedAttributesCount())
	require.EqualValues(t, logRecord.Flags(), got.Flags())
	require.Equal(t, pkg.Bytes(traceID[:]), got.TraceID())
	require.Equal(t, pkg.Bytes(spanID[:]), got.SpanID())
	require.Equal(t, "exception", got.EventName())
	require.ErrorIs(t, reader.Read(pkg.ReadOptions{}), io.EOF)
}

func TestConvertSortedGroupsAndOrdersLogs(t *testing.T) {
	src := plog.NewLogs()
	for _, timestamp := range []pcommon.Timestamp{200, 100} {
		rl := src.ResourceLogs().AppendEmpty()
		rl.Resource().Attributes().PutStr("service.name", "checkout")
		sl := rl.ScopeLogs().AppendEmpty()
		sl.Scope().SetName("logger")
		logRecord := sl.LogRecords().AppendEmpty()
		logRecord.SetTimestamp(timestamp)
	}

	buf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewLogsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)
	require.NoError(t, NewOtlpToStef(true).Convert(src, writer))
	require.NoError(t, writer.Flush())
	require.EqualValues(t, 2, writer.RecordCount())

	reader, err := otelstef.NewLogsReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	for _, timestamp := range []uint64{100, 200} {
		require.NoError(t, reader.Read(pkg.ReadOptions{}))
		require.Equal(t, timestamp, reader.Record.Log().TimeUnixNano())
	}
	require.ErrorIs(t, reader.Read(pkg.ReadOptions{}), io.EOF)
}
