package metrics

import (
	"log"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/stef-otel/oteltef"
	"github.com/splunk/stef/stef-pdata/internal/otlptools"
	"github.com/splunk/stef/stef-pdata/metrics/internal"
)

type OtlpToTEFUnsorted struct {
	internal.BaseOTLPToTEF
}

func convertTemporality(temporality pmetric.AggregationTemporality) oteltef.MetricFlags {
	switch temporality {
	case pmetric.AggregationTemporalityCumulative:
		return oteltef.MetricTemporalityCumulative
	case pmetric.AggregationTemporalityDelta:
		return oteltef.MetricTemporalityDelta
	case pmetric.AggregationTemporalityUnspecified:
		return oteltef.MetricTemporalityUnspecified
	default:
		panic("unhandled default case")
	}
}

func metricType(typ pmetric.MetricType) oteltef.MetricType {
	switch typ {
	case pmetric.MetricTypeGauge:
		return oteltef.Gauge
	case pmetric.MetricTypeSum:
		return oteltef.Sum
	case pmetric.MetricTypeHistogram:
		return oteltef.Histogram
	default:
		log.Fatalf("Unsupported value type: %v", typ)
	}
	return 0
}

func metric2metric(
	src pmetric.Metric, // histogramBounds []float64,
	dst *oteltef.Metric,
	otlp2tef *otlptools.Otlp2Tef,
) {
	otlp2tef.MapUnsorted(src.Metadata(), dst.Metadata())
	dst.SetName(src.Name())
	dst.SetDescription(src.Description())
	dst.SetUnit(src.Unit())
	dst.SetType(uint64(metricType(src.Type())))
}

func (d *OtlpToTEFUnsorted) WriteMetrics(src pmetric.Metrics, writer *oteltef.MetricsWriter) error {
	otlp2tef := &otlptools.Otlp2Tef{}
	for i := 0; i < src.ResourceMetrics().Len(); i++ {
		rmm := src.ResourceMetrics().At(i)
		otlp2tef.ResourceUnsorted(writer.Record.Resource(), rmm.Resource(), rmm.SchemaUrl())

		for j := 0; j < rmm.ScopeMetrics().Len(); j++ {
			smm := rmm.ScopeMetrics().At(j)
			otlp2tef.ScopeUnsorted(writer.Record.Scope(), smm.Scope(), smm.SchemaUrl())
			for k := 0; k < smm.Metrics().Len(); k++ {
				m := smm.Metrics().At(k)
				metric2metric(m, writer.Record.Metric(), otlp2tef)
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

func (d *OtlpToTEFUnsorted) writeNumeric(writer *oteltef.MetricsWriter, src pmetric.NumberDataPointSlice) error {
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

func (d *OtlpToTEFUnsorted) writeHistogram(writer *oteltef.MetricsWriter, src pmetric.HistogramDataPointSlice) error {
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
