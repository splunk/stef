package internal

import "github.com/splunk/stef/go/otel/otelstef"

type MetricFlags struct {
	Temporality otelstef.AggregationTemporality
	Monotonic   bool
}
