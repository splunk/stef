package schema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettyPrint_SimpleStruct(t *testing.T) {
	schema := &Schema{
		PackageName: []string{"com", "example", "test"},
		Structs: map[string]*Struct{
			"Person": {
				Name: "Person",
				Fields: []*StructField{
					{Name: "Name", FieldType: FieldType{Primitive: &PrimitiveType{Type: PrimitiveTypeString}}},
					{Name: "Age", FieldType: FieldType{Primitive: &PrimitiveType{Type: PrimitiveTypeUint64}}},
				},
			},
		},
	}

	expected := `package com.example.test

struct Person {
  Name string
  Age uint64
}`
	actual := strings.TrimSpace(schema.PrettyPrint())
	require.Equal(t, expected, actual)
}

func TestPrettyPrint_EnumAndMultimap(t *testing.T) {
	schema := &Schema{
		PackageName: []string{"com", "example", "test"},
		Enums: map[string]*Enum{
			"MetricType": {
				Name: "MetricType",
				Fields: []EnumField{
					{Name: "Gauge", Value: 0},
					{Name: "Counter", Value: 1},
				},
			},
		},
		Multimaps: map[string]*Multimap{
			"Labels": {
				Name: "Labels",
				Key: MultimapField{
					Type: FieldType{
						Primitive: &PrimitiveType{Type: PrimitiveTypeString}, DictName: "LabelKeys",
					},
				},
				Value: MultimapField{
					Type: FieldType{
						Primitive: &PrimitiveType{Type: PrimitiveTypeString}, DictName: "LabelValues",
					},
				},
			},
		},
	}

	expected := `package com.example.test

enum MetricType {
  Gauge = 0
  Counter = 1
}

multimap Labels {
  key string dict(LabelKeys)
  value string dict(LabelValues)
}`
	actual := strings.TrimSpace(schema.PrettyPrint())
	require.Equal(t, expected, actual)
}

func TestPrettyPrint_OneofAndArray(t *testing.T) {
	schema := &Schema{
		PackageName: []string{"example"},
		Structs: map[string]*Struct{
			"JsonValue": {
				Name:  "JsonValue",
				OneOf: true,
				Fields: []*StructField{
					{Name: "String", FieldType: FieldType{Primitive: &PrimitiveType{Type: PrimitiveTypeString}}},
					{Name: "Number", FieldType: FieldType{Primitive: &PrimitiveType{Type: PrimitiveTypeFloat64}}},
					{Name: "Array", FieldType: FieldType{Array: &ArrayType{ElemType: FieldType{Struct: "JsonValue"}}}},
				},
			},
		},
	}
	expected := `package example

oneof JsonValue {
  String string
  Number float64
  Array []JsonValue
}`
	actual := strings.TrimSpace(schema.PrettyPrint())
	require.Equal(t, expected, actual)
}

func TestPrettyPrint_OptionalAndDict(t *testing.T) {
	schema := &Schema{
		PackageName: []string{"example"},
		Structs: map[string]*Struct{
			"User": {
				Name: "User",
				Fields: []*StructField{
					{Name: "Name", FieldType: FieldType{Primitive: &PrimitiveType{Type: PrimitiveTypeString}}},
					{
						Name:      "Email",
						FieldType: FieldType{Primitive: &PrimitiveType{Type: PrimitiveTypeString}, DictName: "Emails"},
						Optional:  true,
					},
				},
			},
		},
	}
	expected := `package example

struct User {
  Name string
  Email string dict(Emails) optional
}`
	actual := strings.TrimSpace(schema.PrettyPrint())
	require.Equal(t, expected, actual)
}
