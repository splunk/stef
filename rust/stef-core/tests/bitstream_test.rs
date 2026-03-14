use stef_core::{BitsReader, BitsWriter};

#[test]
fn test_write_bit() {
    let mut bw = BitsWriter::new(0);
    for _ in 0..11 {
        bw.write_bits(1, 1);
    }
    bw.close();
    assert_eq!(bw.bytes(), &[0b11111111, 0b11100000]);
}

#[test]
fn test_increasing_write_read_bits() {
    let mut bw = BitsWriter::new(0);
    const COUNT: u64 = 0x100000;
    let mut i = 1;
    while i <= COUNT {
        let bit_count = 64 - i.leading_zeros();
        bw.write_bits(i, bit_count);
        i += 111;
    }
    bw.close();

    let mut br = BitsReader::default();
    br.reset(bw.bytes());

    let mut i = 1;
    while i <= COUNT {
        let bit_count = 64 - i.leading_zeros();
        let val = br.read_bits(bit_count);
        assert_eq!(i, val);
        i += 111;
    }
    br.read_bits(64);
    assert_eq!(br.error(), Some(std::io::ErrorKind::UnexpectedEof));
}

#[test]
fn test_rand_write_read_bits() {
    let mut bw = BitsWriter::new(0);
    let mut rng = rand::rngs::StdRng::seed_from_u64(0);
    use rand::{Rng, SeedableRng};

    const COUNT: u64 = 0x10000;
    for _ in 0..COUNT {
        let shift = rng.random_range(0..64);
        let v = rng.random::<u64>() >> shift;
        let bit_count = if v == 0 { 0 } else { 64 - v.leading_zeros() };
        bw.write_bits(v, bit_count);
    }
    bw.close();

    let mut br = BitsReader::default();
    br.reset(bw.bytes());

    let mut rng = rand::rngs::StdRng::seed_from_u64(0);
    for _ in 0..COUNT {
        let shift = rng.random_range(0..64);
        let v = rng.random::<u64>() >> shift;
        let bit_count = if v == 0 { 0 } else { 64 - v.leading_zeros() };
        let val = br.read_bits(bit_count);
        assert_eq!(v, val);
    }
}
