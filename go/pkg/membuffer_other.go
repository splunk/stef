//go:build !amd64 || !goexperiment.simd

package pkg

// ReadUvar64x4 reads 4 variable length integers using SIMD operations when available
func (r *BytesReader) ReadUvar64x4() ([4]uint64, error) {
	return r.readUvar64x4Scalar()
}

// ReadUvar32x4 reads 4 variable length integers using SIMD operations when available
func (r *BytesReader) ReadUvar32x4() ([4]uint32, error) {
	return r.readUvar32x4Scalar()
}

func (w *BytesWriter) WriteUvar32x4(values [4]uint32) {
	w.writeUvar32x4Scalar(values)
}
