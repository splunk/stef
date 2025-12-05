package encoders

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/splunk/stef/go/pkg"
)

func TestFloat64EncoderDecoder_Basic(t *testing.T) {
	values := []float64{0, 1.5, -2.3, math.NaN(), math.Inf(1), math.Inf(-1), 0, 1.5, -2.3}

	encoder := &Float64Encoder{}
	encoder.leadingBits = -1
	encoder.limiter = &pkg.SizeLimiter{}

	for _, v := range values {
		encoder.Encode(v)
	}
	encoder.buf.Close()

	encoded := encoder.buf.Bytes()
	decoder := &Float64Decoder{}
	decoder.buf.Reset(encoded)

	var decoded []float64
	for range values {
		var v float64
		assert.NoError(t, decoder.Decode(&v), "decode error")
		decoded = append(decoded, v)
	}

	for i, want := range values {
		got := decoded[i]
		if math.IsNaN(want) {
			assert.True(t, math.IsNaN(got), "value %d: expected NaN, got %v", i, got)
		} else if want != got {
			assert.EqualValues(t, want, got, "value %d: expected %v, got %v", i, want, got)
		}
	}
}

func TestFloat64EncoderDecoder_Random(t *testing.T) {
	const count = 10000
	seed := int64(42)
	rng := rand.New(rand.NewSource(seed))

	encoder := &Float64Encoder{}
	encoder.leadingBits = -1
	encoder.limiter = &pkg.SizeLimiter{}
	for range count {
		v := rng.NormFloat64() * 1e6
		encoder.Encode(v)
	}
	encoder.buf.Close()

	encoded := encoder.buf.Bytes()
	decoder := &Float64Decoder{}
	decoder.buf.Reset(encoded)

	// Re-generate the same sequence for comparison
	rng2 := rand.New(rand.NewSource(seed))
	for i := 0; i < count; i++ {
		var got float64
		assert.NoError(t, decoder.Decode(&got), "decode error at %d", i)
		want := rng2.NormFloat64() * 1e6
		assert.EqualValues(t, want, got)
	}
}

func TestFloat64EncoderDecoder_VerifyGivenSequence(t *testing.T) {
	values := []float64{
		0.516459, 0.516459, 0.516459, 0.026014, 0.516459, 0.026014, 0.516459, 0.026014,
		0.516459, 0.026014, 0.516459, 0.026014, 0.516459, 0.026014, 0.516459, 0.026014,
		0.516459, 0.026014, 0.516459, 0.026014, 0.516459, 0.026014, 0.516459, 0.026014,
		0.516459, 0.026014, 0.516459, 0.404796, 0.516459, 0.404796, 0.516459, 0.404796,
		0.516459, 0.404796, 0.516459, 0.404796, 0.516459, 0.404796, 0.516459, 0.404796,
		0.516459, 0.404796, 0.516459, 0.404796, 0.516459, 0.404796, 0.516459, 0.613012,
		0.681386, 0.759367, 0.909626, 0.356013, 0.265753, 0.891809, 0.482783, 0.369160,
		0.779877, 0.286262, 0.102260, 0.937321, 0.109212, 0.606182, 0.656072, 0.262938,
		0.602772, 0.820342, 0.166441, 0.107999, 0.151798, 0.034763, 0.100905, 0.673938,
		0.624203, 0.494612, 0.043941, 0.859274, 0.135444, 0.363221, 0.443968,
	}

	encoder := &Float64Encoder{}
	encoder.leadingBits = -1
	encoder.limiter = &pkg.SizeLimiter{}
	for _, v := range values {
		encoder.Encode(v)
	}
	encoder.buf.Close()

	encoded := encoder.buf.Bytes()
	decoder := &Float64Decoder{}
	decoder.leadingBits = 0
	decoder.trailingBits = 0
	decoder.lastVal = 0
	decoder.buf.Reset(encoded)

	for i, want := range values {
		var got float64
		assert.NoError(t, decoder.Decode(&got), "decode error at %d", i)
		assert.EqualValues(t, want, got)
	}
}
