package internal

// MetricFlags is a bitmask
type MetricFlags uint

const (
	MetricMonotonic              MetricFlags = 0b001
	MetricTemporalityMask        MetricFlags = 0b110
	MetricTemporalityUnspecified MetricFlags = 0b000
	MetricTemporalityDelta       MetricFlags = 0b010
	MetricTemporalityCumulative  MetricFlags = 0b100
)
