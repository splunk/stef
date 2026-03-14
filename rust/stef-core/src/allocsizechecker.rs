use crate::errors::{ERR_RECORD_ALLOC_LIMIT_EXCEEDED, Result, StefError};
use crate::limits::RECORD_ALLOC_LIMIT;

/// Tracks allocation volume during decoding to prevent memory abuse.
#[derive(Debug, Default, Clone)]
pub struct AllocSizeChecker {
    allocated_size: usize,
}

impl AllocSizeChecker {
    /// Returns currently accumulated allocation size.
    pub fn allocated_size(&self) -> usize {
        self.allocated_size
    }

    /// Resets accumulated allocation size.
    pub fn reset_alloc_size(&mut self) {
        self.allocated_size = 0;
    }

    /// Adds size and saturates to `usize::MAX` on overflow.
    pub fn add_alloc_size(&mut self, size: usize) {
        self.allocated_size = self.allocated_size.saturating_add(size);
    }

    /// Returns `true` when allocation exceeds `RECORD_ALLOC_LIMIT`.
    pub fn is_over_limit(&self) -> bool {
        self.allocated_size > RECORD_ALLOC_LIMIT
    }

    /// Checks and applies one allocation.
    pub fn prep_alloc_size(&mut self, size: usize) -> Result<()> {
        if let Some(v) = self.allocated_size.checked_add(size) {
            self.allocated_size = v;
            if self.is_over_limit() {
                return Err(ERR_RECORD_ALLOC_LIMIT_EXCEEDED);
            }
            Ok(())
        } else {
            self.allocated_size = usize::MAX;
            Err(ERR_RECORD_ALLOC_LIMIT_EXCEEDED)
        }
    }

    /// Checks and applies `size * count` allocation.
    pub fn prep_alloc_size_n(&mut self, size: usize, count: usize) -> Result<()> {
        let total = size.checked_mul(count).ok_or_else(|| {
            self.allocated_size = usize::MAX;
            StefError::decode("record allocation limit exceeded")
        })?;
        self.prep_alloc_size(total)
    }
}
