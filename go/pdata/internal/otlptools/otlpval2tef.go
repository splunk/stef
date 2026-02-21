package otlptools

import (
	"slices"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/otelstef"
)

//func Map(m pcommon.Map, encoder *anyvalue.Encoder, out *otelstef.Attributes) {
//	out.EnsureLen(m.Len())
//	i := 0
//	m.Range(
//		func(k string, v pcommon.Value) bool {
//			otlpValueToTef(v, encoder)
//			out.SetKey(i, k)
//			out.At(i).Value().SetBytes(pkg.Bytes(encoder.ConsumeBytes()))
//			i++
//			return true
//		},
//	)
//	out.Sort()
//}

type elem struct {
	str string
	val pcommon.Value
}

type Otlp2Stef struct {
	attrElems []elem
	SortAttrs bool
}

func (o *Otlp2Stef) ResourceSorted(dst *otelstef.Resource, src pcommon.Resource, schemaUrl string) {
	dst.SetSchemaURL(schemaUrl)
	o.MapSorted(src.Attributes(), dst.Attributes())
	dst.SetDroppedAttributesCount(uint64(src.DroppedAttributesCount()))
}

func (o *Otlp2Stef) ResourceUnsorted(dst *otelstef.Resource, src pcommon.Resource, schemaUrl string) {
	dst.SetSchemaURL(schemaUrl)
	o.MapUnsorted(src.Attributes(), dst.Attributes())
	dst.SetDroppedAttributesCount(uint64(src.DroppedAttributesCount()))
}

func (o *Otlp2Stef) ScopeSorted(dst *otelstef.Scope, src pcommon.InstrumentationScope, schemaUrl string) {
	dst.SetSchemaURL(schemaUrl)
	dst.SetName(src.Name())
	dst.SetVersion(src.Version())
	o.MapSorted(src.Attributes(), dst.Attributes())
	dst.SetDroppedAttributesCount(uint64(src.DroppedAttributesCount()))
}

func (o *Otlp2Stef) ScopeUnsorted(dst *otelstef.Scope, src pcommon.InstrumentationScope, schemaUrl string) {
	dst.SetSchemaURL(schemaUrl)
	dst.SetName(src.Name())
	dst.SetVersion(src.Version())
	o.MapUnsorted(src.Attributes(), dst.Attributes())
	dst.SetDroppedAttributesCount(uint64(src.DroppedAttributesCount()))
}

func (o *Otlp2Stef) MapSorted(m pcommon.Map, out *otelstef.Attributes) {
	o.attrElems = pkg.EnsureLen(o.attrElems, m.Len())
	i := 0
	m.Range(
		func(k string, v pcommon.Value) bool {
			o.attrElems[i] = elem{str: k, val: v}
			i++
			return true
		},
	)

	slices.SortFunc(
		o.attrElems, func(a, b elem) int {
			return strings.Compare(a.str, b.str)
		},
	)

	out.EnsureLen(m.Len())
	for i := range o.attrElems {
		otlpValueToTefAnyValue(o.attrElems[i].val, out.Value(i))
		out.SetKey(i, o.attrElems[i].str)
	}
}

func (o *Otlp2Stef) MapUnsorted(m pcommon.Map, out *otelstef.Attributes) {
	if o.SortAttrs {
		o.MapSorted(m, out)
		return
	}
	out.EnsureLen(m.Len())
	i := 0
	m.Range(
		func(k string, v pcommon.Value) bool {
			otlpValueToTefAnyValue(v, out.Value(i))
			out.SetKey(i, k)
			i++
			return true
		},
	)
}

func otlpValueToTefAnyValue(val pcommon.Value, into *otelstef.AnyValue) {
	switch val.Type() {
	case pcommon.ValueTypeEmpty:
		into.SetType(otelstef.AnyValueTypeNone)

	case pcommon.ValueTypeStr:
		into.SetString(val.Str())

	case pcommon.ValueTypeBool:
		into.SetBool(val.Bool())

	case pcommon.ValueTypeDouble:
		into.SetFloat64(val.Double())

	case pcommon.ValueTypeInt:
		into.SetInt64(val.Int())

	case pcommon.ValueTypeBytes:
		into.SetBytes(pkg.Bytes(val.Bytes().AsRaw()))

	case pcommon.ValueTypeSlice:
		into.SetType(otelstef.AnyValueTypeArray)
		arr := into.Array()
		arr.EnsureLen(val.Slice().Len())

		for i := 0; i < val.Slice().Len(); i++ {
			otlpValueToTefAnyValue(val.Slice().At(i), arr.At(i))
		}

	case pcommon.ValueTypeMap:
		into.SetType(otelstef.AnyValueTypeKVList)
		kvList := into.KVList()
		kvList.EnsureLen(val.Map().Len())

		i := 0
		val.Map().Range(
			func(k string, v pcommon.Value) bool {
				kvList.SetKey(i, k)
				otlpValueToTefAnyValue(v, kvList.Value(i))
				return true
			},
		)

	default:
		panic("unknown anyValue type")
	}
}
