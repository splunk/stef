/// Logging interface used by client/server transport components.
pub trait Logger: Send + Sync {
    /// Logs debug-level message.
    fn debugf(&self, msg: &str);
    /// Logs error-level message.
    fn errorf(&self, msg: &str);
}

/// No-op logger.
#[derive(Default)]
pub struct NopLogger;

impl Logger for NopLogger {
    fn debugf(&self, _msg: &str) {}
    fn errorf(&self, _msg: &str) {}
}
