//! Schema and wire-schema model used for compatibility checks.

#[path = "schema/recursestack.rs"]
mod recursestack;
#[path = "schema/structcounttree.rs"]
mod structcounttree;
#[path = "schema/utils.rs"]
mod utils;
#[path = "schema/wireschema.rs"]
mod wireschema;

use std::collections::{BTreeMap, HashMap, HashSet};

use crate::errors::{Result, StefError};

pub use utils::shrink_randomly;
pub use wireschema::{WireSchema, WireSchemaIter};

/// Schema compatibility classification.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Compatibility {
    /// Schemas match exactly.
    Exact,
    /// New schema is backward-compatible superset.
    Superset,
    /// Schemas are incompatible.
    Incompatible,
}

/// Full STEF schema model.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct Schema {
    /// Optional package path.
    #[serde(rename = "package", default, skip_serializing_if = "Vec::is_empty")]
    pub package_name: Vec<String>,
    /// Struct declarations keyed by name.
    pub structs: HashMap<String, Struct>,
    /// Multimap declarations keyed by name.
    #[serde(default)]
    pub multimaps: HashMap<String, Multimap>,
    /// Enum declarations keyed by name.
    #[serde(default)]
    pub enums: HashMap<String, Enum>,
}

impl Schema {
    /// Checks backward compatibility with `old_schema`.
    pub fn compatible(&self, old_schema: &Schema) -> std::result::Result<Compatibility, StefError> {
        let mut exact = self.structs.len() == old_schema.structs.len();

        for (name, old_struct) in &old_schema.structs {
            let new_struct = self
                .structs
                .get(name)
                .ok_or_else(|| StefError::message(format!("struct {name} does not exist in new schema")))?;
            let comp = self.compatible_struct(name, new_struct, old_struct)?;
            if comp == Compatibility::Superset {
                exact = false;
            }
        }

        for (name, old_map) in &old_schema.multimaps {
            let new_map = self
                .multimaps
                .get(name)
                .ok_or_else(|| StefError::message(format!("multimap {name} does not exist in new schema")))?;
            let comp = self.compatible_multimap(name, new_map, old_map)?;
            if comp == Compatibility::Superset {
                exact = false;
            }
        }

        if exact {
            Ok(Compatibility::Exact)
        } else {
            Ok(Compatibility::Superset)
        }
    }

    fn compatible_struct(&self, name: &str, new_struct: &Struct, old_struct: &Struct) -> Result<Compatibility> {
        if new_struct.fields.len() < old_struct.fields.len() {
            return Err(StefError::message(format!("new struct {name} has fewer fields than old struct")));
        }
        if new_struct.one_of != old_struct.one_of {
            return Err(StefError::message(format!(
                "new struct {name} has different oneof flag than theold struct"
            )));
        }
        if new_struct.dict_name != old_struct.dict_name {
            return Err(StefError::message(format!(
                "new struct {name} dictionary name is {}, old struct dictionary name is {}",
                new_struct.dict_name, old_struct.dict_name
            )));
        }

        let exact = new_struct.fields.len() == old_struct.fields.len();
        for i in 0..old_struct.fields.len() {
            is_compatible_field(name, i, &new_struct.fields[i], &old_struct.fields[i])?;
        }

        if exact {
            Ok(Compatibility::Exact)
        } else {
            Ok(Compatibility::Superset)
        }
    }

    fn compatible_multimap(&self, name: &str, new_map: &Multimap, old_map: &Multimap) -> Result<Compatibility> {
        if !is_compatible_field_type(&new_map.key.r#type, &old_map.key.r#type) {
            return Err(StefError::message(format!("multimap {name} key type does not match")));
        }
        if !is_compatible_field_type(&new_map.value.r#type, &old_map.value.r#type) {
            return Err(StefError::message(format!("multimap {name} value type does not match")));
        }
        Ok(Compatibility::Exact)
    }

    /// Returns a copy pruned to the specified root and reachable types.
    pub fn pruned_for_root(&self, root_struct_name: &str) -> Result<Schema> {
        let mut out = Schema { structs: HashMap::new(), multimaps: HashMap::new(), enums: HashMap::new(), package_name: self.package_name.clone() };
        self.copy_pruned_struct(root_struct_name, &mut out)?;
        out.resolve_refs()?;
        Ok(out)
    }

