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
	const recCount = 1000000
	bw := NewBitsWriter(recCount)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < recCount; j++ {
			bw.WriteBit(uint(j % 2))
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkBstreamReadBit(b *testing.B) {
	const recCount = 1000000
	bw := NewBitsWriter(0)
	for j := 0; j < recCount; j++ {
		bw.WriteBit(uint(j % 2))
	}
	bw.Close()
	byts := bw.Bytes()
	br := NewBitsReader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Reset(byts)
		for j := 0; j < recCount; j++ {
			v := br.ReadBit()
			if v != uint64(j%2) {
				panic("invalid value")
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkBstreamReadBits(b *testing.B) {
	const recCount = 63

	bw := NewBitsWriter(0)

	val := uint64(1)
	for j := uint(1); j <= recCount; j++ {
		bw.WriteBits(val, j)
		val *= 2
	}
	bw.Close()

	br := NewBitsReader()

	for i := 0; i < b.N; i++ {
		br.Reset(bw.Bytes())

		val = 1
		for j := uint(1); j <= recCount; j++ {
			v := br.ReadBits(j)
			if v != val {
				panic("mismatch")
			}
			val *= 2
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkBstreamWriteUvarintCompactSmall(b *testing.B) {
	const recCount = 47

	bw := NewBitsWriter(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < recCount; j++ {
			bw.WriteUvarintCompact(uint64(j))
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkBstreamReadUvarintCompactSmall(b *testing.B) {
	const recCount = 47

	bw := NewBitsWriter(0)
	for j := 0; j < recCount; j++ {
		bw.WriteUvarintCompact(uint64(j))
	}
	bw.Close()
	byts := bw.Bytes()
	br := NewBitsReader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Reset(byts)
		for j := 0; j < recCount; j++ {
			v := br.ReadUvarintCompact()
			if v != uint64(j) {
				panic("invalid value")
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkBstreamWriteUvarintCompact(b *testing.B) {
	const recCount = 47

	bw := NewBitsWriter(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < recCount; j++ {
			bw.WriteUvarintCompact(uint64(1 << j))
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkBstreamReadUvarintCompact(b *testing.B) {
	const recCount = 47

	bw := NewBitsWriter(0)
	for j := 0; j < recCount; j++ {
		bw.WriteUvarintCompact(uint64(1 << j))
	}
	bw.Close()
	byts := bw.Bytes()
	br := NewBitsReader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Reset(byts)
		for j := 0; j < recCount; j++ {
			v := br.ReadUvarintCompact()
			if v != uint64(1<<j) {
				panic("invalid value")
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}
