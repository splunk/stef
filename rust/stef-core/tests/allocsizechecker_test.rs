use stef_core::{AllocSizeChecker, limits::RECORD_ALLOC_LIMIT};

#[test]
fn test_add_alloc_size_saturates_on_overflow() {
    let mut checker = AllocSizeChecker::default();
    checker.add_alloc_size(usize::MAX - 5);
    checker.add_alloc_size(10);
    assert_eq!(checker.allocated_size(), usize::MAX);
    assert!(checker.is_over_limit());
}

#[test]
fn test_prep_alloc_size_n_handles_mul_overflow() {
    let mut checker = AllocSizeChecker::default();
    let err = checker.prep_alloc_size_n(usize::MAX, 2).unwrap_err();
    assert_eq!(checker.allocated_size(), usize::MAX);
    assert!(err.to_string().contains("record allocation limit exceeded"));
}

#[test]
fn test_prep_alloc_size_n_happy_path() {
    let mut checker = AllocSizeChecker::default();
    checker.prep_alloc_size_n(1024, 10).unwrap();
    assert_eq!(checker.allocated_size(), 10240);
    assert!(!checker.is_over_limit());
}

#[test]
fn test_prep_alloc_size_over_limit() {
    let mut checker = AllocSizeChecker::default();
    let err = checker.prep_alloc_size(RECORD_ALLOC_LIMIT + 1).unwrap_err();
    assert!(err.to_string().contains("record allocation limit exceeded"));
    assert_eq!(checker.allocated_size(), RECORD_ALLOC_LIMIT + 1);
    assert!(checker.is_over_limit());
}
