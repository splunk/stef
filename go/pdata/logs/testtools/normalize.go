package testtools

import (
	"slices"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/splunk/stef/go/pdata/internal/otlptools"
)

// NormalizeLogs canonicalizes grouping, record order, and map order so that
// semantically equivalent logs have identical protobuf encodings.
func NormalizeLogs(data plog.Logs) {
	resourceLogs := data.ResourceLogs()
	for i := 0; i < resourceLogs.Len(); i++ {
		normalizeMap(resourceLogs.At(i).Resource().Attributes())
	}
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
	for i := 0; i < resourceLogs.Len(); i++ {
		normalizeScopeLogs(resourceLogs.At(i).ScopeLogs())
	}
}

func normalizeScopeLogs(scopeLogs plog.ScopeLogsSlice) {
	for i := 0; i < scopeLogs.Len(); i++ {
		normalizeMap(scopeLogs.At(i).Scope().Attributes())
	}
	scopeLogs.Sort(func(a, b plog.ScopeLogs) bool {
		return otlptools.CmpScopeLogs(a, b) < 0
	})
	for i := 0; i < scopeLogs.Len()-1; {
		if otlptools.CmpScopeLogs(scopeLogs.At(i), scopeLogs.At(i+1)) != 0 {
			i++
			continue
		}
		scopeLogs.At(i + 1).LogRecords().MoveAndAppendTo(scopeLogs.At(i).LogRecords())
		removeScopeLogsAt(scopeLogs, i+1)
	}
	for i := 0; i < scopeLogs.Len(); i++ {
		records := scopeLogs.At(i).LogRecords()
		for j := 0; j < records.Len(); j++ {
			normalizeValue(records.At(j).Body())
			normalizeMap(records.At(j).Attributes())
		}
		records.Sort(func(a, b plog.LogRecord) bool {
			return otlptools.CmpLogRecords(a, b) < 0
		})
	}
}

func normalizeMap(m pcommon.Map) {
	type entry struct {
		key   string
		value pcommon.Value
	}
	entries := make([]entry, 0, m.Len())
	m.Range(func(key string, value pcommon.Value) bool {
		normalizeValue(value)
		valueCopy := pcommon.NewValueEmpty()
		value.CopyTo(valueCopy)
		entries = append(entries, entry{key: key, value: valueCopy})
		return true
	})
	slices.SortFunc(entries, func(a, b entry) int { return strings.Compare(a.key, b.key) })
	m.Clear()
	m.EnsureCapacity(len(entries))
	for _, entry := range entries {
		entry.value.CopyTo(m.PutEmpty(entry.key))
	}
}

func normalizeValue(value pcommon.Value) {
	switch value.Type() {
	case pcommon.ValueTypeMap:
		normalizeMap(value.Map())
	case pcommon.ValueTypeSlice:
		values := value.Slice()
		for i := 0; i < values.Len(); i++ {
			normalizeValue(values.At(i))
		}
	}
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
