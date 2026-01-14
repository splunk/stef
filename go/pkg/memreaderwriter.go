package pkg

import "bytes"

// MemReaderWriter is an in-memory implementation of ChunkWriter and ByteAndBlockReader interfaces
// that allows to first write to the buffer and then read from it.
// Typically used in tests.
type MemReaderWriter struct {
	buf bytes.Buffer
}

func (m *MemReaderWriter) ReadByte() (byte, error) {
	return m.buf.ReadByte()
}

func (m *MemReaderWriter) Read(p []byte) (n int, err error) {
	return m.buf.Read(p)
}

func (m *MemReaderWriter) WriteChunk(header []byte, content []byte) error {
	if _, err := m.buf.Write(header); err != nil {
		return err
	}
	_, err := m.buf.Write(content)
	return err
}

func (m *MemReaderWriter) Bytes() []byte {
	return m.buf.Bytes()
}
