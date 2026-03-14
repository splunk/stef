/// Counts optional fields among the first `keep_field_count` fields.
pub fn optional_field_count(optionals_mask: u64, keep_field_count: u32) -> u32 {
    let keep_field_mask = if keep_field_count == 64 {
        u64::MAX
    } else {
        !(!0u64 << keep_field_count)
    };
    (optionals_mask & keep_field_mask).count_ones()
}
