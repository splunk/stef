package encoders

import "github.com/splunk/stef/go/pkg"

type Uint64Encoder struct {
	buf     pkg.BytesWriter
	limiter *pkg.SizeLimiter
}

func (e *Uint64Encoder) Init(limiter *pkg.SizeLimiter, columns *pkg.WriteColumnSet) error {
	e.limiter = limiter
	return nil
}

func (e *Uint64Encoder) Reset() {
}

func (e *Uint64Encoder) Encode(val uint64) {
	oldLen := len(e.buf.Bytes())
	e.buf.WriteUvarint(val)

	newLen := len(e.buf.Bytes())
	e.limiter.AddFrameBytes(uint(newLen - oldLen))
}

func (e *Uint64Encoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBytes(&e.buf)
}

func (e *Uint64Encoder) WriteTo(buf *pkg.BytesWriter) {
	buf.WriteBytes(e.buf.Bytes())
}

type Uint64Decoder struct {
	buf    pkg.BytesReader
	column *pkg.ReadableColumn
}

func (d *Uint64Decoder) Continue() {
	d.buf.Reset(d.column.Data())
}

func (d *Uint64Decoder) Decode(dst *uint64) (err error) {
	*dst, err = d.buf.ReadUvarint()
	return err
}

func (d *Uint64Decoder) Reset() {
}

func (d *Uint64Decoder) Init(columns *pkg.ReadColumnSet) error {
	d.column = columns.Column()
	return nil
}
