package pkg

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testLastFrameAndContinue(t *testing.T, compression Compression) {
	// This test verifies that it is possible to decode until the end of available
	// data, get a correct indication that it is the end of the frame and end
	// of all available data, then once new data becomes available the decoding
	// can continue successfully from the newly added data.
	// The continuation is only possible at the frame boundary.

	// Encode one frame with some data.
	encoder := FrameEncoder{}
	buf := &MemReaderWriter{}
	err := encoder.Init(buf, compression)
	require.NoError(t, err)
	defer encoder.Close()
	writeStr := []byte(strings.Repeat("hello", 10))
	_, err = encoder.Write(writeStr)
	require.NoError(t, err)

	err = encoder.CloseFrame()
	require.NoError(t, err)

	// Now decode that frame.
	decoder := FrameDecoder{}
	err = decoder.Init(buf, compression)
	require.NoError(t, err)
	_, err = decoder.Next()
	require.NoError(t, err)

	readStr := make([]byte, len(writeStr))
	n, err := decoder.Read(readStr)
	require.NoError(t, err)
	require.EqualValues(t, len(writeStr), n)
	require.EqualValues(t, writeStr, readStr)

	// Try decoding more, past the end of frame.
	n, err = decoder.Read(readStr)

	// Make sure the error indicates end of the frame.
	require.ErrorIs(t, err, EndOfFrame)
	require.EqualValues(t, 0, n)

	for i := 1; i <= 10; i++ {
		// Try decoding the next frame and make sure we get the EOF from the source byte Reader.
		_, err = decoder.Next()
		require.ErrorIs(t, err, io.EOF)

		// Continue adding to the same source byte buffer using encoder.

		// Open a new frame, write new data and close the frame.
		encoder.OpenFrame(0)
		writeStr = []byte(strings.Repeat("foo", i))
		_, err = encoder.Write(writeStr)
		require.NoError(t, err)

		err = encoder.CloseFrame()
		require.NoError(t, err)

		// Try reading again. We should get an EndOfFrame error.
		readStr = make([]byte, len(writeStr))
		n, err = decoder.Read(readStr)
		require.ErrorIs(t, err, EndOfFrame)
		require.EqualValues(t, 0, n)

		// Now try decoding a new frame. This time it should succeed since we added a new frame.
		_, err = decoder.Next()
		require.NoError(t, err)

		// Read the encoded data.
		n, err = decoder.Read(readStr)
		require.EqualValues(t, len(writeStr), n)
		require.EqualValues(t, writeStr, readStr)

		// Try decoding more, past the end of second frame.
		n, err = decoder.Read(readStr)

		// Make sure the error indicates end of the frame.
		require.ErrorIs(t, err, EndOfFrame)
		require.EqualValues(t, 0, n)
	}
}

func TestLastFrameAndContinue(t *testing.T) {
	compressions := []Compression{
		CompressionNone,
		CompressionZstd,
	}

	for _, compression := range compressions {
		t.Run(
			strconv.Itoa(int(compression)), func(t *testing.T) {
				testLastFrameAndContinue(t, compression)
			},
		)
	}
}

func BenchmarkFrameEncoderZstd(b *testing.B) {
	data := []byte(strings.Repeat("hello world this is benchmark data ", 100))

	b.Run("pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := &MemReaderWriter{}
			encoder := FrameEncoder{}
			err := encoder.Init(buf, CompressionZstd)
			if err != nil {
				b.Fatal(err)
			}
			_, err = encoder.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			err = encoder.CloseFrame()
			if err != nil {
				b.Fatal(err)
			}
			encoder.Close()
		}
	})
}

func TestLimitedReader(t *testing.T) {
	data := []byte("abcdef")
	mem := &MemReaderWriter{buf: *bytes.NewBuffer(data)}
	var lr limitedReader
	lr.Init(mem)

	// Test reading with limit 0
	lr.limit = 0
	buf := make([]byte, 3)
	n, err := lr.Read(buf)
	require.Equal(t, 0, n)
	require.ErrorIs(t, err, io.EOF)

	// Test ReadByte with limit 0
	lr.limit = 0
	_, err = lr.ReadByte()
	require.ErrorIs(t, err, io.EOF)

	// Reset and test reading less than limit
	mem = &MemReaderWriter{buf: *bytes.NewBuffer(data)}
	lr.Init(mem)
	lr.limit = 3
	buf = make([]byte, 2)
	n, err = lr.Read(buf)
	require.Equal(t, 2, n)
	require.NoError(t, err)
	require.Equal(t, []byte("ab"), buf)
	require.Equal(t, int64(1), lr.limit)

	// Test ReadByte with remaining limit
	b, err := lr.ReadByte()
	require.NoError(t, err)
	require.Equal(t, byte('c'), b)
	require.Equal(t, int64(0), lr.limit)

	// Test ReadByte at limit 0 after reading
	_, err = lr.ReadByte()
	require.ErrorIs(t, err, io.EOF)

	// Test reading more than limit
	mem = &MemReaderWriter{buf: *bytes.NewBuffer(data)}
	lr.Init(mem)
	lr.limit = 4
	buf = make([]byte, 10)
	n, err = lr.Read(buf)
	require.Equal(t, 4, n)
	require.NoError(t, err)
	require.Equal(t, []byte("abcd"), buf[:n])
	require.Equal(t, int64(0), lr.limit)

	// Test Read after limit exhausted
	n, err = lr.Read(buf)
	require.Equal(t, 0, n)
	require.ErrorIs(t, err, io.EOF)
}
