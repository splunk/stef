use crate::schema::FieldType;

/// Intermediate DFS tree used to linearize wire schema struct counts.
#[derive(Debug, Clone, Default)]
pub(crate) struct StructCountTree {
    pub struct_name: String,
    pub field_count: u32,
    pub struct_fields: Vec<StructCountTree>,
}

pub(crate) fn schema_to_struct_count_tree(
    src: &FieldType,
    dst: &mut Vec<StructCountTree>,
    as_map: &mut std::collections::HashSet<String>,
) {
    if src.primitive.is_some() {
        return;
    }
    if !src.r#struct.is_empty() {
        if as_map.contains(&src.r#struct) {
            return;
        }
        as_map.insert(src.r#struct.clone());
        dst.push(StructCountTree { struct_name: src.r#struct.clone(), field_count: 0, struct_fields: Vec::new() });
        return;
    }
    if let Some(arr) = &src.array {
        schema_to_struct_count_tree(&arr.elem_type, dst, as_map);
        return;
    }
    if !src.multimap.is_empty() {
        if as_map.contains(&src.multimap) {
            return;
        }
        as_map.insert(src.multimap.clone());
    }
}
