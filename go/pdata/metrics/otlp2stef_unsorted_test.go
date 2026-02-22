package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pkg"
)

func TestConvertUnsorted_EmptyNumberDataPointValueType(t *testing.T) {
	// Build metrics with a Gauge that has an empty value type data point.
	// This previously caused a panic in ConvertNumDatapoint.
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	m := sm.Metrics().AppendEmpty()
	m.SetName("test_gauge")
	m.SetEmptyGauge()
	dp := m.Gauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	// Don't set any value — ValueType will be NumberDataPointValueTypeEmpty

	buf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewMetricsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)

	converter := OtlpToStefUnsorted{}
	err = converter.Convert(metrics, writer)
	require.NoError(t, err, "Convert should not panic or error on empty value type")

	err = writer.Flush()
	require.NoError(t, err)
}

func TestConvertUnsorted_MixedValueTypes(t *testing.T) {
	// Build metrics with a mix of int, double, and empty value types.
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	m := sm.Metrics().AppendEmpty()
	m.SetName("test_gauge")
	m.SetEmptyGauge()

	// Data point with int value
	dp1 := m.Gauge().DataPoints().AppendEmpty()
	dp1.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp1.SetIntValue(42)

	// Data point with empty value (previously caused panic)
	dp2 := m.Gauge().DataPoints().AppendEmpty()
	dp2.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	// Data point with double value
	dp3 := m.Gauge().DataPoints().AppendEmpty()
	dp3.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp3.SetDoubleValue(3.14)

	buf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewMetricsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)

	converter := OtlpToStefUnsorted{}
	err = converter.Convert(metrics, writer)
	require.NoError(t, err)

	err = writer.Flush()
	require.NoError(t, err)

	// All 3 data points should have been written
	require.EqualValues(t, 3, writer.RecordCount())
}

func TestConvertSorted_EmptyNumberDataPointValueType(t *testing.T) {
	// Same test for the sorted converter path.
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	m := sm.Metrics().AppendEmpty()
	m.SetName("test_gauge")
	m.SetEmptyGauge()
	dp := m.Gauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	// Don't set any value — ValueType will be NumberDataPointValueTypeEmpty

	buf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewMetricsWriter(buf, pkg.WriterOptions{})
	require.NoError(t, err)

	converter := OtlpToStefSorted{}
	err = converter.Convert(metrics, writer)
	require.NoError(t, err, "Convert should not panic or error on empty value type")

	err = writer.Flush()
	require.NoError(t, err)
}
