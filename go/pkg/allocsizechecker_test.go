package pkg

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddAllocSizeSaturatesOnOverflow(t *testing.T) {
	checker := &AllocSizeChecker{allocatedSize: math.MaxUint - 5}

	checker.AddAllocSize(10)

	require.Equal(t, uint(math.MaxUint), checker.allocatedSize)
	require.True(t, checker.IsOverLimit())
}

func TestPrepAllocSizeNHandlesMulOverflow(t *testing.T) {
	checker := &AllocSizeChecker{}

	err := checker.PrepAllocSizeN(math.MaxUint, 2)

	require.Equal(t, uint(math.MaxUint), checker.allocatedSize)
	require.ErrorIs(t, err, ErrRecordAllocLimitExceeded)
}

func TestPrepAllocSizeNHandlesAddOverflow(t *testing.T) {
	checker := &AllocSizeChecker{allocatedSize: math.MaxUint - 1}

	err := checker.PrepAllocSizeN(2, 1)

	require.Equal(t, uint(math.MaxUint), checker.allocatedSize)
	require.ErrorIs(t, err, ErrRecordAllocLimitExceeded)
}

func TestPrepAllocSizeNHappyPath(t *testing.T) {
	checker := &AllocSizeChecker{}

	err := checker.PrepAllocSizeN(1024, 10)

	require.NoError(t, err)
	require.Equal(t, uint(10240), checker.allocatedSize)
	require.False(t, checker.IsOverLimit())
}

func TestPrepAllocSizeHappyPath(t *testing.T) {
	checker := &AllocSizeChecker{}

	err := checker.PrepAllocSize(2048)

	require.NoError(t, err)
	require.Equal(t, uint(2048), checker.allocatedSize)
	require.False(t, checker.IsOverLimit())
}

func TestPrepAllocSizeOverLimit(t *testing.T) {
	checker := &AllocSizeChecker{}

	err := checker.PrepAllocSize(RecordAllocLimit + 1)

	require.ErrorIs(t, err, ErrRecordAllocLimitExceeded)
	require.Equal(t, uint(RecordAllocLimit+1), checker.allocatedSize)
	require.True(t, checker.IsOverLimit())
}

func TestPrepAllocSizeOverflowSaturates(t *testing.T) {
	checker := &AllocSizeChecker{allocatedSize: math.MaxUint - 1}

	err := checker.PrepAllocSize(10)

	require.Equal(t, uint(math.MaxUint), checker.allocatedSize)
	require.ErrorIs(t, err, ErrRecordAllocLimitExceeded)
}

func BenchmarkAddAllocSize(b *testing.B) {
	cases := []struct {
		name    string
		initial uint
		inc     uint
	}{
		{name: "small", inc: 64},
		{name: "near-limit", initial: RecordAllocLimit - 8, inc: 4},
		{name: "overflow", initial: math.MaxUint - 1, inc: 5},
	}
	for _, tt := range cases {
		b.Run(
			tt.name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					checker := AllocSizeChecker{allocatedSize: tt.initial}
					checker.AddAllocSize(tt.inc)
				}
			},
		)
	}
}

func BenchmarkPrepAllocSize(b *testing.B) {
	cases := []struct {
		name      string
		initial   uint
		size      uint
		expectErr bool
	}{
		{name: "below-limit", size: 1},
		{name: "at-limit", initial: RecordAllocLimit - 1, size: 1},
		{name: "over-limit", size: RecordAllocLimit + 1, expectErr: true},
	}
	for _, tt := range cases {
		b.Run(
			tt.name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					checker := AllocSizeChecker{allocatedSize: tt.initial}
					if err := checker.PrepAllocSize(tt.size); (err != nil) != tt.expectErr {
						b.Fatalf("unexpected error state: %v", err)
					}
				}
			},
		)
	}
}

func BenchmarkPrepAllocSizeN(b *testing.B) {
	cases := []struct {
		name      string
		initial   uint
		size      uint
		count     uint
		expectErr bool
	}{
		{name: "small", size: 4, count: 1},
		{name: "large-count", size: 1024, count: 1024},
		{name: "mul-overflow", size: math.MaxUint, count: 2, expectErr: true},
		{name: "add-overflow", initial: math.MaxUint - 1, size: 2, count: 1, expectErr: true},
	}
	for _, tt := range cases {
		b.Run(
			tt.name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					checker := AllocSizeChecker{allocatedSize: tt.initial}
					if err := checker.PrepAllocSizeN(tt.size, tt.count); (err != nil) != tt.expectErr {
						b.Fatalf("unexpected error state: %v", err)
					}
				}
			},
		)
	}
}
