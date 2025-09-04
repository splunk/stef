package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/examples/jsonl/internal/jsonstef"
)

// convertToJsonValue recursively converts a Go value (from json.Unmarshal) to a jsonstef.JsonValue.
func convertToJsonValue(src interface{}, dst *jsonstef.JsonValue) {
	switch v := src.(type) {
	case map[string]interface{}:
		dst.SetType(jsonstef.JsonValueTypeObject)
		obj := dst.Object()
		obj.EnsureLen(len(v))
		i := 0

		var keys []string
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			obj.SetKey(i, k)
			convertToJsonValue(v[k], obj.At(i).Value())
			i++
		}
		//obj.Sort()
	case []interface{}:
		dst.SetType(jsonstef.JsonValueTypeArray)
		arr := dst.Array()
		arr.EnsureLen(len(v))
		for i, subv := range v {
			convertToJsonValue(subv, arr.At(i))
		}
	case string:
		dst.SetString(v)
	case float64:
		dst.SetNumber(v)
	case bool:
		dst.SetBool(v)
	case nil:
		dst.SetType(jsonstef.JsonValueTypeNone)
	default:
		dst.SetType(jsonstef.JsonValueTypeNone)
	}
}

// convertJSONLToSTEF reads JSONL bytes, converts each line to a jsonstef.Record, and writes to a STEF stream.
func convertJSONLToSTEF(jsonlData []byte, writerOpts pkg.WriterOptions) ([]byte, error) {
	r := bufio.NewReaderSize(bytes.NewReader(jsonlData), 4096)
	buf := pkg.MemChunkWriter{}
	w, err := jsonstef.NewRecordWriter(&buf, writerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create STEF writer: %w", err)
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		var raw interface{}
		if err := json.Unmarshal(line, &raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		convertToJsonValue(raw, w.Record.Value())
		if err := w.Write(); err != nil {
			return nil, fmt.Errorf("failed to write record: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}
	if err := w.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush writer: %w", err)
	}
	return buf.Bytes(), nil
}
