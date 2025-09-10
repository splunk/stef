package pkg

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkMembufWriteVaruint(b *testing.B) {
	//b.Skip()
	bw := NewBytesWriter(10000000)
	b.ResetTimer()
	const recCount = 1000000
	for i := 0; i < b.N; i++ {
		bw.Reset()
		for j := 0; j < recCount; j++ {
			bw.WriteUvarint(uint64(j))
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkMembufReadVaruintExp(b *testing.B) {
	bw := BytesWriter{}
	const recCount = 63
	val := uint64(1)
	for j := 0; j < recCount; j++ {
		bw.WriteUvarint(val)
		val *= 2
	}
	b.ResetTimer()
	br := BytesReader{buf: bw.buf}
	for i := 0; i < b.N; i++ {
		br.byteIndex = 0
		checkVal := uint64(1)
		for j := 0; j < 63; j++ {
			val, err := br.ReadUvarint()
			if val != checkVal || err != nil {
				panic(nil)
			}
			checkVal *= 2
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
}

func BenchmarkMembufReadVaruintSizes(b *testing.B) {
	for size := 1; size <= 9; size++ {
		val := uint64((1 << (size * 7)) - 1)
		b.Run(
			strconv.Itoa(size), func(b *testing.B) {
				const recCount = 1000
				bw := BytesWriter{buf: make([]byte, 0, 1000)}
				bw.Reset()
				for j := 0; j < recCount; j++ {
					bw.WriteUvarint(val)
				}
				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.ReadUvarint()
						if rVal != val || err != nil {
							panic(nil)
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
			},
		)
	}
}

func BenchmarkMembufWriteVaruintSizes(b *testing.B) {
	for size := 1; size <= 9; size++ {
		val := uint64((1 << (size * 7)) - 1)
		b.Run(
			strconv.Itoa(size), func(b *testing.B) {
				const recCount = 1000
				bw := BytesWriter{buf: make([]byte, 0, 1000)}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bw.Reset()
					for j := 0; j < recCount; j++ {
						bw.WriteUvarint(val)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount), "ns/op")
			},
		)
	}
}

// Define test patterns for different byte length combinations
var uvar64x4TestPatterns = []struct {
	name   string
	values [4]uint64
}{
	{
		name:   "all_1byte",
		values: [4]uint64{255, 100, 50, 200},
	},
	{
		name:   "all_8byte",
		values: [4]uint64{^uint64(0), ^uint64(0) - 1000, ^uint64(0) - 2000, ^uint64(0) - 3000},
	},
}

func BenchmarkMembufWriteUvar64x4(b *testing.B) {
	for _, pattern := range uvar64x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bw.Reset()
					for j := 0; j < recCount; j++ {
						bw.WriteUvar64x4(pattern.values)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func BenchmarkMembufReadUvar64x4(b *testing.B) {
	for _, pattern := range uvar64x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				// Pre-encode the data for reading
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				for j := 0; j < recCount; j++ {
					bw.WriteUvar64x4(pattern.values)
				}

				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.ReadUvar64x4()
						if err != nil || rVal != pattern.values {
							panic("read error or value mismatch")
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func BenchmarkMembufReadUvar64x4Scalar(b *testing.B) {
	for _, pattern := range uvar64x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				// Pre-encode the data for reading
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				for j := 0; j < recCount; j++ {
					bw.WriteUvar64x4(pattern.values)
				}

				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.readUvar64x4Scalar()
						if err != nil || rVal != pattern.values {
							panic("read error or value mismatch")
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func TestUvar64x4RoundTrip(t *testing.T) {
	// Seed random number generator for reproducible tests
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	// Test multiple iterations with different random values
	for iteration := 0; iteration < 100; iteration++ {
		// Generate 8 random uint64 numbers that span the 1-8 byte range
		// We'll create 2 sets of 4 values each to test two WriteUvar64x4 calls
		testValues := generateSpanningUint64Values(rng)

		// Split into two groups of 4
		values1 := [4]uint64{testValues[0], testValues[1], testValues[2], testValues[3]}
		values2 := [4]uint64{testValues[4], testValues[5], testValues[6], testValues[7]}

		// Write both groups to buffer
		writer := NewBytesWriter(100)
		writer.WriteUvar64x4(values1)
		writer.WriteUvar64x4(values2)

		// Read back both groups
		reader := BytesReader{buf: writer.Bytes()}

		readValues1, err := reader.ReadUvar64x4()
		require.NoError(t, err, "Failed to read first group (seed: %d)", seed)

		readValues2, err := reader.readUvar64x4Scalar()
		require.NoError(t, err, "Failed to read second group (seed: %d)", seed)

		// Verify all values match using testify assertions
		assert.Equal(t, values1, readValues1, "First group values should match (seed: %d)", seed)
		assert.Equal(t, values2, readValues2, "Second group values should match (seed: %d)", seed)

		// Verify we've read all data
		assert.Equal(
			t, len(reader.buf), reader.byteIndex, "Should have read all bytes from buffer (seed: %d)", seed,
		)
	}
}

// generateSpanningUint64Values generates 8 uint64 values that span all byte lengths (1-8 bytes)
// to ensure we test all encoding paths
func generateSpanningUint64Values(rng *rand.Rand) [8]uint64 {
	var values [8]uint64

	// Define ranges for each byte length:
	// 1 byte: 0 to 255 (2^8 - 1)
	// 2 bytes: 256 to 65535 (2^16 - 1)
	// 4 bytes: 65536 to 4294967295 (2^32 - 1)
	// 8 bytes: 4294967296 to 2^64 - 1

	ranges := []struct {
		min, max uint64
		name     string
	}{
		{0, 255, "1-byte"},
		{256, 65535, "2-byte"},
		{65536, 4294967295, "4-byte"},
		{4294967296, ^uint64(0), "8-byte"},
	}

	// Generate 2 values from each range to ensure we test all paths
	for i := 0; i < 8; i++ {
		rangeIdx := i % 4
		r := ranges[rangeIdx]

		if r.max == ^uint64(0) {
			// For 8-byte range, use a different strategy to avoid overflow
			// Generate a random value in the upper range
			values[i] = rng.Uint64()
			if values[i] < r.min {
				values[i] = r.min + rng.Uint64()%(1<<32) // Ensure it's in 8-byte range
			}
		} else {
			// For other ranges, generate within the specified bounds
			values[i] = r.min + rng.Uint64()%(r.max-r.min+1)
		}
	}

	// Shuffle the array to randomize the order and test different combinations
	for i := len(values) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		values[i], values[j] = values[j], values[i]
	}

	return values
}

func TestUvar64x4EdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		values [4]uint64
	}{
		{
			name:   "all_zeros",
			values: [4]uint64{0, 0, 0, 0},
		},
		{
			name:   "all_max_1_byte",
			values: [4]uint64{255, 255, 255, 255},
		},
		{
			name:   "all_max_2_byte",
			values: [4]uint64{65535, 65535, 65535, 65535},
		},
		{
			name:   "all_max_4_byte",
			values: [4]uint64{4294967295, 4294967295, 4294967295, 4294967295},
		},
		{
			name:   "all_max_8_byte",
			values: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)},
		},
		{
			name:   "mixed_boundaries",
			values: [4]uint64{255, 256, 65535, 65536},
		},
		{
			name:   "powers_of_two",
			values: [4]uint64{1, 256, 65536, 4294967296},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				// Write values
				writer := NewBytesWriter(50)
				writer.WriteUvar64x4(tc.values)

				// Read values back
				reader := BytesReader{buf: writer.Bytes()}
				readValues, err := reader.ReadUvar64x4()
				require.NoError(t, err, "Should successfully read values")

				// Verify all values match using testify assertion
				assert.Equal(t, tc.values, readValues, "All values should match exactly")
			},
		)
	}
}

// Define test patterns for different byte length combinations for 64-bit 2-value encoding
var uvar64x2TestPatterns = []struct {
	name   string
	values [2]uint64
}{
	{
		name:   "all_1byte",
		values: [2]uint64{255, 100},
	},
	{
		name:   "all_8byte",
		values: [2]uint64{^uint64(0), ^uint64(0) - 1000},
	},
}

func BenchmarkMembufWriteUvar64x2(b *testing.B) {
	for _, pattern := range uvar64x2TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bw.Reset()
					for j := 0; j < recCount; j++ {
						bw.WriteUvar64x2(pattern.values)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*2), "ns/op")
			},
		)
	}
}

func BenchmarkMembufReadUvar64x2(b *testing.B) {
	for _, pattern := range uvar64x2TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				// Pre-encode the data for reading
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				for j := 0; j < recCount; j++ {
					bw.WriteUvar64x2(pattern.values)
				}

				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.ReadUvar64x2()
						if err != nil || rVal != pattern.values {
							panic("read error or value mismatch")
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*2), "ns/op")
			},
		)
	}
}

