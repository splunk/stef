/// Resizes a vector to `data_len`, preserving existing values.
pub fn ensure_len<T: Default + Clone>(data: &mut Vec<T>, data_len: usize) {
    if data.len() < data_len {
        data.resize(data_len, T::default());
    } else {
        data.truncate(data_len);
    }
}
