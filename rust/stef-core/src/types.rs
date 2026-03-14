use rand::Rng;

/// Immutable byte sequence represented as raw bytes.
pub type Bytes = Vec<u8>;

/// Empty immutable byte sequence.
pub const EMPTY_BYTES: &[u8] = &[];

pub fn uint64_compare(left: u64, right: u64) -> i8 {
    match left.cmp(&right) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    }
}

pub fn int64_compare(left: i64, right: i64) -> i8 {
    match left.cmp(&right) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    }
}

pub fn bool_compare(left: bool, right: bool) -> i8 {
    match left.cmp(&right) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    }
}

pub fn float64_compare(left: f64, right: f64) -> i8 {
    if left > right {
        1
    } else if left < right {
        -1
    } else {
        0
    }
}

pub fn string_compare(left: &str, right: &str) -> i8 {
    match left.cmp(right) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    }
}

pub fn bytes_compare(left: &[u8], right: &[u8]) -> i8 {
    match left.cmp(right) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    }
}

pub fn uint64_random<R: Rng + ?Sized>(random: &mut R) -> u64 {
    random.random::<u64>()
}

pub fn int64_random<R: Rng + ?Sized>(random: &mut R) -> i64 {
    random.random::<i64>()
}

pub fn bool_random<R: Rng + ?Sized>(random: &mut R) -> bool {
    random.random::<u8>() % 2 == 0
}

pub fn float64_random<R: Rng + ?Sized>(random: &mut R) -> f64 {
    random.random::<f64>()
}

pub fn string_random<R: Rng + ?Sized>(random: &mut R) -> String {
    format!("{}", random.random::<u8>() % 10)
}

pub fn bytes_random<R: Rng + ?Sized>(random: &mut R) -> Bytes {
    string_random(random).into_bytes()
}
