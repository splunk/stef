package encoders

type Int64Encoder struct {
	Uint64Encoder
}

func (e *Int64Encoder) Encode(val int64) {
	oldLen := len(e.buf.Bytes())
	e.buf.WriteVarint(val)
	e.limiter.AddFrameBytes(uint(len(e.buf.Bytes()) - oldLen))
}

type Int64Decoder struct {
	Uint64Decoder
}

func (d *Int64Decoder) Decode(dst *int64) (err error) {
	*dst, err = d.buf.ReadVarint()
	return err
}
