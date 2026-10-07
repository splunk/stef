package sortedbyresource

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

func TestOtlpToSortedTreeGroupsAndSortsLogs(t *testing.T) {
	src := plog.NewLogs()
	for _, timestamp := range []pcommon.Timestamp{200, 100} {
		resourceLogs := src.ResourceLogs().AppendEmpty()
		resourceLogs.Resource().Attributes().PutStr("service.name", "checkout")
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		scopeLogs.Scope().SetName("logger")
		scopeLogs.LogRecords().AppendEmpty().SetTimestamp(timestamp)
	}

	tree := OtlpToSortedTree(src)
	output := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewLogsWriter(output, pkg.WriterOptions{})
	require.NoError(t, err)
	require.NoError(t, tree.ToStef(writer))
	require.NoError(t, writer.Flush())

	reader, err := otelstef.NewLogsReader(bytes.NewReader(output.Bytes()))
	require.NoError(t, err)
	for i, timestamp := range []uint64{100, 200} {
		require.NoError(t, reader.Read(pkg.ReadOptions{}))
		require.Equal(t, timestamp, reader.Record.Log().TimeUnixNano())
		if i > 0 {
			require.False(t, reader.Record.IsResourceModified())
			require.False(t, reader.Record.IsScopeModified())
		}
	}
	require.ErrorIs(t, reader.Read(pkg.ReadOptions{}), io.EOF)
}
