package generators

import (
	"strconv"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Random struct {
	Name                    string
	TimeseriesCount         int
	DatapointsPerTimeseries int
	IncludeInt64            bool
	IncludeHistogram        bool
	IncludeSummary          bool
}

func (r *Random) GetName() string {
	return r.Name
}

func (r *Random) Generate() pmetric.Metrics {
	ret := pmetric.NewMetrics()
	rms := ret.ResourceMetrics()
	rm := rms.AppendEmpty()

	genResource(rm.Resource())
	ils := rm.ScopeMetrics()
	il := ils.AppendEmpty()

	for tsIndex := 0; tsIndex < r.TimeseriesCount; tsIndex++ {
		startTime := time.Date(2019, 10, 31, 10, 11, 12, 13, time.UTC)

		if r.IncludeInt64 {
			metric := il.Metrics().AppendEmpty()
			genInt64Gauge(startTime, tsIndex, r.DatapointsPerTimeseries, metric)
		}
		if r.IncludeHistogram {
			metric := il.Metrics().AppendEmpty()
			genHistogram(startTime, tsIndex, r.DatapointsPerTimeseries, metric)
		}
		if r.IncludeSummary {
			metric := il.Metrics().AppendEmpty()
			genSummary(startTime, tsIndex, r.DatapointsPerTimeseries, metric)
		}
	}
	return ret
}

func genResource(resource pcommon.Resource) {
	resource.Attributes().FromRaw(
		map[string]any{
			"StartTimeUnixnano": 12345678,
			"Pid":               1234,
			"HostName":          "fakehost",
			"ServiceName":       "generator",
		},
	)
}

func timeToTimestamp(t time.Time) uint64 {
	return uint64(t.UnixNano())
}

func genInt64Timeseries(
	startTime time.Time, tsIndex int, valuesPerTimeseries int, into pmetric.NumberDataPointSlice,
) {
	for j := 0; j < 5; j++ {
		for k := 0; k < valuesPerTimeseries; k++ {
			pointTs := timeToTimestamp(startTime.Add(time.Duration(j*k) * time.Millisecond))
			point := into.AppendEmpty()

			point.SetTimestamp(pcommon.Timestamp(pointTs))
			point.SetIntValue(int64(tsIndex * j * k))

			point.Attributes().FromRaw(
				map[string]any{
					"label1": "val1",
					"label2": strconv.Itoa(tsIndex),
				},
			)

			if k == 0 {
				point.SetStartTimestamp(pcommon.Timestamp(pointTs))
			}
		}
	}
}

func genInt64Gauge(startTime time.Time, tsIndex int, valuesPerTimeseries int, metric1 pmetric.Metric) {
	genMetricDescriptor(0, metric1)
	genInt64Timeseries(startTime, tsIndex, valuesPerTimeseries, metric1.SetEmptyGauge().DataPoints())
}

func genMetricDescriptor(metricIndex int, into pmetric.Metric) {
	into.SetName("metric" + strconv.Itoa(metricIndex))
	into.SetDescription("some description: " + strconv.Itoa(metricIndex))
}

func genHistogram(startTime time.Time, tsIndex int, valuesPerTimeseries int, into pmetric.Metric) {
	// Add Histogram
	genMetricDescriptor(0, into)

	histogram := into.SetEmptyHistogram()

	for j := 0; j < 1; j++ {
		for k := 0; k < valuesPerTimeseries; k++ {
			pointTs := timeToTimestamp(startTime.Add(time.Duration(j*k) * time.Millisecond))
			point := histogram.DataPoints().AppendEmpty()

			point.SetTimestamp(pcommon.Timestamp(pointTs))
			val := float64(tsIndex * j * k)
			point.SetSum(val)
			point.SetCount(1)
			point.BucketCounts().FromRaw([]uint64{12, 345})

			point.Attributes().FromRaw(
				map[string]any{
					"label1": "val1",
					"label2": strconv.Itoa(tsIndex),
				},
			)

			if k == 0 {
				point.SetStartTimestamp(pcommon.Timestamp(pointTs))
			}
			ex := point.Exemplars().AppendEmpty()

			ex.SetDoubleValue(val)
			ex.SetTimestamp(pcommon.Timestamp(pointTs))

			point.ExplicitBounds().FromRaw([]float64{0, 1000000})
		}
	}
}

func genSummary(startTime time.Time, tsIndex int, valuesPerTimeseries int, into pmetric.Metric) {
	genMetricDescriptor(0, into)
	summary := into.SetEmptySummary()

	for j := 0; j < 1; j++ {
		for k := 0; k < valuesPerTimeseries; k++ {
			pointTs := timeToTimestamp(startTime.Add(time.Duration(j*k) * time.Millisecond))
			point := summary.DataPoints().AppendEmpty()

			point.SetTimestamp(pcommon.Timestamp(pointTs))
			val := float64(tsIndex * j * k)
			point.SetSum(val)
			point.SetCount(1)

			point.Attributes().FromRaw(
				map[string]any{
					"label1": "val1",
					"label2": strconv.Itoa(tsIndex),
				},
			)

			if k == 0 {
				point.SetStartTimestamp(pcommon.Timestamp(pointTs))
			}

			q := point.QuantileValues().AppendEmpty()

			q.SetQuantile(99)
			q.SetValue(val / 10)
		}
	}
}