func BenchmarkMembufReadUvar64x2Scalar(b *testing.B) {
	for _, pattern := range uvar64x2TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				// Pre-encode the data for reading
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				for j := 0; j < recCount; j++ {
					bw.WriteUvar64x2(pattern.values)
				}

				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.readUvar64x2Scalar()
						if err != nil || rVal != pattern.values {
							panic("read error or value mismatch")
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*2), "ns/op")
			},
		)
	}
}

func TestUvar64x2RoundTrip(t *testing.T) {
	// Seed random number generator for reproducible tests
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	// Test multiple iterations with different random values
	for iteration := 0; iteration < 100; iteration++ {
		// Generate 4 random uint64 numbers that span the 0-8 byte range
		// We'll create 2 sets of 2 values each to test two WriteUvar64x2 calls
		testValues := generateSpanningUint64ValuesFor2x2(rng)

		// Split into two groups of 2
		values1 := [2]uint64{testValues[0], testValues[1]}
		values2 := [2]uint64{testValues[2], testValues[3]}

		// Write both groups to buffer
		writer := NewBytesWriter(100)
		writer.WriteUvar64x2(values1)
		writer.WriteUvar64x2(values2)

		// Read back both groups
		reader := BytesReader{buf: writer.Bytes()}

		readValues1, err := reader.ReadUvar64x2()
		require.NoError(t, err, "Failed to read first group (seed: %d)", seed)

		readValues2, err := reader.readUvar64x2Scalar()
		require.NoError(t, err, "Failed to read second group (seed: %d)", seed)

		// Verify all values match using testify assertions
		assert.Equal(t, values1, readValues1, "First group values should match (seed: %d)", seed)
		assert.Equal(t, values2, readValues2, "Second group values should match (seed: %d)", seed)

		// Verify we've read all data
		assert.Equal(
			t, len(reader.buf), reader.byteIndex, "Should have read all bytes from buffer (seed: %d)", seed,
		)
	}
}

