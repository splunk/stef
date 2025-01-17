package pkg

type FrameFlags byte

const (
	// RestartDictionaries resets and restarts all dictionaries at frame beginning.
	RestartDictionaries FrameFlags = 1 << iota
	// RestartCompression resets and restarts the compression stream at frame beginning.
	RestartCompression

	FrameFlagsMask = RestartDictionaries | RestartCompression
)
