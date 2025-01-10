package testutils

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func openOTLPFile(filePath string) (io.Reader, func(), error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	var reader io.Reader
	reader = bufio.NewReader(f)

	var zcloser func()
	if strings.HasSuffix(filePath, ".zst") {
		zreader, err := zstd.NewReader(reader)
		if err != nil {
			return nil, nil, err
		}
		reader = zreader
		zcloser = func() {
			zreader.Close()
		}
	}

	closer := func() {
		if zcloser != nil {
			zcloser()
		}
		_ = f.Close()
	}

	return reader, closer, nil
}

func ReadMultipartOTLPFile(filePath string) ([]pmetric.Metrics, error) {
	var result []pmetric.Metrics

	reader, closer, err := openOTLPFile(filePath)
	if err != nil {
		return nil, err
	}
	defer closer()

	protoUnmarshaler := pmetric.ProtoUnmarshaler{}

	for {
		var inputBytes []byte
		var sizeBytes [4]byte
		n, err := reader.Read(sizeBytes[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if n != 4 {
			return nil, errors.New("invalid input")
		}

		bytesSize := uint64(binary.BigEndian.Uint32(sizeBytes[:]))

		inputBytes = make([]byte, bytesSize)
		n, err = io.ReadFull(reader, inputBytes)
		if err != nil {
			return nil, err
		}
		if n != int(bytesSize) {
			return nil, errors.New("invalid input")
		}

		metrics, err := protoUnmarshaler.UnmarshalMetrics(inputBytes)
		if err != nil {
			return nil, err
		}

		result = append(result, metrics)
	}

	return result, nil
}

func ReadMultipartOTLPFileGeneric(filePath string, unmarshaler func([]byte) (any, error)) ([]any, error) {
	var result []any

	reader, closer, err := openOTLPFile(filePath)
	if err != nil {
		return nil, err
	}
	defer closer()

	for {
		var inputBytes []byte
		var sizeBytes [4]byte
		n, err := reader.Read(sizeBytes[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if n != 4 {
			return nil, errors.New("invalid input")
		}

		bytesSize := uint64(binary.BigEndian.Uint32(sizeBytes[:]))

		inputBytes = make([]byte, bytesSize)
		n, err = io.ReadFull(reader, inputBytes)
		if err != nil {
			return nil, err
		}
		if n != int(bytesSize) {
			return nil, errors.New("invalid input")
		}

		metrics, err := unmarshaler(inputBytes)
		if err != nil {
			return nil, err
		}

		result = append(result, metrics)
	}

	return result, nil
}

func readOnePartOTLPFile(filePath string) (pmetric.Metrics, error) {
	reader, closer, err := openOTLPFile(filePath)
	if err != nil {
		return pmetric.Metrics{}, err
	}
	defer closer()

	protoUnmarshaler := pmetric.ProtoUnmarshaler{}

	var inputBytes []byte
	inputBytes = make([]byte, 0, 65536)
	temp := make([]byte, 65536)
	for {
		n, err := io.ReadFull(reader, temp)
		inputBytes = append(inputBytes, temp[:n]...)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return pmetric.Metrics{}, err
		}
	}

	return protoUnmarshaler.UnmarshalMetrics(inputBytes)
}

func ReadOTLPFile(filePath string, multipart bool) (pmetric.Metrics, error) {
	if multipart {
		combined := pmetric.NewMetrics()
		parts, err := ReadMultipartOTLPFile(filePath)
		if err != nil {
			return combined, err
		}
		for _, part := range parts {
			part.ResourceMetrics().MoveAndAppendTo(combined.ResourceMetrics())
		}
		return combined, nil
	}

	return readOnePartOTLPFile(filePath)
}