// generateSpanningUint64ValuesFor2x2 generates 4 uint64 values that span all byte lengths (0-8 bytes)
// to ensure we test all encoding paths for 2x2 operations
func generateSpanningUint64ValuesFor2x2(rng *rand.Rand) [4]uint64 {
	var values [4]uint64

	// Define ranges for each byte length:
	// 0 bytes: 0 (special case)
	// 1 byte: 1 to 255
	// 2 bytes: 256 to 65535
	// 3 bytes: 65536 to 16777215
	// 4 bytes: 16777216 to 4294967295
	// 5 bytes: 4294967296 to 1099511627775
	// 6 bytes: 1099511627776 to 281474976710655
	// 7 bytes: 281474976710656 to 72057594037927935
	// 8 bytes: 72057594037927936 to 2^64 - 1

	ranges := []struct {
		min, max uint64
		name     string
	}{
		{0, 0, "0-byte"},
		{1, 255, "1-byte"},
		{256, 65535, "2-byte"},
		{65536, 16777215, "3-byte"},
		{16777216, 4294967295, "4-byte"},
		{4294967296, 1099511627775, "5-byte"},
		{1099511627776, 281474976710655, "6-byte"},
		{281474976710656, 72057594037927935, "7-byte"},
		{72057594037927936, ^uint64(0), "8-byte"},
	}

	// Generate 1 value from different ranges to ensure we test various byte length combinations
	for i := 0; i < 4; i++ {
		rangeIdx := (i * 2) % len(ranges) // Use different ranges for variety
		r := ranges[rangeIdx]

		if r.min == r.max {
			// Special case for 0
			values[i] = 0
		} else if r.max == ^uint64(0) {
			// For 8-byte range, use a different strategy to avoid overflow
			values[i] = rng.Uint64()
			if values[i] < r.min {
				values[i] = r.min + rng.Uint64()%(1<<32) // Ensure it's in 8-byte range
			}
		} else {
			// For other ranges, generate within the specified bounds
			values[i] = r.min + rng.Uint64()%(r.max-r.min+1)
		}
	}

	// Shuffle the array to randomize the order and test different combinations
	for i := len(values) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		values[i], values[j] = values[j], values[i]
	}

	return values
}

