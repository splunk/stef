package encoders

type Int64DeltaEncoder struct {
	Uint64DeltaEncoder
}

func (e *Int64DeltaEncoder) Encode(val int64) {
	e.Uint64DeltaEncoder.Encode(uint64(val))
}

type Int64DeltaDecoder struct {
	Uint64DeltaDecoder
}

func (d *Int64DeltaDecoder) Decode(dst *int64) error {
	tsDelta, err := d.buf.ReadVarint()
	if err != nil {
		return err
	}

	d.lastVal += uint64(tsDelta)

	*dst = int64(d.lastVal)
	return nil
}
