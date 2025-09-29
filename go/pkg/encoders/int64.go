package encoders

type Int64Encoder struct {
	Uint64Encoder
}

func (e *Int64Encoder) Encode(val int64) {
	oldLen := len(e.buf.Bytes())
	e.buf.WriteVarint(val)

	newLen := len(e.buf.Bytes())
	e.limiter.AddFrameBytes(uint(newLen - oldLen))
}

type Int64Decoder struct {
	Uint64Decoder
}

func (d *Int64Decoder) Decode(dst *int64) error {
	var err error
	*dst, err = d.buf.ReadVarint()
	return err
}
