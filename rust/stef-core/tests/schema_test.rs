use stef_core::schema::{Compatibility, FieldType, PrimitiveFieldType, PrimitiveType, Schema, Struct, StructField, WireSchema};

#[test]
fn test_schema_self_compatible() {
    let mut schema = Schema::default();
    schema.package_name = vec!["pkg".into()];
    schema.structs.insert("Root".into(), Struct { name: "Root".into(), ..Default::default() });
    schema.resolve_refs().unwrap();

    let wire = WireSchema::new(&schema, "Root");
    assert_eq!(wire.compatible(&wire).unwrap(), Compatibility::Exact);
}

#[test]
fn test_schema_superset() {
    let p = PrimitiveType { r#type: PrimitiveFieldType::Int64 };
    let mut old = Schema::default();
    old.structs.insert(
        "Root".into(),
        Struct {
            name: "Root".into(),
            fields: vec![StructField { name: "F1".into(), field_type: FieldType { primitive: Some(p.clone()), ..Default::default() }, optional: false }],
            ..Default::default()
        },
    );

    let mut new = old.clone();
    new.structs.get_mut("Root").unwrap().fields.push(StructField { name: "F2".into(), field_type: FieldType { primitive: Some(p), ..Default::default() }, optional: false });

    old.resolve_refs().unwrap();
    new.resolve_refs().unwrap();

    let old_wire = WireSchema::new(&old, "Root");
    let new_wire = WireSchema::new(&new, "Root");
    assert_eq!(new_wire.compatible(&old_wire).unwrap(), Compatibility::Superset);
}

#[test]
fn test_pretty_print_simple_struct() {
    let mut schema = Schema::default();
    schema.package_name = vec!["com".into(), "example".into(), "test".into()];
    schema.structs.insert(
        "Person".into(),
        Struct {
            name: "Person".into(),
            fields: vec![
                StructField { name: "Name".into(), field_type: FieldType { primitive: Some(PrimitiveType { r#type: PrimitiveFieldType::String }), ..Default::default() }, optional: false },
                StructField { name: "Age".into(), field_type: FieldType { primitive: Some(PrimitiveType { r#type: PrimitiveFieldType::Uint64 }), ..Default::default() }, optional: false },
            ],
            ..Default::default()
        },
    );

    let expected = "package com.example.test\n\nstruct Person {\n  Name string\n  Age uint64\n}";
    assert_eq!(schema.pretty_print().trim(), expected);
}
