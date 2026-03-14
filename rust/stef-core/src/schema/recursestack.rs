use std::collections::HashSet;

#[derive(Debug, Default)]
pub(crate) struct RecurseStack {
    pub fields: Vec<String>,
    pub as_stack: Vec<String>,
    pub as_map: HashSet<String>,
}
