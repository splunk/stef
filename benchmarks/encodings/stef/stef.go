package stef

import (
	"bytes"
	"io"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/oteltef"
	otlpconvert "github.com/splunk/stef/go/pdata/metrics"
	"github.com/splunk/stef/go/pdata/metrics/sortedbymetric"

	"github.com/splunk/stef/benchmarks/encodings"
)

type STEFEncoding struct {
	Opts pkg.WriterOptions
}

func (d *STEFEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	return sortedbymetric.OtlpToSortedTree(data)
}

func (d *STEFEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	sorted := data.(*sortedbymetric.SortedTree)

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := oteltef.NewMetricsWriter(outputBuf, d.Opts)
	if err != nil {
		return nil, err
	}

	err = sorted.ToStef(writer)
	if err != nil {
		return nil, err
	}

	err = writer.Flush()
	if err != nil {
		return nil, err
	}

	return outputBuf.Bytes(), nil
}

func (d *STEFEncoding) Decode(b []byte) (any, error) {
	buf := bytes.NewBuffer(b)
	r, err := oteltef.NewMetricsReader(buf)
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

func (*STEFEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
	buf := bytes.NewBuffer(data)
	reader, err := oteltef.NewMetricsReader(buf)
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

func (e *STEFEncoding) Name() string {
	str := "STEF"
	if e.Opts.Compression != pkg.CompressionNone {
		str += "Z"
	}
	return str
}

func (e *STEFEncoding) LongName() string {
	return "STEF Sorted"
}

func (e *STEFEncoding) StartMultipart(compression string) (encodings.MetricMultipartStream, error) {
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
	return &stefMultipart{
		outputBuf: outputBuf,
		writer:    writer,
	}, nil
}

type stefMultipart struct {
	outputBuf *pkg.MemChunkWriter
	writer    *oteltef.MetricsWriter
}

func (s *stefMultipart) AppendPart(part pmetric.Metrics) error {
	tree, err := sortedbymetric.OtlpToSortedTree(part)
	if err != nil {
		return err
	}

	if err := tree.ToStef(s.writer); err != nil {
		return err
	}
	return s.writer.Flush()
}

func (s *stefMultipart) FinishStream() ([]byte, error) {
	err := s.writer.Flush()
	if err != nil {
		return nil, err
	}
	return s.outputBuf.Bytes(), nil
}
