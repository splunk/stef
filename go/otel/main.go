package main

/*
var metricSchema = schema.AstSchema{
	PackageName: "otelstef",
	Multimaps: map[string]*schema.AstMapDef{
		"EnvelopeAttributes": {
			Name:  "EnvelopeAttributes",
			Key:   schema.AstMapFieldDef{Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString}},
			Value: schema.AstMapFieldDef{Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeBytes}},
		},
		"Attributes": {
			Name: "Attributes",
			Key: schema.AstMapFieldDef{
				Type: &schema.AstSimpleTypeRef{
					Type: schema.SimpleTypeString, Dict: "AttributeKey",
				},
			},
			Value: schema.AstMapFieldDef{
				Type: &schema.AstStructTypeRef{Name: "AnyValue"},
			},
		},
		"KeyValueList": {
			Name:  "KeyValueList",
			Key:   schema.AstMapFieldDef{Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString}},
			Value: schema.AstMapFieldDef{Type: &schema.AstStructTypeRef{Name: "AnyValue"}, Recursive: true},
		},
	},
	Structs: map[string]*schema.AstStructDef{
		"AnyValue": {
			Name: "AnyValue",
			Fields: []*schema.AstStructFieldDef{
				{Name: "String", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "AnyValueString"}},
				{Name: "Bool", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeBool}},
				{Name: "Int64", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeInt64}},
				{Name: "Float64", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64}},
				{
					Name: "Array", Recursive: true,
					Type: &schema.AstArrayTypeRef{ElemType: &schema.AstStructTypeRef{Name: "AnyValue"}},
				},
				{
					Name: "KVList", Recursive: true,
					Type: &schema.AstMultimapTypeRef{Name: "KeyValueList"},
				},
				{Name: "Bytes", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeBytes}},
			},
			OneOf: true,
		},
		"Envelope": {
			Name: "Envelope",
			Fields: []*schema.AstStructFieldDef{
				{
					Name: "Attributes", Type: &schema.AstMultimapTypeRef{Name: "EnvelopeAttributes"},
				},
			},
		},
		"Metric": {
			Name: "Metric",
			Dict: "Metric",
			Fields: []*schema.AstStructFieldDef{
				{Name: "Name", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "MetricName"}},
				{
					Name: "Description",
					Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "MetricDescription"},
				},
				{Name: "Unit", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "MetricUnit"}},
				{Name: "Type", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeUint64}},
				{
					Name: "Metadata", Type: &schema.AstMultimapTypeRef{Name: "Attributes"},
				},
				{
					Name: "HistogramBounds",
					Type: &schema.AstArrayTypeRef{ElemType: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64}},
				},
				//{Name: "Flags", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeUint64}},
				{Name: "AggregationTemporality", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeUint64}},
				{Name: "Monotonic", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeBool}},
			},
		},
		"Resource": {
			Name: "Resource",
			Dict: "Resource",
			Fields: []*schema.AstStructFieldDef{
				{Name: "SchemaURL", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "SchemaURL"}},
				{
					Name: "Attributes", Type: &schema.AstMultimapTypeRef{Name: "Attributes"},
				},
			},
		},
		"Scope": {
			Name: "Scope",
			Dict: "Scope",
			Fields: []*schema.AstStructFieldDef{
				{Name: "Name", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "ScopeName"}},
				{Name: "Version", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "ScopeVersion"}},
				{Name: "SchemaURL", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeString, Dict: "SchemaURL"}},
				{Name: "Attributes", Type: &schema.AstMultimapTypeRef{Name: "Attributes"}},
			},
		},
		"Point": {
			Name: "Point",
			Fields: []*schema.AstStructFieldDef{
				{Name: "StartTimestamp", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeUint64}},
				{Name: "Timestamp", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeUint64}},
				{
					Name: "Value",
					Type: &schema.AstStructTypeRef{Name: "PointValue"},
				},
				{
					Name: "Exemplars",
					Type: &schema.AstArrayTypeRef{ElemType: &schema.AstStructTypeRef{Name: "Exemplar"}},
				},
			},
		},
		"PointValue": {
			Name: "PointValue",
			Fields: []*schema.AstStructFieldDef{
				{
					Name: "Int64",
					Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeInt64},
				},
				{
					Name: "Float64",
					Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64},
				},
				{
					Name: "Histogram",
					Type: &schema.AstStructTypeRef{Name: "HistogramValue"},
				},
			},
			OneOf: true,
		},
		"HistogramValue": {
			Name: "HistogramValue",
			Fields: []*schema.AstStructFieldDef{
				{
					Name: "Count",
					Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeInt64},
				},
				{
					Name:     "Sum",
					Type:     &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64},
					Optional: true,
				},
				{
					Name:     "Min",
					Type:     &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64},
					Optional: true,
				},
				{
					Name:     "Max",
					Type:     &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64},
					Optional: true,
				},
				{
					Name: "BucketCounts",
					Type: &schema.AstArrayTypeRef{ElemType: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeInt64}},
				},
				//{
				//	Name: "Float64Vals",
				//	Type: &schema.AstArrayTypeRef{ElemType: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64}},
				//},
			},
		},
		"Exemplar": {
			Name: "Exemplar",
			Fields: []*schema.AstStructFieldDef{
				{Name: "Timestamp", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeUint64}},
				{Name: "Value", Type: &schema.AstStructTypeRef{Name: "ExemplarValue"}},
				{Name: "SpanID", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeBytes, Dict: "Span"}},
				{Name: "TraceID", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeBytes, Dict: "Trace"}},
				{
					Name: "FilteredAttributes", Type: &schema.AstMultimapTypeRef{Name: "Attributes"},
				},
			},
		},
		"ExemplarValue": {
			Name: "ExemplarValue",
			Fields: []*schema.AstStructFieldDef{
				{Name: "Int64", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeInt64}},
				{Name: "Float64", Type: &schema.AstSimpleTypeRef{Type: schema.SimpleTypeFloat64}},
			},
			OneOf: true,
		},
		"Record": {
			Name: "Record",
			Fields: []*schema.AstStructFieldDef{
				{Name: "Envelope", Type: &schema.AstStructTypeRef{Name: "Envelope"}},
				{Name: "Metric", Type: &schema.AstStructTypeRef{Name: "Metric"}},
				{Name: "Resource", Type: &schema.AstStructTypeRef{Name: "Resource"}},
				{Name: "Scope", Type: &schema.AstStructTypeRef{Name: "Scope"}},
				{Name: "Attributes", Type: &schema.AstMultimapTypeRef{Name: "Attributes"}},
				{Name: "Point", Type: &schema.AstStructTypeRef{Name: "Point"}},
				//{Name: "AnyValue", Type: &schema.AstStructTypeRef{Name: "AnyValue"}},
			},
		},
	},
	MainStruct: "Record",
}


func writeJSONSchema() *schema.Schema {
	//jsonSchema, err := json.MarshalIndent(metricSchema, "", "  ")
	////jsonSchema, err := json.Marshal(metricSchema)
	//if err != nil {
	//	panic(err)
	//}
	//err = os.WriteFile(metricSchema.PackageName+".schema.json", jsonSchema, 0755)
	//if err != nil {
	//	panic(err)
	//}

	wireSchema := schema.AstToWire(&metricSchema)
	wireJson, err := json.MarshalIndent(wireSchema, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(metricSchema.PackageName+".wire.json", wireJson, 0755)
	if err != nil {
		panic(err)
	}

	var schemaCopy schema.Schema
	err = json.Unmarshal(wireJson, &schemaCopy)
	if err != nil {
		panic(err)
	}

	wireSchema.Minify()
	wireJson, err = json.MarshalIndent(wireSchema, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(metricSchema.PackageName+".wire.min.json", wireJson, 0755)
	if err != nil {
		panic(err)
	}

	wireJson, err = json.Marshal(wireSchema)
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"JSNO descriptor size is %d bytes uncompressed, %d bytes compressed\n",
		len(wireJson),
		len(compressZstd(wireJson)),
	)

	return &schemaCopy
}
*/

func main() {
}
