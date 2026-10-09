package testtools

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestNormalizeLogsCanonicalizesLogRecordOrder(t *testing.T) {
	tests := map[string]func(plog.LogRecord){
		"severity number": func(record plog.LogRecord) {
			record.SetSeverityNumber(plog.SeverityNumberError)
		},
		"severity text": func(record plog.LogRecord) {
			record.SetSeverityText("error")
		},
		"body": func(record plog.LogRecord) {
			record.Body().SetStr("body")
		},
		"attributes": func(record plog.LogRecord) {
			record.Attributes().PutStr("key", "value")
		},
		"dropped attribute count": func(record plog.LogRecord) {
			record.SetDroppedAttributesCount(1)
		},
		"flags": func(record plog.LogRecord) {
			record.SetFlags(plog.DefaultLogRecordFlags.WithIsSampled(true))
		},
		"event name": func(record plog.LogRecord) {
			record.SetEventName("event")
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			forward := newLogsWithRecordOrder(false, mutate)
			reversed := newLogsWithRecordOrder(true, mutate)

			NormalizeLogs(forward)
			NormalizeLogs(reversed)

			marshaler := plog.ProtoMarshaler{}
			forwardBytes, err := marshaler.MarshalLogs(forward)
			require.NoError(t, err)
			reversedBytes, err := marshaler.MarshalLogs(reversed)
			require.NoError(t, err)
			require.Equal(t, forwardBytes, reversedBytes)
		})
	}
}

func newLogsWithRecordOrder(reversed bool, mutate func(plog.LogRecord)) plog.Logs {
	logs := plog.NewLogs()
	records := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords()
	if reversed {
		mutate(records.AppendEmpty())
		records.AppendEmpty()
	} else {
		records.AppendEmpty()
		mutate(records.AppendEmpty())
	}
	return logs
}
