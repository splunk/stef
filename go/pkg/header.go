package pkg

import (
	"encoding/json"
)

type FixedHeader struct {
	Compression         Compression
	TimestampMultiplier uint64
}

type VarHeader struct {
	Schema   *json.RawMessage `json:"schema"`
	UserData map[string]any   `json:"user,omitempty"`
}
