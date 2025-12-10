package otlptools

import (
	"errors"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/splunk/stef/go/otel/otelstef"
)

//func TefToOtlpMap(in *otelstef.Attributes, out pcommon.Map) error {
//	out.EnsureCapacity(in.Len())
//
//	decoder := anyvalue.Decoder{}
//	for i := 0; i < in.Len(); i++ {
//		kv := in.At(i)
//		val := out.PutEmpty(kv.Key())
//		decoder.Reset(anyvalue.ImmutableBytes(kv.Value().Bytes()))
//		err := tefToOtlpValue(&decoder, val)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

func TefToOtlpMap(in *otelstef.Attributes, out pcommon.Map) error {
	out.EnsureCapacity(in.Len())

	//decoder := anyvalue.Decoder{}
	for i := 0; i < in.Len(); i++ {
		val := out.PutEmpty(in.Key(i))
		err := tefAnyValueToOtlp(in.Value(i), val)
		if err != nil {
			return err
		}
	}
	return nil
}

var errDecode = errors.New("decode error")

func tefAnyValueToOtlp(anyVal *otelstef.AnyValue, into pcommon.Value) error {
	switch anyVal.Type() {
	case otelstef.AnyValueTypeString:
		into.SetStr(anyVal.String())

	case otelstef.AnyValueTypeBytes:
		bytes := into.SetEmptyBytes()
		bytes.Append([]byte(anyVal.Bytes())...)

	case otelstef.AnyValueTypeInt64:
		into.SetInt(anyVal.Int64())

	case otelstef.AnyValueTypeBool:
		into.SetBool(anyVal.Bool())

	case otelstef.AnyValueTypeNone:

	case otelstef.AnyValueTypeFloat64:
		into.SetDouble(anyVal.Float64())

	case otelstef.AnyValueTypeArray:
		values := into.SetEmptySlice()
		arr := anyVal.Array()
		for i := 0; i < arr.Len(); i++ {
			val := values.AppendEmpty()
			err := tefAnyValueToOtlp(arr.At(i), val)
			if err != nil {
				return err
			}
		}

	case otelstef.AnyValueTypeKVList:
		values := into.SetEmptyMap()
		kvList := anyVal.KVList()
		for i := 0; i < kvList.Len(); i++ {
			val := values.PutEmpty(kvList.Key(i))
			err := tefAnyValueToOtlp(kvList.Value(i), val)
			if err != nil {
				return err
			}
		}

	default:
		return errDecode
	}
	return nil
}
