package pkg

import (
	"strconv"
	"testing"
)

func BenchmarkMembufWriteVaruint(b *testing.B) {
	//b.Skip()
	bw := NewBytesWriter(10000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < 1000000; j++ {
			bw.WriteUvarint(uint64(j))
		}
	}
}

func BenchmarkMembufReadVaruintExp(b *testing.B) {
	bw := BytesWriter{}
	val := uint64(1)
	for j := 0; j < 63; j++ {
		bw.WriteUvarint(val)
		val *= 2
	}
	b.ResetTimer()
	br := BytesReader{buf: bw.buf}
	for i := 0; i < b.N; i++ {
		br.byteIndex = 0
		checkVal := uint64(1)
		for j := 0; j < 63; j++ {
			val, err := br.ReadUvarint()
			if val != checkVal || err != nil {
				panic(nil)
			}
			checkVal *= 2
		}
	}
}

func BenchmarkMembufWriteVaruintSizes(b *testing.B) {
	for size := 1; size <= 9; size++ {
		val := uint64((1 << (size * 7)) - 1)
		b.Run(
			strconv.Itoa(size), func(b *testing.B) {
				bw := BytesWriter{buf: make([]byte, 0, 1000)}
				for i := 0; i < b.N; i++ {
					bw.Reset()
					for j := 0; j < 1000; j++ {
						bw.WriteUvarint(val)
					}
				}
			},
		)
	}
}