func TestUvar64x2EdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		values [2]uint64
	}{
		{
			name:   "both_zeros",
			values: [2]uint64{0, 0},
		},
		{
			name:   "zero_and_max",
			values: [2]uint64{0, ^uint64(0)},
		},
		{
			name:   "max_and_zero",
			values: [2]uint64{^uint64(0), 0},
		},
		{
			name:   "all_max_1_byte",
			values: [2]uint64{255, 255},
		},
		{
			name:   "all_max_2_byte",
			values: [2]uint64{65535, 65535},
		},
		{
			name:   "all_max_3_byte",
			values: [2]uint64{16777215, 16777215},
		},
		{
			name:   "all_max_4_byte",
			values: [2]uint64{4294967295, 4294967295},
		},
		{
			name:   "all_max_5_byte",
			values: [2]uint64{1099511627775, 1099511627775},
		},
		{
			name:   "all_max_6_byte",
			values: [2]uint64{281474976710655, 281474976710655},
		},
		{
			name:   "all_max_7_byte",
			values: [2]uint64{72057594037927935, 72057594037927935},
		},
		{
			name:   "all_max_8_byte",
			values: [2]uint64{^uint64(0), ^uint64(0)},
		},
		{
			name:   "mixed_boundaries",
			values: [2]uint64{255, 256},
		},
		{
			name:   "powers_of_two",
			values: [2]uint64{1, 256},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				// Write values
				writer := NewBytesWriter(50)
				writer.WriteUvar64x2(tc.values)

				// Read values back
				reader := BytesReader{buf: writer.Bytes()}
				readValues, err := reader.ReadUvar64x2()
				require.NoError(t, err, "Should successfully read values")

				// Verify all values match using testify assertion
				assert.Equal(t, tc.values, readValues, "All values should match exactly")
			},
		)
	}
}

// Define test patterns for different byte length combinations for 32-bit values
var uvar32x4TestPatterns = []struct {
	name   string
	values [4]uint32
}{
	{
		name:   "all_1byte",
		values: [4]uint32{255, 100, 50, 200},
	},
	{
		name:   "all_4byte",
		values: [4]uint32{4294967295, 1000000000, 2000000000, 3000000000},
	},
}

