package pkg

import "errors"

type ReadOptions struct {
	// TillAvailable indicates that Read() operation must only succeed if
	// the record is available, i.e. is part of the already fetched frame.
	// If Read() is attempted when there are no more record remaining in the
	// current frame then ErrEndOfFrame is returned by Read().
	// If TillAvailable is false then Read() will fetch new frames from the
	// underlying reader as needed.
	TillAvailable bool
}

var ErrEndOfFrame = errors.New("end of frame")
