#[cfg(bench_nightly)]
mod benches {
    extern crate test;

    use test::Bencher;
    use crate::{AllocSizeChecker, BitsReader, BitsWriter, BytesWriter};

    #[bench]
    fn bench_add_alloc_size(b: &mut Bencher) {
        b.iter(|| {
            let mut checker = AllocSizeChecker::default();
            checker.add_alloc_size(64);
            test::black_box(checker);
        });
    }

    #[bench]
    fn bench_bstream_write_bit(b: &mut Bencher) {
        b.iter(|| {
            let mut bw = BitsWriter::new(1_000_000);
            bw.reset();
            for j in 0..1_000_000 {
                bw.write_bit((j % 2) as u32);
            }
            test::black_box(bw);
        });
    }

    #[bench]
    fn bench_bstream_read_bit(b: &mut Bencher) {
        let mut bw = BitsWriter::new(0);
        for j in 0..1_000_000 {
            bw.write_bit((j % 2) as u32);
        }
        bw.close();
        let byts = bw.bytes().to_vec();
        let mut br = BitsReader::default();

        b.iter(|| {
            br.reset(&byts);
            for j in 0..1_000_000 {
                let v = br.read_bit();
                if v != (j % 2) as u64 {
                    panic!("invalid value");
                }
            }
            test::black_box(&br);
        });
    }

    #[bench]
    fn bench_bstream_read_bits(b: &mut Bencher) {
        let mut bw = BitsWriter::new(0);
        let mut val = 1u64;
        for j in 1..64u32 {
            bw.write_bits(val, j);
            val *= 2;
        }
        bw.close();
        let byts = bw.bytes().to_vec();
        let mut br = BitsReader::default();

        b.iter(|| {
            br.reset(&byts);
            let mut val = 1u64;
            for j in 1..64u32 {
                let v = br.read_bits(j);
                if v != val {
                    panic!("mismatch");
                }
                val *= 2;
            }
            test::black_box(&br);
        });
    }

    #[bench]
    fn bench_bstream_write_uvarint_compact_small(b: &mut Bencher) {
        let mut bw = BitsWriter::new(1000);
        b.iter(|| {
            bw.reset();
            for j in 0..47u64 {
                bw.write_uvarint_compact(j);
            }
            test::black_box(&bw);
        });
    }

    #[bench]
    fn bench_bstream_read_uvarint_compact_small(b: &mut Bencher) {
        let mut bw = BitsWriter::new(0);
        for j in 0..47u64 {
            bw.write_uvarint_compact(j);
        }
        bw.close();
        let byts = bw.bytes().to_vec();
        let mut br = BitsReader::default();

        b.iter(|| {
            br.reset(&byts);
            for j in 0..47u64 {
                let v = br.read_uvarint_compact();
                if v != j {
                    panic!("invalid value");
                }
            }
            test::black_box(&br);
        });
    }

    #[bench]
    fn bench_bstream_write_uvarint_compact(b: &mut Bencher) {
        let mut bw = BitsWriter::new(1000);
        b.iter(|| {
            bw.reset();
            for j in 0..47u32 {
                bw.write_uvarint_compact(1u64 << j);
            }
            test::black_box(&bw);
        });
    }

    #[bench]
    fn bench_bstream_read_uvarint_compact(b: &mut Bencher) {
        let mut bw = BitsWriter::new(0);
        for j in 0..47u32 {
            bw.write_uvarint_compact(1u64 << j);
        }
        bw.close();
        let byts = bw.bytes().to_vec();
        let mut br = BitsReader::default();

        b.iter(|| {
            br.reset(&byts);
            for j in 0..47u32 {
                let v = br.read_uvarint_compact();
                if v != (1u64 << j) {
                    panic!("invalid value");
                }
            }
            test::black_box(&br);
        });
    }

    #[bench]
    fn bench_membuf_write_varuint(b: &mut Bencher) {
        b.iter(|| {
            let mut bw = BytesWriter::new(10_000_000);
            for j in 0..1_000_000 {
                bw.write_uvarint(j as u64);
            }
            test::black_box(bw);
        });
    }

    #[bench]
    fn bench_membuf_read_varuint_exp(b: &mut Bencher) {
        let mut bw = BytesWriter::new(0);
        let mut val = 1u64;
        for _ in 0..63 {
            bw.write_uvarint(val);
            val *= 2;
        }
        let byts = bw.as_vec();
        let mut br = crate::BytesReader::default();
        br.reset(byts);

        b.iter(|| {
            br.rewind();
            let mut check_val = 1u64;
            for _ in 0..63 {
                let val = br.read_uvarint().expect("read_uvarint");
                if val != check_val {
                    panic!("invalid value");
                }
                check_val *= 2;
            }
            test::black_box(&br);
        });
    }

    macro_rules! bench_membuf_write_varuint_sizes {
        ($name:ident, $size:expr) => {
            #[bench]
            fn $name(b: &mut Bencher) {
                let val: u64 = (1u64 << ($size * 7)) - 1;
                b.iter(|| {
                    let mut bw = BytesWriter::new(1000);
                    bw.reset();
                    for _ in 0..1000 {
                        bw.write_uvarint(val);
                    }
                    test::black_box(&bw);
                });
            }
        };
    }

    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_1, 1);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_2, 2);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_3, 3);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_4, 4);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_5, 5);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_6, 6);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_7, 7);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_8, 8);
    bench_membuf_write_varuint_sizes!(bench_membuf_write_varuint_size_9, 9);
}
