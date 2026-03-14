/// Flags attached to each frame header.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub struct FrameFlags(pub u8);

impl FrameFlags {
    /// Restart all dictionaries at frame start.
    pub const RESTART_DICTIONARIES: u8 = 1 << 0;
    /// Restart compression stream at frame start.
    pub const RESTART_COMPRESSION: u8 = 1 << 1;
    /// Reset codec state at frame start.
    pub const RESTART_CODECS: u8 = 1 << 2;
    /// Valid flags mask.
    pub const MASK: u8 = Self::RESTART_DICTIONARIES | Self::RESTART_COMPRESSION | Self::RESTART_CODECS;

    /// Returns true when all set bits are valid frame flags.
    pub fn is_valid(self) -> bool {
        (self.0 | Self::MASK) == Self::MASK
    }

    /// Returns true when a specific flag bit is set.
    pub fn has(self, flag: u8) -> bool {
        self.0 & flag != 0
    }
}
