package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/go/otel/otelstef"
)

func TestConvertNumDatapoint_EmptyValueType(t *testing.T) {
	c := &BaseOtlpToStef{}
	dst := otelstef.NewPoint()
	src := pmetric.NewNumberDataPoint()
	src.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	// Don't set any value — ValueType will be NumberDataPointValueTypeEmpty

	err := c.ConvertNumDatapoint(dst, src)
	require.NoError(t, err)
	assert.Equal(t, otelstef.PointValueTypeNone, dst.Value().Type())
}

func TestConvertNumDatapoint_IntValue(t *testing.T) {
	c := &BaseOtlpToStef{}
	dst := otelstef.NewPoint()
	src := pmetric.NewNumberDataPoint()
	src.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	src.SetIntValue(42)

	err := c.ConvertNumDatapoint(dst, src)
	require.NoError(t, err)
	assert.Equal(t, int64(42), dst.Value().Int64())
}

func TestConvertNumDatapoint_DoubleValue(t *testing.T) {
	c := &BaseOtlpToStef{}
	dst := otelstef.NewPoint()
	src := pmetric.NewNumberDataPoint()
	src.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	src.SetDoubleValue(3.14)

	err := c.ConvertNumDatapoint(dst, src)
	require.NoError(t, err)
	assert.InDelta(t, 3.14, dst.Value().Float64(), 0.001)
}

func TestConvertNumDatapoint_NoRecordedValue(t *testing.T) {
	c := &BaseOtlpToStef{}
	dst := otelstef.NewPoint()
	src := pmetric.NewNumberDataPoint()
	src.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	src.SetFlags(pmetric.DefaultDataPointFlags.WithNoRecordedValue(true))

	err := c.ConvertNumDatapoint(dst, src)
	require.NoError(t, err)
	assert.Equal(t, otelstef.PointValueTypeNone, dst.Value().Type())
}

func TestAggregationTemporalityToStef_AllValues(t *testing.T) {
	tests := []struct {
		name     string
		input    pmetric.AggregationTemporality
		expected otelstef.AggregationTemporality
		wantErr  bool
	}{
		{name: "delta", input: pmetric.AggregationTemporalityDelta, expected: otelstef.AggregationTemporalityDelta},
		{name: "cumulative", input: pmetric.AggregationTemporalityCumulative, expected: otelstef.AggregationTemporalityCumulative},
		{name: "unspecified", input: pmetric.AggregationTemporalityUnspecified, expected: otelstef.AggregationTemporalityUnspecified},
		{name: "unknown", input: pmetric.AggregationTemporality(99), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AggregationTemporalityToStef(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
