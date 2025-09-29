package encoders

type Int64DeltaDeltaEncoder struct {
	Uint64DeltaDeltaEncoder
}

func (e *Int64DeltaDeltaEncoder) Encode(val int64) {
	e.Uint64DeltaDeltaEncoder.Encode(uint64(val))
}

type Int64DeltaDeltaDecoder struct {
	Uint64DeltaDeltaDecoder
}

func (d *Int64DeltaDeltaDecoder) Decode(dst *int64) error {
	tsDeltaOfDelta, err := d.buf.ReadVarint()
	if err != nil {
		return err
	}

	tsDelta := d.lastDelta + uint64(tsDeltaOfDelta)
	d.lastDelta = tsDelta

	d.lastVal += tsDelta

	*dst = int64(d.lastVal)
	return nil
}
