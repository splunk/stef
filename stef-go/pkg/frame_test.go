package pkg

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type memChunkReaderWriter struct {
	buf bytes.Buffer
}

func (m *memChunkReaderWriter) ReadByte() (byte, error) {
	return m.buf.ReadByte()
}

func (m *memChunkReaderWriter) Read(p []byte) (n int, err error) {
	return m.buf.Read(p)
}

func (m *memChunkReaderWriter) WriteChunk(header []byte, content []byte) error {
	_, err := m.buf.Write(header)
	if err != nil {
		return err
	}
	_, err = m.buf.Write(content)
	return err
}

func (m *memChunkReaderWriter) Bytes() []byte {
	return m.buf.Bytes()
}

func TestLastFrameAndContinue(t *testing.T) {
	// This test verifies that it is possible to decode until the end of available
	// data, get a correct indication that it is the end of the frame and end
	// of all available data, then once new data becomes available the decoding
	// can continue successfully from the newly added data.
	// The continuation is only possible at the frame boundary.

	// Encode one frame with some data.
	encoder := FrameEncoder{}
	buf := &memChunkReaderWriter{}
	err := encoder.Init(buf, CompressionZstd)
	require.NoError(t, err)
	writeStr := []byte(strings.Repeat("hello", 10))
	_, err = encoder.Write(writeStr)
	require.NoError(t, err)

	err = encoder.CloseFrame()
	require.NoError(t, err)

	// Now decode that frame.
	decoder := FrameDecoder{}
	err = decoder.Init(buf, CompressionZstd)
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

	// Try decoding the next frame and make sure we get the EOF from the source byte Reader.
	_, err = decoder.Next()
	require.ErrorIs(t, err, io.EOF)

	// Continue adding to the same source byte buffer using encoder.

	// Open a new frame, write new data and close the frame.
	encoder.OpenFrame(0)
	writeStr = []byte(strings.Repeat("foo", 10))
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
