use std::collections::HashMap;

use stef_core::VarHeader;

#[test]
fn test_var_header_serialization() {
    let mut tests = Vec::new();
    tests.push(VarHeader::default());
    tests.push(VarHeader { schema_wire_bytes: vec![b'0', b'1', b'2'], user_data: HashMap::new() });
    let mut map = HashMap::new();
    map.insert("abc".to_string(), "def".to_string());
    map.insert("0".to_string(), "world".to_string());
    tests.push(VarHeader { schema_wire_bytes: b"012345".to_vec(), user_data: map });

    for orig in tests {
        let mut buf = Vec::new();
        orig.serialize(&mut buf).unwrap();

        let mut cpy = VarHeader::default();
        cpy.deserialize(&buf).unwrap();

        assert_eq!(orig.schema_wire_bytes, cpy.schema_wire_bytes);
        assert_eq!(orig.user_data, cpy.user_data);
    }
}
