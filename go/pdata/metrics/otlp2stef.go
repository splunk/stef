package metrics

import (
	"log"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/internal/otlptools"
	"github.com/splunk/stef/go/pdata/metrics/internal"
)

type OtlpToSTEFUnsorted struct {
	internal.BaseOTLPToSTEF
}

func convertTemporality(temporality pmetric.AggregationTemporality) internal.MetricFlags {
	switch temporality {
	case pmetric.AggregationTemporalityCumulative:
		return internal.MetricTemporalityCumulative
	case pmetric.AggregationTemporalityDelta:
		return internal.MetricTemporalityDelta
	case pmetric.AggregationTemporalityUnspecified:
		return internal.MetricTemporalityUnspecified
	default:
		panic("unhandled default case")
	}
}

func metricType(typ pmetric.MetricType) oteltef.MetricType {
	switch typ {
	case pmetric.MetricTypeGauge:
		return oteltef.MetricTypeGauge
	case pmetric.MetricTypeSum:
		return oteltef.MetricTypeSum
	case pmetric.MetricTypeHistogram:
		return oteltef.MetricTypeHistogram
	default:
		log.Fatalf("Unsupported value type: %v", typ)
	}
	return 0
}

func metric2metric(
	src pmetric.Metric, // histogramBounds []float64,
	dst *oteltef.Metric,
	otlp2stef *otlptools.Otlp2Stef,
) {
	otlp2stef.MapUnsorted(src.Metadata(), dst.Metadata())
	dst.SetName(src.Name())
	dst.SetDescription(src.Description())
	dst.SetUnit(src.Unit())
	dst.SetType(metricType(src.Type()))
}

func (d *OtlpToSTEFUnsorted) WriteMetrics(src pmetric.Metrics, writer *oteltef.MetricsWriter) error {
	otlp2stef := &otlptools.Otlp2Stef{}
	for i := 0; i < src.ResourceMetrics().Len(); i++ {
		rmm := src.ResourceMetrics().At(i)
		otlp2stef.ResourceUnsorted(writer.Record.Resource(), rmm.Resource(), rmm.SchemaUrl())

		for j := 0; j < rmm.ScopeMetrics().Len(); j++ {
			smm := rmm.ScopeMetrics().At(j)
			otlp2stef.ScopeUnsorted(writer.Record.Scope(), smm.Scope(), smm.SchemaUrl())
			for k := 0; k < smm.Metrics().Len(); k++ {
				m := smm.Metrics().At(k)
				metric2metric(m, writer.Record.Metric(), otlp2stef)
				var err error
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					err = d.writeNumeric(writer, m.Gauge().DataPoints())
				case pmetric.MetricTypeSum:
					writer.Record.Metric().SetAggregationTemporality(uint64(convertTemporality(m.Sum().AggregationTemporality())))
					writer.Record.Metric().SetMonotonic(m.Sum().IsMonotonic())
					err = d.writeNumeric(writer, m.Sum().DataPoints())
				case pmetric.MetricTypeHistogram:
					writer.Record.Metric().SetAggregationTemporality(uint64(convertTemporality(m.Histogram().AggregationTemporality())))
					err = d.writeHistogram(writer, m.Histogram().DataPoints())
				default:
					panic("unhandled default case")
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *OtlpToSTEFUnsorted) writeNumeric(writer *oteltef.MetricsWriter, src pmetric.NumberDataPointSlice) error {
	for i := 0; i < src.Len(); i++ {
		srcPoint := src.At(i)
		dstPoint := writer.Record.Point()

		d.ConvertNumDatapoint(dstPoint, srcPoint)
		d.Otlp2tef.MapUnsorted(srcPoint.Attributes(), writer.Record.Attributes())
		d.ConvertExemplars(dstPoint.Exemplars(), srcPoint.Exemplars())

		err := writer.Write()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *OtlpToSTEFUnsorted) writeHistogram(writer *oteltef.MetricsWriter, src pmetric.HistogramDataPointSlice) error {
	for i := 0; i < src.Len(); i++ {
		src := src.At(i)
		dst := writer.Record.Point()

		err := d.ConvertHistogram(dst, src)
		if err != nil {
			return err
		}

		d.Otlp2tef.MapUnsorted(src.Attributes(), writer.Record.Attributes())
		writer.Record.Metric().HistogramBounds().CopyFromSlice(src.ExplicitBounds().AsRaw())
		d.ConvertExemplars(dst.Exemplars(), src.Exemplars())

		err = writer.Write()
		if err != nil {
			return err
		}
	}
	return nil
}
