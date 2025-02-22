package metrics

import (
	"errors"
	"io"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pdata/metrics/sortedbyresource"
)

type STEFToSortedTree struct {
}

func NewSTEFToSortedTree() *STEFToSortedTree {
	return &STEFToSortedTree{}
}

func (c *STEFToSortedTree) FromTef(reader *oteltef.MetricsReader) (*sortedbyresource.SortedTree, error) {
	sm := sortedbyresource.NewSortedByResource()

	i := 0
	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		resource := sm.ByResource(record.Resource())
		scope := resource.ByScope(record.Scope())
		metric := scope.ByMetric(record.Metric())
		timedValues := metric.ByAttrs(record.Attributes())
		point := oteltef.NewPoint()
		point.CopyFrom(record.Point())
		*timedValues = append(*timedValues, point)
		i++
	}

	return sm, nil
}
