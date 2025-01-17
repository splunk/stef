package encoders

import (
	"errors"
	"unsafe"

	"github.com/splunk/stef/go/pkg"
)

type StringEncoder struct {
	buf     pkg.BytesWriter
	dict    *StringEncoderDict
	limiter *pkg.SizeLimiter
}

type StringEncoderDict struct {
	m       map[string]int
	limiter *pkg.SizeLimiter
}

func (e *StringEncoderDict) Init(limiter *pkg.SizeLimiter) {
	e.m = make(map[string]int)
	e.limiter = limiter
}

func (e *StringEncoderDict) Reset() {
	e.m = make(map[string]int)
}

func (e *StringEncoder) Init(dict *StringEncoderDict, limiter *pkg.SizeLimiter, columns *pkg.WriteColumnSet) error {
	e.dict = dict
	e.limiter = limiter
	return nil
}

func (e *StringEncoder) Encode(val string) {
	oldLen := len(e.buf.Bytes())
	if e.dict != nil {
		if refNum, exists := e.dict.m[val]; exists {
			e.buf.WriteVarint(int64(-refNum - 1))
			newLen := len(e.buf.Bytes())
			e.dict.limiter.AddFrameBytes(uint(newLen - oldLen))
			return
		}
	}
	strLen := len(val)
	if e.dict != nil && strLen > 1 {
		refNum := len(e.dict.m)
		e.dict.m[val] = refNum
		e.dict.limiter.AddDictElemSize(uint(strLen) + uint(unsafe.Sizeof(val)))
	}
	e.buf.WriteVarint(int64(strLen))
	e.buf.WriteStringBytes(val)
	newLen := len(e.buf.Bytes())
	e.limiter.AddFrameBytes(uint(newLen - oldLen))
}

func (e *StringEncoder) CollectColumns(columnSet *pkg.WriteColumnSet) {
	columnSet.SetBytes(&e.buf)
}

func (e *StringEncoder) WriteTo(buf *pkg.BytesWriter) {
	buf.WriteBytes(e.buf.Bytes())
}

func (e *StringEncoder) Reset() {
}

type StringDecoder struct {
	buf    pkg.BytesReader
	column *pkg.ReadableColumn
	dict   *StringDecoderDict
}

type StringDecoderDict struct {
	dict []string
}

func (d *StringDecoderDict) Init() {
}

var ErrInvalidRefNum = errors.New("invalid RefNum, out of dictionary range")

func (d *StringDecoder) Continue() {
	d.buf.Reset(d.column.Data())
}

func (d *StringDecoder) Reset() {
}

func (d *StringDecoder) Decode(dst *string) error {
	varint, err := d.buf.ReadVarint()
	if err != nil {
		return err
	}

	if varint >= 0 {
		strLen := int(varint)
		*dst, err = d.buf.ReadStringBytes(strLen)
		if err != nil {
			return err
		}
		if strLen > 1 && d.dict != nil {
			d.dict.dict = append(d.dict.dict, *dst)
		}
		return nil
	}
	if d.dict == nil {
		return ErrInvalidRefNum
	}
	refNum := int(-varint - 1)
	if refNum >= len(d.dict.dict) {
		return ErrInvalidRefNum
	}
	*dst = d.dict.dict[refNum]
	return nil
}

func (d *StringDecoder) Init(dict *StringDecoderDict, columns *pkg.ReadColumnSet) error {
	d.dict = dict
	d.column = columns.Column()
	return nil
}