    /// Resolves references to structs/multimaps/enums.
    pub fn resolve_refs(&mut self) -> Result<()> {
        let struct_names: std::collections::HashSet<String> = self.structs.keys().cloned().collect();
        let multimap_names: std::collections::HashSet<String> = self.multimaps.keys().cloned().collect();
        let enum_names: std::collections::HashSet<String> = self.enums.keys().cloned().collect();

        let struct_keys: Vec<String> = self.structs.keys().cloned().collect();
        for k in struct_keys {
            let field_len = self.structs[&k].fields.len();
            for i in 0..field_len {
                let ft = &mut self.structs.get_mut(&k).expect("exists").fields[i].field_type;
                resolve_field_type(ft, &struct_names, &multimap_names, &enum_names)?;
            }
        }

        let mm_keys: Vec<String> = self.multimaps.keys().cloned().collect();
        for k in mm_keys {
            let key = &mut self.multimaps.get_mut(&k).expect("exists").key.r#type;
            resolve_field_type(key, &struct_names, &multimap_names, &enum_names)?;
            let value = &mut self.multimaps.get_mut(&k).expect("exists").value.r#type;
            resolve_field_type(value, &struct_names, &multimap_names, &enum_names)?;
        }

        self.compute_recursive()
    }

    fn copy_pruned_field_type(&self, field_type: &FieldType, dst: &mut Schema) -> Result<()> {
        if !field_type.r#struct.is_empty() {
            self.copy_pruned_struct(&field_type.r#struct, dst)?;
        } else if !field_type.multimap.is_empty() {
            self.copy_pruned_multimap(&field_type.multimap, dst)?;
        } else if let Some(arr) = &field_type.array {
            self.copy_pruned_field_type(&arr.elem_type, dst)?;
        } else if !field_type.r#enum.is_empty() {
            self.copy_pruned_enum(&field_type.r#enum, dst)?;
        }
        Ok(())
    }

    fn copy_pruned_struct(&self, struct_name: &str, dst: &mut Schema) -> Result<()> {
        if dst.structs.contains_key(struct_name) {
            return Ok(());
        }
        let src = self
            .structs
            .get(struct_name)
            .ok_or_else(|| StefError::message(format!("no struct named {struct_name} found")))?;

        let mut out = Struct {
            name: struct_name.to_string(),
            one_of: src.one_of,
            dict_name: src.dict_name.clone(),
            is_root: src.is_root,
            fields: Vec::with_capacity(src.fields.len()),
            field_map: HashMap::new(),
            recursive: false,
        };

        for f in &src.fields {
            let cloned = f.clone();
            self.copy_pruned_field_type(&cloned.field_type, dst)?;
            out.field_map.insert(cloned.name.clone(), cloned.clone());
            out.fields.push(cloned);
        }

        dst.structs.insert(struct_name.to_string(), out);
        Ok(())
    }

