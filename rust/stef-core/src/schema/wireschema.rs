use std::collections::HashSet;

use crate::{errors::StefError, membuffer::decode_uvarint, schema::{Compatibility, Schema}};

const MAX_STRUCT_COUNT: usize = 1024;

/// Compact schema subset used for wire compatibility checks.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct WireSchema {
    pub(crate) struct_counts: Vec<u32>,
}

impl WireSchema {
    /// Builds wire schema reachable from `root_struct_name`.
    pub fn new(schema: &Schema, root_struct_name: &str) -> Self {
        let mut out = WireSchema::default();
        let mut seen = HashSet::new();

        fn dfs(schema: &Schema, struct_name: &str, seen: &mut HashSet<String>, out: &mut WireSchema) {
            if seen.contains(struct_name) {
                return;
            }
            let Some(s) = schema.structs.get(struct_name) else { return };
            seen.insert(struct_name.to_string());
            out.struct_counts.push(s.fields.len() as u32);
            for field in &s.fields {
                if !field.field_type.r#struct.is_empty() {
                    dfs(schema, &field.field_type.r#struct, seen, out);
                } else if let Some(arr) = &field.field_type.array {
                    if !arr.elem_type.r#struct.is_empty() {
                        dfs(schema, &arr.elem_type.r#struct, seen, out);
                    }
                }
            }
        }

        dfs(schema, root_struct_name, &mut seen, &mut out);
        out
    }

    /// Serializes wire schema as varint count + varint values.
    pub fn serialize(&self, dst: &mut Vec<u8>) {
        append_uvarint(dst, self.struct_counts.len() as u64);
        for c in &self.struct_counts {
            append_uvarint(dst, *c as u64);
        }
    }

    /// Deserializes wire schema from bytes.
    pub fn deserialize(&mut self, src: &[u8]) -> Result<(), StefError> {
        let mut ofs = 0usize;
        let (count_u64, n) = decode_uvarint(&src[ofs..])?;
        ofs += n;
        let count = count_u64 as usize;
        if count > MAX_STRUCT_COUNT {
            return Err(StefError::message("struct count limit exceeded"));
        }
        self.struct_counts.clear();
        self.struct_counts.reserve(count);
        for _ in 0..count {
            let (field_count, m) = decode_uvarint(&src[ofs..])?;
            ofs += m;
            self.struct_counts.push(field_count as u32);
        }
        Ok(())
    }

    /// Performs backward compatibility check against `old_schema`.
    pub fn compatible(&self, old_schema: &WireSchema) -> Result<Compatibility, StefError> {
        if self.struct_counts.len() > old_schema.struct_counts.len() {
            return Ok(Compatibility::Superset);
        }
        if self.struct_counts.len() < old_schema.struct_counts.len() {
            return Err(StefError::message(format!(
                "new schema has fewers structs than old schema ({} vs {})",
                self.struct_counts.len(),
                old_schema.struct_counts.len()
            )));
        }
        let new_total: u32 = self.struct_counts.iter().sum();
        let old_total: u32 = old_schema.struct_counts.iter().sum();
        if new_total > old_total {
            return Ok(Compatibility::Superset);
        }
        if new_total < old_total {
            return Err(StefError::message(format!(
                "new schema has fewers fields than old schema ({} vs {})",
                new_total, old_total
            )));
        }
        Ok(Compatibility::Exact)
    }

    /// Returns struct counts for testing and integrations.
    pub fn struct_counts(&self) -> &[u32] {
        &self.struct_counts
    }
}

/// Iterator over wire-schema struct field counts.
#[derive(Debug, Clone)]
pub struct WireSchemaIter<'a> {
    schema: &'a WireSchema,
    struct_idx: usize,
}

impl<'a> WireSchemaIter<'a> {
    /// Creates a new iterator.
    pub fn new(schema: &'a WireSchema) -> Self {
        Self { schema, struct_idx: 0 }
    }

    /// Returns next field count.
    pub fn next_field_count(&mut self) -> Result<u32, StefError> {
        if self.struct_idx >= self.schema.struct_counts.len() {
            return Err(StefError::message("struct count limit exceeded"));
        }
        let c = self.schema.struct_counts[self.struct_idx];
        self.struct_idx += 1;
        Ok(c)
    }

    /// Returns true when iteration consumed all counts.
    pub fn done(&self) -> bool {
        self.struct_idx >= self.schema.struct_counts.len()
    }
}

fn append_uvarint(dst: &mut Vec<u8>, mut v: u64) {
    loop {
        let mut b = (v & 0x7f) as u8;
        v >>= 7;
        if v != 0 {
            b |= 0x80;
        }
        dst.push(b);
        if v == 0 {
            break;
        }
    }
}
