use crate::{bitstream::{BitsReader, BitsWriter}, recordbuf::{ReadColumnSet, WriteColumnSet}, SizeLimiter};

/// Gorila-style float64 encoder.
#[derive(Default)]
pub struct Float64Encoder {
    buf: BitsWriter,
    limiter: Option<*mut SizeLimiter>,
    last_val: f64,
    leading_bits: i32,
    trailing_bits: i32,
}

impl Float64Encoder {
    pub fn init(&mut self, limiter: &mut SizeLimiter, _columns: &mut WriteColumnSet) {
        self.limiter = Some(limiter);
    }

    pub fn is_equal(&self, val: f64) -> bool {
        self.last_val == val
    }

    pub fn encode(&mut self, val: f64) {
        let xor_val = val.to_bits() ^ self.last_val.to_bits();
        self.last_val = val;
        if xor_val == 0 {
            self.buf.write_bit(0);
            if let Some(ptr) = self.limiter {
                unsafe { (*ptr).add_frame_bits(1) };
            }
            return;
        }

        let mut leading = xor_val.leading_zeros() as i32;
        if leading >= 32 {
            leading = 31;
        }
        let trailing = xor_val.trailing_zeros() as i32;
        let prev_leading = self.leading_bits;
        let prev_trailing = self.trailing_bits;
        let sigbits = 64 - leading - trailing;

        if leading >= prev_leading && trailing >= prev_trailing {
            if 53 - prev_leading - prev_trailing <= sigbits {
                self.buf.write_bits(0b10, 2);
                let bit_count = (64 - prev_leading - prev_trailing) as u32;
                self.buf.write_bits(xor_val >> prev_trailing, bit_count);
                if let Some(ptr) = self.limiter {
                    unsafe { (*ptr).add_frame_bits((2 + bit_count) as usize) };
                }
                return;
            }
        }

        self.leading_bits = leading;
        self.trailing_bits = trailing;

        let mut bits_val = 0b11u64;
        bits_val = (bits_val << 5) | (leading as u64);
        bits_val = (bits_val << 6) | ((sigbits - 1) as u64);
        self.buf.write_bits(bits_val, 13);
        self.buf.write_bits(xor_val >> trailing, sigbits as u32);
        if let Some(ptr) = self.limiter {
            unsafe { (*ptr).add_frame_bits((13 + sigbits) as usize) };
        }
    }

    pub fn collect_columns(&mut self, column_set: &mut WriteColumnSet) {
        column_set.set_bits(&mut self.buf);
    }

    pub fn reset(&mut self) {
        self.last_val = 0.0;
        self.leading_bits = 0;
        self.trailing_bits = 0;
    }
}

/// Gorilla float64 decoder.
#[derive(Default)]
pub struct Float64Decoder {
    buf: BitsReader,
    column: Vec<u8>,
    last_val: f64,
    leading_bits: u64,
    trailing_bits: u64,
}

impl Float64Decoder {
    pub fn init(&mut self, columns: &mut ReadColumnSet) {
        self.column = columns.column().data().to_vec();
    }

    pub fn continue_(&mut self) {
        self.buf.reset(&self.column);
    }

    pub fn decode(&mut self, dst: &mut f64) {
        const NON_IDENTICAL: u64 = 0b1000000000000;
        const NEW_LEADING_TRAILING: u64 = 0b0100000000000;
        const LEADING_MASK: u64 = 0b0011111000000;
        const SIG_MASK: u64 = 0b0000000111111;
        const SIG_BITS_COUNT: u64 = 6;

        let hdr_bits = self.buf.peek_bits(13);
        if hdr_bits & NON_IDENTICAL == 0 {
            self.buf.consume(1);
            *dst = self.last_val;
            return;
        }

        let (leading, trailing, sigbits) = if hdr_bits & NEW_LEADING_TRAILING == 0 {
            self.buf.consume(2);
            let l = self.leading_bits;
            let t = self.trailing_bits;
            (l, t, 64 - l - t)
        } else {
            self.buf.consume(13);
            let l = (hdr_bits & LEADING_MASK) >> SIG_BITS_COUNT;
            let mut s = hdr_bits & SIG_MASK;
            s += 1;
            let t = 64 - l - s;
            self.leading_bits = l;
            self.trailing_bits = t;
            (l, t, s)
        };

        let mut xor_val = self.buf.read_bits(sigbits as u32);
        xor_val <<= trailing;
        self.last_val = f64::from_bits(xor_val ^ self.last_val.to_bits());
        *dst = self.last_val;
    }

    pub fn reset(&mut self) {
        self.last_val = 0.0;
        self.leading_bits = 0;
        self.trailing_bits = 0;
    }
}
