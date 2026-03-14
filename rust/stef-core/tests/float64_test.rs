use rand::{Rng, SeedableRng};
use stef_core::{SizeLimiter, WriteColumnSet, ReadColumnSet, codecs::float64::{Float64Decoder, Float64Encoder}};

fn encode(values: &[f64]) -> Vec<u8> {
    let mut enc = Float64Encoder::default();
    let mut limiter = SizeLimiter::default();
    enc.init(&mut limiter, &mut WriteColumnSet::default());
    for &v in values {
        enc.encode(v);
    }
    let mut col = WriteColumnSet::default();
    enc.collect_columns(&mut col);
    col.data().to_vec()
}

fn decode(bytes: Vec<u8>, n: usize) -> Vec<f64> {
    let mut read_cols = ReadColumnSet::default();
    read_cols.set_column_data(bytes);

    let mut dec = Float64Decoder::default();
    dec.init(&mut read_cols);
    dec.continue_();

    let mut out = Vec::with_capacity(n);
    for _ in 0..n {
        let mut v = 0.0;
        dec.decode(&mut v);
        out.push(v);
    }
    out
}

#[test]
fn test_float64_basic() {
    let values = [1.0, 1.0, 2.0, 2.0, 3.1415, -1.0, 0.0, -0.0, f64::INFINITY, f64::NEG_INFINITY];
    let bytes = encode(&values);
    let got = decode(bytes, values.len());
    assert_eq!(got, values);
}

#[test]
fn test_float64_random_sequence() {
    let mut rng = rand::rngs::StdRng::seed_from_u64(77);
    let mut vals = Vec::new();
    for _ in 0..1000 {
        vals.push(rng.random::<f64>() * 1e6 - 5e5);
    }
    let bytes = encode(&vals);
    let got = decode(bytes, vals.len());
    assert_eq!(got, vals);
}
