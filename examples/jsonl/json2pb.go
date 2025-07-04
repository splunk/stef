package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/splunk/stef/examples/jsonl/internal/jsonpb"
)

// convertToProtoJsonValue recursively converts a Go value (from json.Unmarshal) to a jsonpb.JsonValue (protobuf).
func convertToProtoJsonValue(src interface{}) *jsonpb.JsonValue {
	switch v := src.(type) {
	case map[string]interface{}:
		obj := &jsonpb.JsonObject{}
		for k, subv := range v {
			obj.Elems = append(
				obj.Elems, &jsonpb.JsonObjectElem{
					Key:   k,
					Value: convertToProtoJsonValue(subv),
				},
			)
		}
		return &jsonpb.JsonValue{Kind: &jsonpb.JsonValue_Object{Object: obj}}
	case []interface{}:
		arr := &jsonpb.JsonArray{}
		for _, subv := range v {
			arr.Values = append(arr.Values, convertToProtoJsonValue(subv))
		}
		return &jsonpb.JsonValue{Kind: &jsonpb.JsonValue_Array{Array: arr}}
	case string:
		return &jsonpb.JsonValue{Kind: &jsonpb.JsonValue_String_{String_: v}}
	case float64:
		return &jsonpb.JsonValue{Kind: &jsonpb.JsonValue_Number{Number: v}}
	case bool:
		return &jsonpb.JsonValue{Kind: &jsonpb.JsonValue_Bool{Bool: v}}
	case nil:
		return &jsonpb.JsonValue{} // nil/empty oneof = null
	default:
		return &jsonpb.JsonValue{}
	}
}

// convertJSONLToProto reads JSONL bytes, converts each line to a jsonpb.Record (protobuf), and returns the concatenated proto binary.
func convertJSONLToProto(jsonlData []byte) ([]byte, error) {
	r := bufio.NewReaderSize(bytes.NewReader(jsonlData), 4096)
	var out bytes.Buffer
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		var raw interface{}
		if err := json.Unmarshal(line, &raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		rec := &jsonpb.Record{Value: convertToProtoJsonValue(raw)}
		bin, err := proto.Marshal(rec)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proto: %w", err)
		}
		// Write size prefix (varint)
		var szbuf [10]byte
		n := binary.PutUvarint(szbuf[:], uint64(len(bin)))
		out.Write(szbuf[:n])
		out.Write(bin)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}
	return out.Bytes(), nil
}
