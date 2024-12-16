package stef

import (
	"bytes"
	"io"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/tigrannajaryan/stef/stef-go/pkg"

	"github.com/tigrannajaryan/stef/stef-otel/oteltef"
	otlpconvert "github.com/tigrannajaryan/stef/stef-pdata/metrics"

	"github.com/tigrannajaryan/stef/benchmarks/encodings"
)

// STEFUEncoding is unsorted TEF format.
type STEFUEncoding struct {
	Opts pkg.WriterOptions
}

func (d *STEFUEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return data, nil
}

func (d *STEFUEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	metrics := data.(pmetric.Metrics)

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := oteltef.NewMetricsWriter(outputBuf, d.Opts)
	if err != nil {
		return nil, err
	}

	converter := otlpconvert.OtlpToTEFUnsorted{}
	err = converter.WriteMetrics(metrics, writer)
	if err != nil {
		return nil, err
	}

	err = writer.Flush()
	if err != nil {
		return nil, err
	}

	return outputBuf.Bytes(), nil
}

func (d *STEFUEncoding) Decode(b []byte) (any, error) {
	buf := bytes.NewBuffer(b)
	r, err := oteltef.NewMetricsReader(buf)
	if err != nil {
		return nil, err
	}

	for {
		readRecord, err := r.Read()
		if err == io.EOF {
			break
		}
		if readRecord == nil {
			panic("nil record")
		}
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (*STEFUEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	buf := bytes.NewBuffer(data)
	reader, err := oteltef.NewMetricsReader(buf)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	converter := otlpconvert.TEFToOTLPUnsorted{}
	metrics, err := converter.Convert(reader)
	if err != nil {
		return pmetric.NewMetrics(), err
	}
	return metrics, nil
}

func (e *STEFUEncoding) Name() string {
	str := "STEFU"
	if e.Opts.Compression != pkg.CompressionNone {
		str += "Z"
	}
	return str
}

func (e *STEFUEncoding) StartMultipart(compression string) (encodings.MetricMultipartStream, error) {
	outputBuf := &pkg.MemChunkWriter{}

	opts := pkg.WriterOptions{}
	if compression == "zstd" {
		opts.Compression = pkg.CompressionZstd
	}

	writer, err := oteltef.NewMetricsWriter(
		outputBuf, opts,
	)
	if err != nil {
		return nil, err
	}
	return &stefuMultipart{
		outputBuf: outputBuf,
		writer:    writer,
	}, nil
}

type stefuMultipart struct {
	outputBuf *pkg.MemChunkWriter
	writer    *oteltef.MetricsWriter
}

func (s *stefuMultipart) AppendPart(part pmetric.Metrics) error {
	converter := otlpconvert.OtlpToTEFUnsorted{}
	err := converter.WriteMetrics(part, s.writer)
	if err != nil {
		return err
	}

	return s.writer.Flush()
}

func (s *stefuMultipart) FinishStream() ([]byte, error) {
	err := s.writer.Flush()
	if err != nil {
		return nil, err
	}
	return s.outputBuf.Bytes(), nil
}
