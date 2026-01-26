package pkg

var ErrMultimap = NewDecodeError("invalid multimap")
var ErrMultimapCountLimit = NewDecodeError("too many elements in the multimap")
var ErrInvalidRefNum = NewDecodeError("invalid refNum")
var ErrInvalidOneOfType = NewDecodeError("invalid oneof type")

var ErrInvalidHeader = NewDecodeError("invalid FixedHeader")
var ErrInvalidHeaderSignature = NewDecodeError("invalid FixedHeader signature")
var ErrInvalidFormatVersion = NewDecodeError("invalid format version in the FixedHeader")
var ErrInvalidCompression = NewDecodeError("invalid compression method")

var ErrInvalidVarHeader = NewDecodeError("invalid VarHeader")

var ErrFrameSizeLimit = NewDecodeError("frame is too large")

var ErrColumnSizeLimitExceeded = NewDecodeError("column size limit exceeded")
var ErrTotalColumnSizeLimitExceeded = NewDecodeError("total column size limit exceeded")

var ErrRecordAllocLimitExceeded = NewDecodeError("record allocation limit exceeded")

var ErrTooManyFieldsToDecode = NewDecodeError("too many fields to decode")
var ErrEmptyRootStructDisallowed = NewDecodeError("cannot decode empty root struct")

type DecodeError struct {
	msg string
}

func (e *DecodeError) Error() string {
	return e.msg
}

func NewDecodeError(msg string) error {
	return &DecodeError{msg: msg}
}
