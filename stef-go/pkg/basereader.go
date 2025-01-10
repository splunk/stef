package pkg

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/splunk/stef/stef-go/schema"
)

type BaseReader struct {
	Source *bufio.Reader

	FixedHeader FixedHeader
	VarHeader   VarHeader
	Schema      *schema.Schema

	ReadBufs ReadBufs

	FrameDecoder     FrameDecoder
	FrameRecordCount uint64
	RecordCount      uint64
}

func (r *BaseReader) Init(source *bufio.Reader) error {
	r.Source = source

	if err := r.ReadFixedHeader(); err != nil {
		return err
	}

	if err := r.FrameDecoder.Init(r.Source, r.FixedHeader.Compression); err != nil {
		return err
	}

	return nil
}

func (r *BaseReader) ReadFixedHeader() error {
	hdrSignature := make([]byte, len(HdrSignature))
	_, err := r.Source.Read(hdrSignature)
	if err != nil {
		return err
	}
	if string(hdrSignature) != HdrSignature {
		return ErrInvalidHeaderSignature
	}

	contentSize, err := binary.ReadUvarint(r.Source)
	if err != nil {
		return err
	}

	if contentSize < 2 || contentSize > HdrContentSizeLimit {
		return ErrInvalidHeader
	}

	hdrContent := make([]byte, contentSize)
	_, err = r.Source.Read(hdrContent)
	if err != nil {
		return err
	}

	versionAndType := hdrContent[0]
	version := versionAndType & HdrFormatVersionMask
	if version != HdrFormatVersion {
		return ErrInvalidFormatVersion
	}

	flags := hdrContent[1]

	r.FixedHeader.Compression = Compression(flags & HdrFlagsCompressionMethod)
	switch r.FixedHeader.Compression {
	case CompressionNone, CompressionZstd:
	default:
		return ErrInvalidCompression
	}

	var n int
	r.FixedHeader.TimestampMultiplier, n = binary.Uvarint(hdrContent[2:])
	if n <= 0 {
		return errors.New("invalid TimestampMultiplier in FixedHeader")
	}
	return nil
}

func (r *BaseReader) ReadVarHeader(ownSchema *schema.Schema) error {
	if _, err := r.FrameDecoder.Next(); err != nil {
		return err
	}

	hdrBytes := make([]byte, r.FrameDecoder.RemainingSize())
	n, err := r.FrameDecoder.Read(hdrBytes)
	if err != nil {
		return err
	}
	hdrBytes = hdrBytes[:n]

	err = json.Unmarshal(hdrBytes, &r.VarHeader)
	if err != nil {
		return err
	}

	if r.VarHeader.Schema != nil {
		r.Schema = &schema.Schema{}
		err = json.Unmarshal(*r.VarHeader.Schema, r.Schema)
		if err != nil {
			return err
		}
		if _, err := ownSchema.Compatible(r.Schema); err != nil {
			return fmt.Errorf("schema is not compatible with BaseReader: %w", err)
		}
	}

	return nil
}

func (r *BaseReader) NextFrame() error {
	frameFlag, err := r.FrameDecoder.Next()
	_ = frameFlag
	if err != nil {
		return err
	}
	//if frameFlag&pkg.RestartDictionaries != 0 {
	//	r.restartDictionaries()
	//}

	r.FrameRecordCount, err = binary.ReadUvarint(&r.FrameDecoder)
	if err != nil {
		return err
	}

	if err := r.ReadBufs.ReadFrom(&r.FrameDecoder); err != nil {
		return err
	}

	return nil
}
