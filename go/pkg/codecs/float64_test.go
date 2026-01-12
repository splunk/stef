package codecs

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/splunk/stef/go/pkg"
)

func equalFloat(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return a == b
}

func verifyDecode(t *testing.T, values []float64) {
	// Encode
	var buf pkg.BitsWriter
	limiter := &pkg.SizeLimiter{}
	enc := &Float64Encoder{buf: buf, limiter: limiter}

	for _, v := range values {
		enc.Encode(v)
	}

	// Get encoded bytes
	var bytesBuf pkg.BytesWriter
	enc.buf.Close()
	enc.WriteTo(&bytesBuf)
	encoded := bytesBuf.Bytes()

	// Decode
	dec := &Float64Decoder{}
	dec.buf.Reset(encoded)

	for i, want := range values {
		var got float64
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("decode failed at %d: %v", i, err)
		}
		if !equalFloat(got, want) {
			t.Errorf(
				"mismatch at %d: got %v (bits 0x%x), want %v (bits 0x%x)", i, got, math.Float64bits(got), want,
				math.Float64bits(want),
			)
		}
	}
}

func TestFloat64EncoderDecoder_Basic(t *testing.T) {
	values := []float64{1.0, 1.0, 2.0, 2.0, 3.1415, 3.1415, -1.0, 0.0, -0.0, math.NaN(), math.Inf(1), math.Inf(-1)}
	verifyDecode(t, values)
}

func TestFloat64EncoderDecoder_Bits(t *testing.T) {
	bits := []uint64{
		0x000000000000,
		0x0FFFFFFFFFF0,
		0x0F0000000000,
		0x0F0000000F00,
		0,
		^uint64(0),
		0,
	}

	values := []float64{}
	for _, bitVal := range bits {
		values = append(values, math.Float64frombits(bitVal))
	}

	verifyDecode(t, values)
}

func TestFloat64EncoderDecoder_LeadingTrailing(t *testing.T) {
	// Values with only one bit difference, to test leading/trailing zeros logic
	base := math.Float64frombits(0x3FF0000000000000) // 1.0
	values := []float64{base}
	for i := 0; i < 64; i++ {
		values = append(values, math.Float64frombits(math.Float64bits(base)^(1<<uint(i))))
	}
	verifyDecode(t, values)
}

func randFloat(random *rand.Rand) float64 {
	// Generate random float64, including negative values
	v := random.Float64()*1e10 - 5e9 // range: [-5e9, +5e9)
	// Occasionally insert special values
	rval := random.IntN(50)
	if rval == 0 {
		v = math.NaN()
	} else if rval == 1 {
		v = math.Inf(1)
	} else if rval == 2 {
		v = math.Inf(-1)
	}
	return v
}

func TestFloat64EncoderDecoder_RandomSequence(t *testing.T) {
	randSeed := uint64(time.Now().UnixNano())
	fmt.Printf("Using random seed: %d\n", randSeed)
	random := rand.New(rand.NewPCG(randSeed, 0))

	n := 1 + random.IntN(2000)

	var buf pkg.BitsWriter
	limiter := &pkg.SizeLimiter{}
	enc := &Float64Encoder{buf: buf, limiter: limiter}

	for range n {
		v := randFloat(random)
		enc.Encode(v)
		if random.IntN(100) == 0 {
			enc.Reset()
		}
	}

	// Get encoded bytes
	var bytesBuf pkg.BytesWriter
	enc.buf.Close()
	enc.WriteTo(&bytesBuf)
	encoded := bytesBuf.Bytes()

	dec := &Float64Decoder{}
	dec.buf.Reset(encoded)

	random = rand.New(rand.NewPCG(randSeed, 0))
	random.IntN(2000)

	for i := range n {
		want := randFloat(random)
		var got float64
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("decode failed at %d: %v", i, err)
		}
		if !equalFloat(got, want) {
			t.Errorf(
				"mismatch at %d: got %v (bits 0x%x), want %v (bits 0x%x)", i, got, math.Float64bits(got), want,
				math.Float64bits(want),
			)
		}
		if random.IntN(100) == 0 {
			dec.Reset()
		}
	}
}
