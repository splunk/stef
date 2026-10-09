package otlptools

import (
	"bytes"
	"cmp"
	"slices"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type AttrAccessible struct {
	Key   string
	Value pcommon.Value
}

func Map2attrs(attributes pcommon.Map) []AttrAccessible {
	attrs := make([]AttrAccessible, 0, attributes.Len())
	attributes.Range(
		func(k string, v pcommon.Value) bool {
			attrs = append(attrs, AttrAccessible{Key: k, Value: v})
			return true
		},
	)
	slices.SortFunc(attrs, func(a, b AttrAccessible) int {
		return strings.Compare(a.Key, b.Key)
	})
	return attrs
}

func CmpResourceMetrics(left, right pmetric.ResourceMetrics) int {
	c := strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return c
	}
	c = cmp.Compare(left.Resource().DroppedAttributesCount(), right.Resource().DroppedAttributesCount())
	if c != 0 {
		return c
	}
	return CmpAttrs(left.Resource().Attributes(), right.Resource().Attributes())
}

func CmpResourceSpans(left, right ptrace.ResourceSpans) int {
	c := strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return c
	}
	c = cmp.Compare(left.Resource().DroppedAttributesCount(), right.Resource().DroppedAttributesCount())
	if c != 0 {
		return c
	}
	return CmpAttrs(left.Resource().Attributes(), right.Resource().Attributes())
}

func CmpResourceLogs(left, right plog.ResourceLogs) int {
	c := strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return c
	}
	c = cmp.Compare(left.Resource().DroppedAttributesCount(), right.Resource().DroppedAttributesCount())
	if c != 0 {
		return c
	}
	return CmpAttrs(left.Resource().Attributes(), right.Resource().Attributes())
}

func CmpScopeMetrics(left, right pmetric.ScopeMetrics) int {
	c := strings.Compare(left.Scope().Name(), right.Scope().Name())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.Scope().Version(), right.Scope().Version())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return c
	}
	c = cmp.Compare(left.Scope().DroppedAttributesCount(), right.Scope().DroppedAttributesCount())
	if c != 0 {
		return c
	}

	return CmpAttrs(left.Scope().Attributes(), right.Scope().Attributes())
}

func CmpScopeSpans(left, right ptrace.ScopeSpans) int {
	c := strings.Compare(left.Scope().Name(), right.Scope().Name())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.Scope().Version(), right.Scope().Version())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return c
	}
	c = cmp.Compare(left.Scope().DroppedAttributesCount(), right.Scope().DroppedAttributesCount())
	if c != 0 {
		return c
	}

	return CmpAttrs(left.Scope().Attributes(), right.Scope().Attributes())
}

func CmpScopeLogs(left, right plog.ScopeLogs) int {
	c := strings.Compare(left.Scope().Name(), right.Scope().Name())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.Scope().Version(), right.Scope().Version())
	if c != 0 {
		return c
	}
	c = strings.Compare(left.SchemaUrl(), right.SchemaUrl())
	if c != 0 {
		return c
	}
	c = cmp.Compare(left.Scope().DroppedAttributesCount(), right.Scope().DroppedAttributesCount())
	if c != 0 {
		return c
	}
	return CmpAttrs(left.Scope().Attributes(), right.Scope().Attributes())
}

// CmpLogRecords compares log records by every field encoded by the STEF logs
// converter. It provides a deterministic order for otherwise identical
// resource and scope groupings.
func CmpLogRecords(left, right plog.LogRecord) int {
	if c := cmp.Compare(left.Timestamp(), right.Timestamp()); c != 0 {
		return c
	}
	if c := cmp.Compare(left.ObservedTimestamp(), right.ObservedTimestamp()); c != 0 {
		return c
	}
	leftTraceID, rightTraceID := left.TraceID(), right.TraceID()
	if c := bytes.Compare(leftTraceID[:], rightTraceID[:]); c != 0 {
		return c
	}
	leftSpanID, rightSpanID := left.SpanID(), right.SpanID()
	if c := bytes.Compare(leftSpanID[:], rightSpanID[:]); c != 0 {
		return c
	}
	if c := cmp.Compare(left.SeverityNumber(), right.SeverityNumber()); c != 0 {
		return c
	}
	if c := strings.Compare(left.SeverityText(), right.SeverityText()); c != 0 {
		return c
	}
	if c := CmpVal(left.Body(), right.Body()); c != 0 {
		return c
	}
	if c := CmpAttrs(left.Attributes(), right.Attributes()); c != 0 {
		return c
	}
	if c := cmp.Compare(left.DroppedAttributesCount(), right.DroppedAttributesCount()); c != 0 {
		return c
	}
	if c := cmp.Compare(left.Flags(), right.Flags()); c != 0 {
		return c
	}
	return strings.Compare(left.EventName(), right.EventName())
}

func CmpAttrs(a, b pcommon.Map) int {
	left := Map2attrs(a)
	right := Map2attrs(b)

	lenDiff := len(left) - len(right)
	l := min(len(left), len(right))
	for i := 0; i < l; i++ {
		c := strings.Compare(left[i].Key, right[i].Key)
		if c != 0 {
			return c
		}
	}
	if lenDiff != 0 {
		return lenDiff
	}
	for i := 0; i < l; i++ {
		c := CmpVal(left[i].Value, right[i].Value)
		if c != 0 {
			return c
		}
	}
	return lenDiff
}

func CmpVal(left, right pcommon.Value) int {
	c := int(left.Type()) - int(right.Type())
	if c != 0 {
		return c
	}

	switch left.Type() {
	case pcommon.ValueTypeStr:
		return strings.Compare(left.Str(), right.Str())

	case pcommon.ValueTypeInt:
		return CmpInt64(left.Int(), right.Int())

	case pcommon.ValueTypeBool:
		return CmpBool(left.Bool(), right.Bool())

	case pcommon.ValueTypeDouble:
		return cmp.Compare(left.Double(), right.Double())

	case pcommon.ValueTypeBytes:
		return bytes.Compare(left.Bytes().AsRaw(), right.Bytes().AsRaw())

	case pcommon.ValueTypeSlice:
		left := left.Slice()
		right := right.Slice()
		if left.Len() != right.Len() {
			return left.Len() - right.Len()
		}
		for i := 0; i < left.Len(); i++ {
			c := CmpVal(left.At(i), right.At(i))
			if c != 0 {
				return c
			}
		}
		return 0

	case pcommon.ValueTypeEmpty:
		return 0

	case pcommon.ValueTypeMap:
		return CmpAttrs(left.Map(), right.Map())

	default:
		panic("unknown value type")
	}
}

func CmpBool(v1, v2 bool) int {
	if v1 == v2 {
		return 0
	}
	if v1 {
		return 1
	}
	return -1
}

func CmpInt64(v1, v2 int64) int {
	if v1 < v2 {
		return -1
	} else if v1 > v2 {
		return 1
	}
	return 0
}