func BenchmarkMembufWriteUvar32x4(b *testing.B) {
	for _, pattern := range uvar32x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bw.Reset()
					for j := 0; j < recCount; j++ {
						bw.WriteUvar32x4(pattern.values)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func BenchmarkMembufWriteUvar32x4Scalar(b *testing.B) {
	for _, pattern := range uvar32x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bw.Reset()
					for j := 0; j < recCount; j++ {
						bw.writeUvar32x4Scalar(pattern.values)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func BenchmarkMembufReadUvar32x4(b *testing.B) {
	for _, pattern := range uvar32x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				// Pre-encode the data for reading
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				for j := 0; j < recCount; j++ {
					bw.WriteUvar32x4(pattern.values)
				}

				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.ReadUvar32x4()
						if err != nil || rVal != pattern.values {
							panic("read error or value mismatch")
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func BenchmarkMembufReadUvar32x4Scalar(b *testing.B) {
	for _, pattern := range uvar32x4TestPatterns {
		b.Run(
			pattern.name, func(b *testing.B) {
				const recCount = 1000
				// Pre-encode the data for reading
				bw := BytesWriter{buf: make([]byte, 0, 10000)}
				for j := 0; j < recCount; j++ {
					bw.WriteUvar32x4(pattern.values)
				}

				b.ResetTimer()
				br := BytesReader{buf: bw.buf}
				for i := 0; i < b.N; i++ {
					br.byteIndex = 0
					for j := 0; j < recCount; j++ {
						rVal, err := br.readUvar32x4Scalar()
						if err != nil || rVal != pattern.values {
							panic("read error or value mismatch")
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*recCount*4), "ns/op")
			},
		)
	}
}

func TestUvar32x4RoundTrip(t *testing.T) {
	// Seed random number generator for reproducible tests
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	// Test multiple iterations with different random values
	for iteration := 0; iteration < 100; iteration++ {
		// Generate 8 random uint32 numbers that span the 1-4 byte range
		// We'll create 2 sets of 4 values each to test two WriteUvar32x4 calls
		testValues := generateSpanningUint32Values(rng)

		// Split into two groups of 4
		values1 := [4]uint32{testValues[0], testValues[1], testValues[2], testValues[3]}
		values2 := [4]uint32{testValues[4], testValues[5], testValues[6], testValues[7]}

		// Write both groups to buffer
		writer := NewBytesWriter(100)
		writer.WriteUvar32x4(values1)
		writer.WriteUvar32x4(values2)
		writer.writeUvar32x4Scalar(values1)
		writer.writeUvar32x4Scalar(values2)

		// Read back both groups
		reader := BytesReader{buf: writer.Bytes()}

		readValues1, err := reader.ReadUvar32x4()
		require.NoError(t, err, "Failed to read first group (seed: %d)", seed)

		readValues2, err := reader.readUvar32x4Scalar()
		require.NoError(t, err, "Failed to read second group (seed: %d)", seed)

		// Verify all values match using testify assertions
		assert.Equal(t, values1, readValues1, "First group values should match (seed: %d)", seed)
		assert.Equal(t, values2, readValues2, "Second group values should match (seed: %d)", seed)

		readValues1, err = reader.ReadUvar32x4()
		require.NoError(t, err, "Failed to read first group (seed: %d)", seed)

		readValues2, err = reader.readUvar32x4Scalar()
		require.NoError(t, err, "Failed to read second group (seed: %d)", seed)

		// Verify all values match using testify assertions
		assert.Equal(t, values1, readValues1, "First group values should match (seed: %d)", seed)
		assert.Equal(t, values2, readValues2, "Second group values should match (seed: %d)", seed)

		// Verify we've read all data
		assert.Equal(
			t, len(reader.buf), reader.byteIndex, "Should have read all bytes from buffer (seed: %d)", seed,
		)
	}
}

// generateSpanningUint32Values generates 8 uint32 values that span all byte lengths (1-4 bytes)
// to ensure we test all encoding paths
func generateSpanningUint32Values(rng *rand.Rand) [8]uint32 {
	var values [8]uint32

	// Define ranges for each byte length:
	// 1 byte: 0 to 255 (2^8 - 1)
	// 2 bytes: 256 to 65535 (2^16 - 1)
	// 3 bytes: 65536 to 16777215 (2^24 - 1)
	// 4 bytes: 16777216 to 2^32 - 1

	ranges := []struct {
		min, max uint32
		name     string
	}{
		{0, 255, "1-byte"},
		{256, 65535, "2-byte"},
		{65536, 16777215, "3-byte"},
		{16777216, ^uint32(0), "4-byte"},
	}

	// Generate 2 values from each range to ensure we test all paths
	for i := 0; i < 8; i++ {
		rangeIdx := i % 4
		r := ranges[rangeIdx]

		if r.max == ^uint32(0) {
			// For 4-byte range, use a different strategy to avoid overflow
			// Generate a random value in the upper range
			values[i] = rng.Uint32()
			if values[i] < r.min {
				values[i] = r.min + rng.Uint32()%(1<<16) // Ensure it's in 4-byte range
			}
		} else {
			// For other ranges, generate within the specified bounds
			values[i] = r.min + rng.Uint32()%(r.max-r.min+1)
		}
	}

	// Shuffle the array to randomize the order and test different combinations
	for i := len(values) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		values[i], values[j] = values[j], values[i]
	}

	return values
}

func TestUvar32x4EdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		values [4]uint32
	}{
		{
			name:   "all_zeros",
			values: [4]uint32{0, 0, 0, 0},
		},
		{
			name:   "all_max_1_byte",
			values: [4]uint32{255, 255, 255, 255},
		},
		{
			name:   "all_max_2_byte",
			values: [4]uint32{65535, 65535, 65535, 65535},
		},
		{
			name:   "all_max_3_byte",
			values: [4]uint32{16777215, 16777215, 16777215, 16777215},
		},
		{
			name:   "all_max_4_byte",
			values: [4]uint32{4294967295, 4294967295, 4294967295, 4294967295},
		},
		{
			name:   "mixed_boundaries",
			values: [4]uint32{255, 256, 65535, 65536},
		},
		{
			name:   "powers_of_two",
			values: [4]uint32{1, 256, 65536, 16777216},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				// Write values
				writer := NewBytesWriter(50)
				writer.WriteUvar32x4(tc.values)

				// Read values back
				reader := BytesReader{buf: writer.Bytes()}
				readValues, err := reader.ReadUvar32x4()
				require.NoError(t, err, "Should successfully read values")

				// Verify all values match using testify assertion
				assert.Equal(t, tc.values, readValues, "All values should match exactly")
			},
		)
	}
}
