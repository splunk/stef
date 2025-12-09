package stef

import (
	"bytes"
	"io"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/otelstef"
	otlpconvert "github.com/splunk/stef/go/pdata/metrics"

	"github.com/splunk/stef/benchmarks/encodings"
)

// STEFUEncoding is unsorted STEF format.
type STEFUEncoding struct {
	Opts pkg.WriterOptions
}

func (d *STEFUEncoding) LongName() string {
	return "STEF Unsorted"
}

func (d *STEFUEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return data, nil
}

func (d *STEFUEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	metrics := data.(pmetric.Metrics)

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewMetricsWriter(outputBuf, d.Opts)
	if err != nil {
		return nil, err
	}

	converter := otlpconvert.OtlpToStefUnsorted{}
	err = converter.Convert(metrics, writer)
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
	r, err := otelstef.NewMetricsReader(buf)
	if err != nil {
		return nil, err
	}

	for {
		err := r.Read(pkg.ReadOptions{})
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (*STEFUEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	buf := bytes.NewBuffer(data)
	reader, err := otelstef.NewMetricsReader(buf)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	converter := otlpconvert.StefToOtlpUnsorted{}
	metrics, err := converter.Convert(reader, true)
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

	writer, err := otelstef.NewMetricsWriter(
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
	writer    *otelstef.MetricsWriter
}

func (s *stefuMultipart) AppendPart(part pmetric.Metrics) error {
	converter := otlpconvert.OtlpToStefUnsorted{}
	err := converter.Convert(part, s.writer)
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
