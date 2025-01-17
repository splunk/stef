package encoders

import (
	"github.com/splunk/stef/go/pkg"
)

type BytesEncoder struct {
	StringEncoder
}

type BytesEncoderDict StringEncoderDict

func (e *BytesEncoderDict) Init(limiter *pkg.SizeLimiter) {
	(*StringEncoderDict)(e).Init(limiter)
}

func (e *BytesEncoderDict) Reset() {
	(*StringEncoderDict)(e).Reset()
}

func (e *BytesEncoder) Init(dict *BytesEncoderDict, limiter *pkg.SizeLimiter, columns *pkg.WriteColumnSet) error {
	return e.StringEncoder.Init((*StringEncoderDict)(dict), limiter, columns)
}

func (e *BytesEncoder) Encode(val pkg.Bytes) {
	e.StringEncoder.Encode(string(val))
}

type BytesDecoder struct {
	StringDecoder
}

type BytesDecoderDict StringDecoderDict

func (d *BytesDecoderDict) Init() {
	(*StringDecoderDict)(d).Init()
}

func (d *BytesDecoder) Continue() {
	d.StringDecoder.Continue()
}

func (d *BytesDecoder) Decode(dst *pkg.Bytes) error {
	return d.StringDecoder.Decode((*string)(dst))
}

func (d *BytesDecoder) Init(dict *BytesDecoderDict, columns *pkg.ReadColumnSet) error {
	return d.StringDecoder.Init((*StringDecoderDict)(dict), columns)
}
