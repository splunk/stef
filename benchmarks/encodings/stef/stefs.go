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

// STEFSEncoding is a partially (fast) sorted STEF format.
type STEFSEncoding struct {
	Opts   pkg.WriterOptions
	sorter otlpconvert.PDataSorter
}

func (d *STEFSEncoding) LongName() string {
	return "STEFS"
}

func (d *STEFSEncoding) FromOTLP(data pmetric.Metrics) (encodings.InMemoryData, error) {
	d.sorter.SortMetrics(data, false)
	//otlpconvert2.NormalizeMetrics(data)
	return data, nil
}

func (d *STEFSEncoding) Encode(data encodings.InMemoryData) ([]byte, error) {
	metrics := data.(pmetric.Metrics)

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewMetricsWriter(outputBuf, d.Opts)
	if err != nil {
		return nil, err
	}

	converter := otlpconvert.NewOtlpToStefUnsorted(true)
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

func (d *STEFSEncoding) Decode(b []byte) (any, error) {
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

func (*STEFSEncoding) ToOTLP(data []byte) (pmetric.Metrics, error) {
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

func (e *STEFSEncoding) Name() string {
	str := "STEFS"
	if e.Opts.Compression != pkg.CompressionNone {
		str += "Z"
	}
	return str
}

func (e *STEFSEncoding) StartMultipart(compression string) (encodings.MetricMultipartStream, error) {
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
	return &stefsMultipart{
		outputBuf: outputBuf,
		writer:    writer,
	}, nil
}

type stefsMultipart struct {
	outputBuf *pkg.MemChunkWriter
	writer    *otelstef.MetricsWriter
	sorter    otlpconvert.PDataSorter
}

func (m *stefsMultipart) AppendPart(part pmetric.Metrics) error {
	m.sorter.SortMetrics(part, false)
	//otlpconvert2.NormalizeMetrics(part)
	converter := otlpconvert.OtlpToStefUnsorted{}
	err := converter.Convert(part, m.writer)
	if err != nil {
		return err
	}

	return m.writer.Flush()
}

func (s *stefsMultipart) FinishStream() ([]byte, error) {
	err := s.writer.Flush()
	if err != nil {
		return nil, err
	}
	return s.outputBuf.Bytes(), nil
}
