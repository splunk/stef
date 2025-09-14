package metrics

import (
	"errors"
	"io"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/metrics/sortedbyresource"
	"github.com/splunk/stef/go/pkg"
)

// StefToOtlpSorted reads and converts STEF records to OTLP metrics, sorted and grouped
// by Resource/Scope/Metric hierarchy.
type StefToOtlpSorted struct {
}

var _ StefToOtlp = (*StefToOtlpSorted)(nil)

func (c *StefToOtlpSorted) Convert(reader *oteltef.MetricsReader, untilEOF bool) (pmetric.Metrics, error) {
	sm := sortedbyresource.NewSortedByResource()
	metrics := pmetric.NewMetrics()

	err := reader.Read(pkg.ReadOptions{})
	if err != nil {
		return metrics, err
	}

	for {
		record := &reader.Record

		resource := sm.ByResource(record.Resource())
		scope := resource.ByScope(record.Scope())
		metric := scope.ByMetric(record.Metric())
		timedValues := metric.ByAttrs(record.Attributes())
		point := oteltef.NewPoint()
		point.CopyFrom(record.Point())
		*timedValues = append(*timedValues, point)

		if untilEOF {
			err = reader.Read(pkg.ReadOptions{})
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return metrics, err
			}
		} else {
			// Read more records. This will not block on I/O.
			err = reader.Read(pkg.ReadOptions{TillEndOfFrame: true})
			if err != nil {
				if errors.Is(err, pkg.ErrEndOfFrame) {
					break
				}
				return metrics, err
			}
		}
	}

	return sm.ToOtlp()
}
