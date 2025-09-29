package encoders

import "github.com/splunk/stef/go/pkg"

type Uint64DeltaEncoder struct {
	buf     pkg.BytesWriter
	limiter *pkg.SizeLimiter
	lastVal uint64
}

func (e *Uint64DeltaEncoder) Init(limiter *pkg.SizeLimiter, columns *pkg.WriteColumnSet) error {
	e.limiter = limiter
	return nil
}

func (e *Uint64DeltaEncoder) Reset() {
	e.lastVal = 0
}

func (e *Uint64DeltaEncoder) Encode(val uint64) {
	delta := int64(val - e.lastVal)
	e.lastVal = val

	oldLen := len(e.buf.Bytes())
	e.buf.WriteVarint(delta)

	newLen := len(e.buf.Bytes())
	e.limiter.AddFrameBytes(uint(newLen - oldLen))
}

func (e *Uint64DeltaEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBytes(&e.buf)
}

func (e *Uint64DeltaEncoder) WriteTo(buf *pkg.BytesWriter) {
	buf.WriteBytes(e.buf.Bytes())
}

type Uint64DeltaDecoder struct {
	buf     pkg.BytesReader
	column  *pkg.ReadableColumn
	lastVal uint64
}

func (d *Uint64DeltaDecoder) Continue() {
	d.buf.Reset(d.column.Data())
}

func (d *Uint64DeltaDecoder) Decode(dst *uint64) error {
	tsDelta, err := d.buf.ReadVarint()
	if err != nil {
		return err
	}

	d.lastVal += uint64(tsDelta)

	*dst = d.lastVal
	return nil
}

func (d *Uint64DeltaDecoder) Reset() {
	d.lastVal = 0
}

func (d *Uint64DeltaDecoder) Init(columns *pkg.ReadColumnSet) error {
	d.column = columns.Column()
	return nil
}
