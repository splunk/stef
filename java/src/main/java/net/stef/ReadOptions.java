package net.stef;

// ReadOptions is a bitmask of options.
public class ReadOptions {
    // TillEndOfFrame indicates that the read operation must only succeed if
    // the record is available in the already fetched frame. This ensures
    // that the Reader will not attempt to fetch new frames from the underlying
    // source reader and guarantees that it will not block on I/O.
    // If read is attempted when there are no more records remaining in the
    // current frame, then ErrEndOfFrame is returned.
    // If TillEndOfFrame is false, then read will fetch new frames from the
    // underlying reader as needed.
    public static int tillEndOfFrame = 1;

    // Default read options to use.
    public static int none = 0;
}