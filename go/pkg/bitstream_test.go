package pkg

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBit(t *testing.T) {
	bw := NewBitsWriter(0)

	for i := 0; i < 11; i++ {
		bw.WriteBits(1, 1)
	}
	bw.Close()
	require.EqualValues(t, []byte{0b11111111, 0b11100000}, bw.stream)
}

func TestIncreasingWriteReadBits(t *testing.T) {
	bw := NewBitsWriter(0)

	const Count = 0x1000000
	for i := uint64(1); i <= Count; i += 111 {
		v := i
		bitCount := uint(math.Floor(math.Log2(float64(v)))) + 1
		bw.WriteBits(v, bitCount)
	}
	bw.Close()

	br := NewBitsReader()
	br.Reset(bw.Bytes())

	for i := uint64(1); i <= Count; i += 111 {
		v := i
		bitCount := uint(math.Floor(math.Log2(float64(v)))) + 1
		val := br.ReadBits(bitCount)
		require.Equalf(t, v, val, "%v", i)
	}
}

func TestRandWriteReadBits(t *testing.T) {
	bw := NewBitsWriter(0)

	const Count = 0x10000

	random := rand.New(rand.NewSource(0))

	for i := uint64(1); i <= Count; i++ {
		shift := random.Intn(64)
		v := random.Uint64() >> shift
		var bitCount uint
		if v == 0 {
			bitCount = 0
		} else {
			bitCount = uint(math.Floor(math.Log2(float64(v)))) + 1
		}
		bw.WriteBits(v, bitCount)
	}
	bw.Close()

	br := NewBitsReader()
	br.Reset(bw.Bytes())

	random = rand.New(rand.NewSource(0))

	for i := uint64(1); i <= Count; i++ {
		shift := random.Intn(64)
		v := random.Uint64() >> shift
		var bitCount uint
		if v == 0 {
			bitCount = 0
		} else {
			bitCount = uint(math.Floor(math.Log2(float64(v)))) + 1
		}
		val := br.ReadBits(bitCount)
		require.Equal(t, v, val)
	}
}

func BenchmarkBstreamWriteBit(b *testing.B) {
	bw := NewBitsWriter(1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < 1000000; j++ {
			bw.WriteBit(uint(j % 2))
		}
	}
}

func BenchmarkBstreamReadBit(b *testing.B) {
	bw := NewBitsWriter(0)
	for j := 0; j < 1000000; j++ {
		bw.WriteBit(uint(j % 2))
	}
	bw.Close()
	byts := bw.Bytes()
	br := NewBitsReader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Reset(byts)
		for j := 0; j < 1000000; j++ {
			v := br.ReadBit()
			if v != uint64(j%2) {
				panic("invalid value")
			}
		}
	}
}

func BenchmarkBstreamReadBits(b *testing.B) {
	bw := NewBitsWriter(0)

	val := uint64(1)
	for j := uint(1); j < 64; j++ {
		bw.WriteBits(val, j)
		val *= 2
	}
	bw.Close()

	br := NewBitsReader()

	for i := 0; i < b.N; i++ {
		br.Reset(bw.Bytes())

		val = 1
		for j := uint(1); j < 64; j++ {
			v := br.ReadBits(j)
			if v != val {
				panic("mismatch")
			}
			val *= 2
		}
	}
}

func BenchmarkBstreamWriteUvarintCompactSmall(b *testing.B) {
	bw := NewBitsWriter(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < 47; j++ {
			bw.WriteUvarintCompact(uint64(j))
		}
	}
}

func BenchmarkBstreamReadUvarintCompactSmall(b *testing.B) {
	bw := NewBitsWriter(0)
	for j := 0; j < 47; j++ {
		bw.WriteUvarintCompact(uint64(j))
	}
	bw.Close()
	byts := bw.Bytes()
	br := NewBitsReader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Reset(byts)
		for j := 0; j < 47; j++ {
			v := br.ReadUvarintCompact()
			if v != uint64(j) {
				panic("invalid value")
			}
		}
	}
}

func BenchmarkBstreamWriteUvarintCompact(b *testing.B) {
	bw := NewBitsWriter(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < 47; j++ {
			bw.WriteUvarintCompact(uint64(1 << j))
		}
	}
}

func BenchmarkBstreamReadUvarintCompact(b *testing.B) {
	bw := NewBitsWriter(0)
	for j := 0; j < 47; j++ {
		bw.WriteUvarintCompact(uint64(1 << j))
	}
	bw.Close()
	byts := bw.Bytes()
	br := NewBitsReader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Reset(byts)
		for j := 0; j < 47; j++ {
			v := br.ReadUvarintCompact()
			if v != uint64(1<<j) {
				panic("invalid value")
			}
		}
	}
}