    fn copy_pruned_multimap(&self, multimap_name: &str, dst: &mut Schema) -> Result<()> {
        if dst.multimaps.contains_key(multimap_name) {
            return Ok(());
        }
        let src = self
            .multimaps
            .get(multimap_name)
            .ok_or_else(|| StefError::message(format!("no multimap named {multimap_name} found")))?;

        let out = Multimap { name: multimap_name.to_string(), key: src.key.clone(), value: src.value.clone(), recursive: false };
        self.copy_pruned_field_type(&out.key.r#type, dst)?;
        self.copy_pruned_field_type(&out.value.r#type, dst)?;
        dst.multimaps.insert(multimap_name.to_string(), out);
        Ok(())
    }

    fn copy_pruned_enum(&self, enum_name: &str, dst: &mut Schema) -> Result<()> {
        if dst.enums.contains_key(enum_name) {
            return Ok(());
        }
        let src = self
            .enums
            .get(enum_name)
            .ok_or_else(|| StefError::message(format!("no enum named {enum_name} found")))?;
        dst.enums.insert(enum_name.to_string(), src.clone());
        Ok(())
    }

    /// Returns field count of a struct.
    pub fn field_count(&self, struct_name: &str) -> Result<usize> {
        self.structs
            .get(struct_name)
            .map(|s| s.fields.len())
            .ok_or_else(|| StefError::message(format!("struct {struct_name} not found")))
    }

    fn compute_recursive(&mut self) -> Result<()> {
        let roots: Vec<String> = self.structs.values().filter(|s| s.is_root).map(|s| s.name.clone()).collect();
        for root in roots {
            let mut stack = recursestack::RecurseStack::default();
            self.compute_recursive_struct(&root, &mut stack)?;
        }
        Ok(())
    }

    fn compute_recursive_struct(&mut self, struct_name: &str, stack: &mut recursestack::RecurseStack) -> Result<()> {
        stack.as_stack.push(struct_name.to_string());
        stack.as_map.insert(struct_name.to_string());

        let fields = self
            .structs
            .get(struct_name)
            .ok_or_else(|| StefError::message(format!("struct {struct_name} not found")))?
            .fields
            .clone();

        for field in fields {
            self.compute_recursive_type(&field.field_type, stack)?;
        }

        stack.as_stack.pop();
        stack.as_map.remove(struct_name);
        Ok(())
    }

    fn compute_recursive_multimap(&mut self, multimap_name: &str, stack: &mut recursestack::RecurseStack) -> Result<()> {
        stack.as_stack.push(multimap_name.to_string());
        stack.as_map.insert(multimap_name.to_string());

        let mm = self
            .multimaps
            .get(multimap_name)
            .ok_or_else(|| StefError::message(format!("multimap {multimap_name} not found")))?
            .clone();

        self.compute_recursive_type(&mm.key.r#type, stack)?;
        self.compute_recursive_type(&mm.value.r#type, stack)?;

        stack.as_stack.pop();
        stack.as_map.remove(multimap_name);
        Ok(())
    }

    fn mark_recursive(&mut self, type_name: &str, stack: &recursestack::RecurseStack) {
        if stack.as_stack.iter().any(|s| s == type_name) {
            if let Some(s) = self.structs.get_mut(type_name) {
                s.recursive = true;
            }
            if let Some(m) = self.multimaps.get_mut(type_name) {
                m.recursive = true;
            }
        }
    }

    fn compute_recursive_type(&mut self, typ: &FieldType, stack: &mut recursestack::RecurseStack) -> Result<()> {
        if typ.primitive.is_some() {
            return Ok(());
        }
        if !typ.r#struct.is_empty() {
            if stack.as_map.contains(&typ.r#struct) {
                self.mark_recursive(&typ.r#struct, stack);
            } else {
                self.compute_recursive_struct(&typ.r#struct, stack)?;
            }
            return Ok(());
        }
        if !typ.multimap.is_empty() {
            if stack.as_map.contains(&typ.multimap) {
                self.mark_recursive(&typ.multimap, stack);
            } else {
                self.compute_recursive_multimap(&typ.multimap, stack)?;
            }
            return Ok(());
        }
        if let Some(arr) = &typ.array {
            self.compute_recursive_type(&arr.elem_type, stack)?;
            return Ok(());
        }
        Err(StefError::message("unknown type"))
    }

    /// Removes unreachable types and returns removed type sets.
    pub fn prune_unused(&mut self) -> Result<UnusedTypes> {
        let mut reachable_structs = HashSet::new();
        let mut reachable_multimaps = HashSet::new();
        let mut reachable_enums = HashSet::new();

        let roots: Vec<String> = self
            .structs
            .iter()
            .filter(|(_, s)| s.is_root)
            .map(|(n, _)| n.clone())
            .collect();
        for root in roots {
            self.mark_reachable_from_struct(&root, &mut reachable_structs, &mut reachable_multimaps, &mut reachable_enums);
        }

        let mut unused = UnusedTypes::default();
        for (name, s) in &self.structs {
            if !reachable_structs.contains(name) {
                unused.structs.push(s.clone());
            }
        }
        unused.structs.sort_by(|a, b| a.name.cmp(&b.name));

        for (name, m) in &self.multimaps {
            if !reachable_multimaps.contains(name) {
                unused.multimaps.push(m.clone());
            }
        }
        unused.multimaps.sort_by(|a, b| a.name.cmp(&b.name));

        for (name, e) in &self.enums {
            if !reachable_enums.contains(name) {
                unused.enums.push(e.clone());
            }
        }
        unused.enums.sort_by(|a, b| a.name.cmp(&b.name));

        for s in &unused.structs {
            self.structs.remove(&s.name);
        }
        for m in &unused.multimaps {
            self.multimaps.remove(&m.name);
        }
        for e in &unused.enums {
            self.enums.remove(&e.name);
        }

        Ok(unused)
    }

    fn mark_reachable_from_struct(
        &self,
        struct_name: &str,
        reachable_structs: &mut HashSet<String>,
        reachable_multimaps: &mut HashSet<String>,
        reachable_enums: &mut HashSet<String>,
    ) {
        if reachable_structs.contains(struct_name) {
            return;
        }
        let Some(s) = self.structs.get(struct_name) else { return };
        reachable_structs.insert(struct_name.to_string());
        for field in &s.fields {
            self.mark_reachable_from_field_type(&field.field_type, reachable_structs, reachable_multimaps, reachable_enums);
        }
    }

    fn mark_reachable_from_multimap(
        &self,
        multimap_name: &str,
        reachable_structs: &mut HashSet<String>,
        reachable_multimaps: &mut HashSet<String>,
        reachable_enums: &mut HashSet<String>,
    ) {
        if reachable_multimaps.contains(multimap_name) {
            return;
        }
        let Some(mm) = self.multimaps.get(multimap_name) else { return };
        reachable_multimaps.insert(multimap_name.to_string());
        self.mark_reachable_from_field_type(&mm.key.r#type, reachable_structs, reachable_multimaps, reachable_enums);
        self.mark_reachable_from_field_type(&mm.value.r#type, reachable_structs, reachable_multimaps, reachable_enums);
    }

    fn mark_reachable_from_field_type(
        &self,
        field_type: &FieldType,
        reachable_structs: &mut HashSet<String>,
        reachable_multimaps: &mut HashSet<String>,
        reachable_enums: &mut HashSet<String>,
    ) {
        if !field_type.r#struct.is_empty() {
            self.mark_reachable_from_struct(&field_type.r#struct, reachable_structs, reachable_multimaps, reachable_enums);
        } else if !field_type.multimap.is_empty() {
            self.mark_reachable_from_multimap(&field_type.multimap, reachable_structs, reachable_multimaps, reachable_enums);
        } else if !field_type.r#enum.is_empty() {
            reachable_enums.insert(field_type.r#enum.clone());
        } else if let Some(arr) = &field_type.array {
            self.mark_reachable_from_field_type(&arr.elem_type, reachable_structs, reachable_multimaps, reachable_enums);
        }
    }

    /// Produces deterministic SDL-like schema text.
    pub fn pretty_print(&self) -> String {
        let mut out = Vec::new();
        out.push(format!("package {}", self.package_name.join(".")));

        for enum_item in sorted_list(&self.enums) {
            out.push(pretty_print_enum(enum_item));
        }
        for mm in sorted_list(&self.multimaps) {
            out.push(pretty_print_multimap(mm));
        }
        for s in sorted_list(&self.structs) {
            out.push(pretty_print_struct(s));
        }

        out.join("\n\n")
    }
}

fn is_compatible_field(struct_name: &str, field_index: usize, new_field: &StructField, old_field: &StructField) -> Result<()> {
    if new_field.optional != old_field.optional {
        return Err(StefError::message(format!(
            "field {field_index} in new struct {struct_name} has different optional flag than the old struct"
        )));
    }
    if !is_compatible_field_type(&new_field.field_type, &old_field.field_type) {
        return Err(StefError::message(format!(
            "field {field_index} in new struct {struct_name} has a different type than the old struct"
        )));
    }
    Ok(())
}

fn resolve_field_type(
    field_type: &mut FieldType,
    struct_names: &std::collections::HashSet<String>,
    multimap_names: &std::collections::HashSet<String>,
    enum_names: &std::collections::HashSet<String>,
) -> Result<()> {
    let mut type_name = String::new();
    if !field_type.r#struct.is_empty() {
        type_name = field_type.r#struct.clone();
    } else if !field_type.multimap.is_empty() {
        type_name = field_type.multimap.clone();
    } else if !field_type.r#enum.is_empty() {
        type_name = field_type.r#enum.clone();
    }

    if !type_name.is_empty() {
        let mut matches = 0;
        if struct_names.contains(&type_name) {
            field_type.struct_def = Some(type_name.clone());
            matches += 1;
        }
        if multimap_names.contains(&type_name) {
            field_type.multimap_def = Some(type_name.clone());
            field_type.multimap = type_name.clone();
            field_type.r#struct.clear();
            matches += 1;
        }
        if enum_names.contains(&type_name) {
            field_type.primitive = Some(PrimitiveType { r#type: PrimitiveFieldType::Uint64 });
            field_type.r#enum = type_name.clone();
            field_type.r#struct.clear();
            matches += 1;
        }
        if matches == 0 {
            return Err(StefError::message(format!("unknown type: {type_name}")));
        }
        if matches > 1 {
            return Err(StefError::message(format!("ambiguous type: {type_name}")));
        }
        return Ok(());
    }

    if let Some(arr) = &mut field_type.array {
        resolve_field_type(&mut arr.elem_type, struct_names, multimap_names, enum_names)?;
    }
    Ok(())
}

fn is_compatible_field_type(new_field: &FieldType, old_field: &FieldType) -> bool {
    if new_field.primitive.is_some() != old_field.primitive.is_some() {
        return false;
    }
    if new_field.primitive != old_field.primitive {
        return false;
    }
    if new_field.array.is_some() != old_field.array.is_some() {
        return false;
    }
    if let (Some(a), Some(b)) = (&new_field.array, &old_field.array) {
        if !is_compatible_field_type(&a.elem_type, &b.elem_type) {
            return false;
        }
    }
    new_field.r#struct == old_field.r#struct
        && new_field.multimap == old_field.multimap
        && new_field.dict_name == old_field.dict_name
}

fn sorted_list<T>(m: &HashMap<String, T>) -> Vec<&T> {
    let mut sorted = BTreeMap::new();
    for (k, v) in m {
        sorted.insert(k, v);
    }
    sorted.into_values().collect()
}

fn pretty_print_enum(e: &Enum) -> String {
    let mut out = format!("enum {} {{", e.name);
    for f in &e.fields {
        out.push_str(&format!("\n  {} = {}", f.name, f.value));
    }
    out.push_str("\n}");
    out
}

fn pretty_print_multimap(m: &Multimap) -> String {
    let mut out = format!("multimap {} {{\n", m.name);
    out.push_str(&format!("  key {}", pretty_print_field_type(&m.key.r#type)));
    if !m.key.r#type.dict_name.is_empty() {
        out.push_str(&format!(" dict({})", m.key.r#type.dict_name));
    }
    out.push('\n');
    out.push_str(&format!("  value {}", pretty_print_field_type(&m.value.r#type)));
    if !m.value.r#type.dict_name.is_empty() {
        out.push_str(&format!(" dict({})", m.value.r#type.dict_name));
    }
    out.push_str("\n}");
    out
}

fn pretty_print_struct(s: &Struct) -> String {
    let mut out = if s.one_of {
        format!("oneof {} {{", s.name)
    } else {
        let mut h = format!("struct {}", s.name);
        if !s.dict_name.is_empty() {
            h.push_str(&format!(" dict({})", s.dict_name));
        }
        if s.is_root {
            h.push_str(" root");
        }
        h.push_str(" {");
        h
    };

    for f in &s.fields {
        out.push_str("\n  ");
        out.push_str(&pretty_print_struct_field(f));
    }
    out.push_str("\n}");
    out
}

fn pretty_print_struct_field(f: &StructField) -> String {
    let mut out = format!("{} {}", f.name, pretty_print_field_type(&f.field_type));
    if !f.field_type.dict_name.is_empty() {
        out.push_str(&format!(" dict({})", f.field_type.dict_name));
    }
    if f.optional {
        out.push_str(" optional");
    }
    out
}

fn pretty_print_field_type(ft: &FieldType) -> String {
    if let Some(p) = &ft.primitive {
        return match p.r#type {
            PrimitiveFieldType::Int64 => "int64".into(),
            PrimitiveFieldType::Uint64 => "uint64".into(),
            PrimitiveFieldType::Float64 => "float64".into(),
            PrimitiveFieldType::Bool => "bool".into(),
            PrimitiveFieldType::String => "string".into(),
            PrimitiveFieldType::Bytes => "bytes".into(),
        };
    }
    if let Some(arr) = &ft.array {
        return format!("[]{}", pretty_print_field_type(&arr.elem_type));
    }
    if !ft.r#struct.is_empty() {
        return ft.r#struct.clone();
    }
    if !ft.multimap.is_empty() {
        return ft.multimap.clone();
    }
    if !ft.r#enum.is_empty() {
        return ft.r#enum.clone();
    }
    "unknown".into()
}

/// Struct declaration.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct Struct {
    #[serde(default)]
    pub name: String,
    #[serde(default, rename = "oneof")]
    pub one_of: bool,
    #[serde(default, rename = "dict")]
    pub dict_name: String,
    #[serde(default, rename = "root")]
    pub is_root: bool,
    pub fields: Vec<StructField>,
    #[serde(skip)]
    pub field_map: HashMap<String, StructField>,
    #[serde(skip)]
    pub recursive: bool,
}

impl Struct {
    /// Creates an empty struct with initialized field map.
    pub fn new() -> Self {
        Self { field_map: HashMap::new(), ..Self::default() }
    }

    /// Returns true if struct is recursive.
    pub fn recursive(&self) -> bool {
        self.recursive
    }

    /// Returns true if field exists by name.
    pub fn has_field(&self, name: &str) -> bool {
        self.field_map.contains_key(name)
    }

    /// Adds a field and updates field map.
    pub fn add_field(&mut self, field: StructField) {
        self.field_map.insert(field.name.clone(), field.clone());
        self.fields.push(field);
    }
}

/// Struct field declaration.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct StructField {
    #[serde(flatten)]
    pub field_type: FieldType,
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub optional: bool,
}

/// Primitive field type kind.
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum PrimitiveFieldType {
    #[serde(rename = "int64")]
    Int64,
    #[serde(rename = "uint64")]
    Uint64,
    #[serde(rename = "float64")]
    Float64,
    #[serde(rename = "bool")]
    Bool,
    #[serde(rename = "string")]
    String,
    #[serde(rename = "bytes")]
    Bytes,
}

/// Primitive type wrapper.
#[derive(Debug, Clone, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub struct PrimitiveType {
    #[serde(rename = "type")]
    pub r#type: PrimitiveFieldType,
}

/// Array type wrapper.
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ArrayType {
    #[serde(rename = "elem")]
    pub elem_type: Box<FieldType>,
    #[serde(skip)]
    pub recursive: bool,
}

/// Field type union used by struct fields and multimap key/value.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct FieldType {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub primitive: Option<PrimitiveType>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub array: Option<Box<ArrayType>>,
    #[serde(default, rename = "struct")]
    pub r#struct: String,
    #[serde(default, rename = "multimap")]
    pub multimap: String,
    #[serde(default)]
    pub r#enum: String,
    #[serde(default, rename = "dict")]
    pub dict_name: String,
    #[serde(skip)]
    pub struct_def: Option<String>,
    #[serde(skip)]
    pub multimap_def: Option<String>,
}

impl FieldType {
    /// Sets recursive marker on container type represented by this field type.
    pub fn set_recursive(&mut self) {
        if let Some(arr) = &mut self.array {
            arr.recursive = true;
        }
    }

    /// Returns recursive marker when applicable.
    pub fn recursive(&self) -> bool {
        self.array.as_ref().map(|a| a.recursive).unwrap_or(false)
    }
}

/// Multimap field declaration.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct MultimapField {
    #[serde(rename = "type")]
    pub r#type: FieldType,
}

/// Multimap declaration.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct Multimap {
    #[serde(default)]
    pub name: String,
    pub key: MultimapField,
    pub value: MultimapField,
    #[serde(skip)]
    pub recursive: bool,
}

/// Enum declaration.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct Enum {
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub fields: Vec<EnumField>,
}

/// Enum value declaration.
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct EnumField {
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub value: u64,
}

/// Set of unreachable types removed by `Schema::prune_unused`.
#[derive(Debug, Clone, Default)]
pub struct UnusedTypes {
    pub structs: Vec<Struct>,
    pub multimaps: Vec<Multimap>,
    pub enums: Vec<Enum>,
}
