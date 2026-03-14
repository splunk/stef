use rand::Rng;

use crate::schema::{Schema, Struct};

/// Deterministically shrinks one random field from schema, preserving root minimum field.
pub fn shrink_randomly<R: Rng + ?Sized>(r: &mut R, schema: &mut Schema) -> bool {
    let mut total_field_count = 0usize;
    let mut struct_names: Vec<String> = schema.structs.keys().cloned().collect();
    struct_names.sort();
    for name in &struct_names {
        let s = &schema.structs[name];
        let mut shrinkable = s.fields.len();
        if s.is_root && shrinkable > 0 {
            shrinkable -= 1;
        }
        total_field_count += shrinkable;
    }
    if total_field_count == 0 {
        return false;
    }

    loop {
        let idx = r.random_range(0..struct_names.len());
        let name = struct_names[idx].clone();
        if shrink_struct(r, schema, &name) {
            return true;
        }
    }
}

fn shrink_struct<R: Rng + ?Sized>(r: &mut R, schema: &mut Schema, struct_name: &str) -> bool {
    let is_root = schema.structs.get(struct_name).map(|s| s.is_root).unwrap_or(false);
    let len = schema.structs.get(struct_name).map(|s| s.fields.len()).unwrap_or(0);
    if is_root && len <= 1 {
        return false;
    }

    if r.random_range(0..10) == 0 && len > 0 {
        if let Some(s) = schema.structs.get_mut(struct_name) {
            s.fields.pop();
            return true;
        }
        return false;
    }

    let child_structs: Vec<String> = schema
        .structs
        .get(struct_name)
        .map(|s| {
            s.fields
                .iter()
                .filter(|f| !f.field_type.r#struct.is_empty())
                .map(|f| f.field_type.r#struct.clone())
                .collect()
        })
        .unwrap_or_default();

    for child in child_structs {
        if r.random_range(0..3) == 0 && shrink_struct(r, schema, &child) {
            return true;
        }
    }
    false
}
