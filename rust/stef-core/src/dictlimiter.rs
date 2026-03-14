use crate::writeropts::WriterOptions;

/// Tracks frame and dictionary growth against configured limits.
#[derive(Debug, Default, Clone)]
pub struct SizeLimiter {
    dict_byte_size: usize,
    dict_byte_size_limit: usize,
    dict_size_limit_reached: bool,
    frame_bit_size: usize,
    frame_bit_size_limit: usize,
}

impl SizeLimiter {
    /// Initializes limiter from writer options.
    pub fn init(&mut self, opts: &WriterOptions) {
        self.dict_byte_size = 0;
        self.frame_bit_size = 0;
        self.dict_byte_size_limit = opts.max_total_dict_size;
        self.frame_bit_size_limit = opts.max_uncompressed_frame_byte_size.saturating_mul(8);
        self.dict_size_limit_reached = false;
    }

    /// Accounts dictionary element byte size.
    pub fn add_dict_elem_size(&mut self, elem_byte_size: usize) {
        if self.dict_byte_size_limit != 0 {
            self.dict_byte_size = self.dict_byte_size.saturating_add(elem_byte_size);
            if self.dict_byte_size >= self.dict_byte_size_limit {
                self.dict_size_limit_reached = true;
            }
        }
    }

    /// Accounts frame bits.
    pub fn add_frame_bits(&mut self, bit_count: usize) {
        self.frame_bit_size = self.frame_bit_size.saturating_add(bit_count);
    }

    /// Accounts frame bytes.
    pub fn add_frame_bytes(&mut self, byte_count: usize) {
        self.add_frame_bits(byte_count.saturating_mul(8));
    }

    /// Returns true when dictionary limit is reached.
    pub fn dict_limit_reached(&self) -> bool {
        self.dict_size_limit_reached
    }

    /// Returns true when frame size limit is reached.
    pub fn frame_limit_reached(&self) -> bool {
        self.frame_bit_size_limit != 0 && self.frame_bit_size >= self.frame_bit_size_limit
    }

    /// Resets dictionary accounting.
    pub fn reset_dict(&mut self) {
        self.dict_byte_size = 0;
        self.dict_size_limit_reached = false;
    }

    /// Resets frame size accounting.
    pub fn reset_frame_size(&mut self) {
        self.frame_bit_size = 0;
    }
}
