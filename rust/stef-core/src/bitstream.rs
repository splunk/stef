use std::io::ErrorKind;

use crate::bitstream_lookuptables::{
    READ_CONSUME_COUNT_BY_ZEROS, READ_MASK_BY_ZEROS, READ_SHIFT_BY_ZEROS, WRITE_BITS_COUNT_BY_ZEROS,
    WRITE_MASK_BY_ZEROS,
};

/// Bit-level writer used by compact codecs.
#[derive(Debug, Clone)]
pub struct BitsWriter {
    stream: Vec<u8>,
    bits_buf: u64,
    bits_buf_used: u32,
}

impl BitsWriter {
    /// Creates a writer with reserved stream capacity.
    pub fn new(size: usize) -> Self {
        Self { stream: Vec::with_capacity(size), bits_buf: 0, bits_buf_used: 0 }
    }

    /// Clears writer state while preserving allocation.
    pub fn reset(&mut self) {
        self.stream.clear();
        self.bits_buf = 0;
        self.bits_buf_used = 0;
    }

    /// Flushes partial byte-aligned suffix into stream.
    pub fn close(&mut self) {
        let target_len = self.stream.len() + ((self.bits_buf_used + 7) / 8) as usize;
        self.stream.extend_from_slice(&self.bits_buf.to_be_bytes());
        self.stream.truncate(target_len);
    }

    /// Returns encoded bytes.
    pub fn bytes(&self) -> &[u8] {
        &self.stream
    }

    /// Returns number of bits currently written.
    pub fn bit_count(&self) -> u32 {
        (self.stream.len() as u32) * 8 + self.bits_buf_used
    }

    /// Writes one bit.
    pub fn write_bit(&mut self, bit: u32) {
        if self.bits_buf_used <= 63 {
            self.bits_buf |= (bit as u64) << (63 - self.bits_buf_used);
            self.bits_buf_used += 1;
            return;
        }
        self.write_bits_slow(bit as u64, 1);
    }

    /// Writes `nbits` least-significant bits from `val`.
    pub fn write_bits(&mut self, val: u64, nbits: u32) {
        if nbits == 0 {
            return;
        }
        let complement = 64 - nbits;
        if self.bits_buf_used <= complement {
            self.bits_buf |= val << (complement - self.bits_buf_used);
            self.bits_buf_used += nbits;
            return;
        }
        self.write_bits_slow(val, nbits);
    }

    fn write_bits_slow(&mut self, val: u64, mut nbits: u32) {
        let bits_buf_free = 64 - self.bits_buf_used;
        if bits_buf_free == 0 {
            self.stream.extend_from_slice(&self.bits_buf.to_be_bytes());
            self.bits_buf = 0;
            self.bits_buf_used = 0;
            self.write_bits(val, nbits);
            return;
        }
        if bits_buf_free >= nbits {
            self.bits_buf |= val << (bits_buf_free - nbits);
            self.bits_buf_used += nbits;
            return;
        }
        self.bits_buf |= val >> (nbits - bits_buf_free);
        self.stream.extend_from_slice(&self.bits_buf.to_be_bytes());
        nbits -= bits_buf_free;
        self.bits_buf = val << (64 - nbits);
        self.bits_buf_used = nbits;
    }

    /// Writes signed compact varint in [-2^47, 2^47-1].
    pub fn write_varint_compact(&mut self, val: i64) -> u32 {
        let ux = ((val >> 63) as u64) ^ ((val as u64) << 1);
        self.write_uvarint_compact(ux)
    }

    /// Writes unsigned compact varint in [0, 2^48-1].
    pub fn write_uvarint_compact(&mut self, mut val: u64) -> u32 {
        let zeros = val.leading_zeros() as usize;
        val |= WRITE_MASK_BY_ZEROS[zeros];
        let bit_count = WRITE_BITS_COUNT_BY_ZEROS[zeros];
        self.write_bits(val, bit_count);
        bit_count
    }
}

impl Default for BitsWriter {
    fn default() -> Self {
        Self::new(0)
    }
}

/// Bit-level reader used by compact codecs.
#[derive(Debug, Clone, Default)]
pub struct BitsReader {
    bit_buf: u64,
    buf: Vec<u8>,
    byte_index: usize,
    avail_bit_count: u32,
    last_error: Option<std::io::ErrorKind>,
}

