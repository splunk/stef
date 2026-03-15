# Rust Generator Progress

- 2026-03-15: Started continuation for Step 4.
- Existing known passing schemas before this session: `array_int64.stef`, `enum_array.stef`, `struct_reuse.stef`.
- Next targets per request: `multimap_string_string.stef`, then `multimap_struct.stef`.
- 2026-03-15: Fixed Rust multimap wire compatibility in `stefc/templates/rust/multimap.rs.tmpl`:
  - Full encoding now writes tagged header `((len << 1) | 1)` to match Go format.
  - Decoder now supports Go multimap modes: `0` (no change), values-only (`lsb=0`), full (`lsb=1`).
  - Count limit check aligned to Go behavior (`>= MULTIMAP_ELEM_COUNT_LIMIT`).
- 2026-03-15: Verified passing generation/tests for:
  - `array_int64.stef`
  - `enum_array.stef`
  - `struct_reuse.stef`
  - `multimap_string_string.stef` (newly passing)
- 2026-03-15: Attempted `multimap_struct.stef`:
  - Initial failure: stack overflow due recursive encoder/decoder init cycle.
  - Tried recursion-pointer approach in templates; reverted due unsafe recursive aliasing/instability.
  - Current status: `multimap_struct.stef` remains failing and needs a safer recursion strategy.
