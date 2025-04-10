package encoders

import (
	"bytes"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg"
)

func TestFloat64(t *testing.T) {
	encoder := &Float64Encoder{}
	decoder := &Float64Decoder{}

	err := encoder.Init(&pkg.SizeLimiter{}, nil)
	require.NoError(t, err)

	// Use values that will ensure all encoding/decoding paths are tested
	values := []float64{1, 1, 2, 1}
	for _, v := range values {
		encoder.Encode(v)
	}

	wb := pkg.WriteBufs{}
	encoder.CollectColumns(&wb.Columns)

	buf := bytes.NewBuffer(nil)
	err = wb.WriteTo(buf)
	require.NoError(t, err)

	rb := pkg.ReadBufs{}
	err = rb.ReadFrom(buf)
	require.NoError(t, err)

	err = decoder.Init(&rb.Columns)
	require.NoError(t, err)
	decoder.Continue()

	for _, v := range values {
		var decodedVal float64
		err = decoder.Decode(&decodedVal)
		require.NoError(t, err)
		assert.Equal(t, v, decodedVal)
	}
}

func TestFloat64Random(t *testing.T) {
	encoder := &Float64Encoder{}
	decoder := &Float64Decoder{}

	err := encoder.Init(&pkg.SizeLimiter{}, nil)
	require.NoError(t, err)

	// Choose a seed (non-pseudo) randomly. We will print the seed
	// on failure for easy reproduction.
	seed1 := uint64(time.Now().UnixNano())
	random := rand.New(rand.NewPCG(seed1, 0))

	// Use values that will ensure all encoding/decoding paths are tested
	var values []float64
	for i := 0; i < 1000; i++ {
		values = append(values, random.Float64())
	}

	for _, v := range values {
		encoder.Encode(v)
	}

	wb := pkg.WriteBufs{}
	encoder.CollectColumns(&wb.Columns)

	buf := bytes.NewBuffer(nil)
	err = wb.WriteTo(buf)
	require.NoError(t, err, seed1)

	rb := pkg.ReadBufs{}
	err = rb.ReadFrom(buf)
	require.NoError(t, err, seed1)

	err = decoder.Init(&rb.Columns)
	require.NoError(t, err, seed1)
	decoder.Continue()

	for _, v := range values {
		var decodedVal float64
		err = decoder.Decode(&decodedVal)
		require.NoError(t, err, seed1)
		assert.Equal(t, v, decodedVal, seed1)
	}
}