impl BitsReader {
    /// Reinitializes reader with a new byte slice.
    pub fn reset(&mut self, buf: &[u8]) {
        self.buf.clear();
        self.buf.extend_from_slice(buf);
        self.byte_index = 0;
        self.avail_bit_count = 0;
        self.bit_buf = 0;
        self.last_error = None;
    }

    /// Returns EOF status from previous reads.
    pub fn error(&self) -> Option<std::io::ErrorKind> {
        self.last_error
    }

    fn refill_slow(&mut self) {
        if self.byte_index >= self.buf.len() {
            self.last_error = Some(ErrorKind::UnexpectedEof);
            return;
        }
        while self.byte_index < self.buf.len() && self.avail_bit_count < 56 {
            let byt = self.buf[self.byte_index] as u64;
            self.bit_buf |= byt << (64 - self.avail_bit_count - 8);
            self.byte_index += 1;
            self.avail_bit_count += 8;
        }
        if self.byte_index >= self.buf.len() {
            self.avail_bit_count += 56;
        }
    }

    /// Peeks up to 56 bits, appending implicit zero bits at EOF.
    pub fn peek_bits(&mut self, nbits: u32) -> u64 {
        if nbits <= self.avail_bit_count {
            return self.bit_buf >> (64 - nbits);
        }
        self.refill_and_peek_bits(nbits)
    }

    /// Peeks one bit, appending implicit zero bits at EOF.
    pub fn peek_bit(&mut self) -> u64 {
        if self.avail_bit_count >= 1 {
            return self.bit_buf >> 63;
        }
        self.refill_and_peek_bits(1)
    }

    fn refill_and_peek_bits(&mut self, nbits: u32) -> u64 {
        assert!(nbits <= 56, "at most 56 bits can be peeked");
        if self.byte_index + 8 < self.buf.len() {
            let u = u64::from_be_bytes(self.buf[self.byte_index..self.byte_index + 8].try_into().expect("slice"));
            self.bit_buf |= u >> self.avail_bit_count;
            self.byte_index += ((63 - self.avail_bit_count) >> 3) as usize;
            self.avail_bit_count |= 56;
        } else {
            self.refill_slow();
        }
        self.bit_buf >> (64 - nbits)
    }

    /// Consumes previously peeked bits.
    pub fn consume(&mut self, nbits: u32) {
        if nbits == 0 {
            return;
        }
        if nbits >= 64 {
            self.bit_buf = 0;
        } else {
            self.bit_buf <<= nbits;
        }
        self.avail_bit_count = self.avail_bit_count.saturating_sub(nbits);
    }

    /// Reads up to 64 bits, appending implicit zero bits at EOF.
    pub fn read_bits(&mut self, nbits: u32) -> u64 {
        if nbits == 0 {
            return 0;
        }
        if nbits <= 56 {
            let v = self.peek_bits(nbits);
            self.consume(nbits);
            return v;
        }
        self.read_bits_more_than_56(nbits)
    }

    fn read_bits_more_than_56(&mut self, mut nbits: u32) -> u64 {
        let mut val = self.peek_bits(56);
        let mut to_consume = self.avail_bit_count;
        if to_consume > 56 {
            to_consume = 56;
        }
        self.consume(to_consume);
        nbits -= to_consume;
        val = (val << nbits) | self.peek_bits(nbits);
        self.consume(nbits);
        val
    }

    /// Reads one bit.
    pub fn read_bit(&mut self) -> u64 {
        if self.avail_bit_count > 0 {
            let v = self.bit_buf >> 63;
            self.bit_buf <<= 1;
            self.avail_bit_count -= 1;
            return v;
        }
        let val = self.peek_bits(1);
        self.consume(1);
        val
    }

    /// Reads signed compact varint.
    pub fn read_varint_compact(&mut self) -> i64 {
        let x = self.read_uvarint_compact();
        ((x >> 1) as i64) ^ (-((x & 1) as i64))
    }

    /// Reads unsigned compact varint.
    pub fn read_uvarint_compact(&mut self) -> u64 {
        let val = self.peek_bits(56);
        let zeros = val.leading_zeros() as usize;
        let ret = (val >> READ_SHIFT_BY_ZEROS[zeros]) & READ_MASK_BY_ZEROS[zeros];
        self.consume(READ_CONSUME_COUNT_BY_ZEROS[zeros]);
        ret
    }
}
