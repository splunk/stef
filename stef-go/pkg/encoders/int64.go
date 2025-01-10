package encoders

type Int64Encoder struct {
	Uint64Encoder
}

func (e *Int64Encoder) IsEqual(val int64) bool {
	return e.Uint64Encoder.IsEqual(uint64(val))
}

func (e *Int64Encoder) Encode(val int64) {
	e.Uint64Encoder.Encode(uint64(val))
}

type Int64Decoder struct {
	Uint64Decoder
	tmpVal uint64
}

func (e *Int64Decoder) Decode(dst *int64) error {
	err := e.Uint64Decoder.Decode(&e.tmpVal)
	*dst = int64(e.tmpVal)
	return err
}
