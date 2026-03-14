/// Read options controlling frame-boundary behavior.
#[derive(Debug, Clone, Copy, Default)]
pub struct ReadOptions {
    /// If true, record reads are restricted to currently loaded frame.
    pub till_end_of_frame: bool,
}

/// End-of-frame sentinel error.
#[derive(Debug, thiserror::Error)]
#[error("end of frame")]
pub struct ErrEndOfFrame;
