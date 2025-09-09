package pkg

import (
	"hash/fnv"
	"strconv"
	"testing"

	b2 "modernc.org/b/v2"
)

// Obj struct for benchmarking
type Obj struct {
	name string
}

// Hash functions for testing
func stringHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func stringEqual(a, b string) bool {
	return a == b
}

func objHash(obj Obj) uint64 {
	h := fnv.New64a()
	h.Write([]byte(obj.name))
	return h.Sum64()
}

func objEqual(a, b Obj) bool {
	return a.name == b.name
}

func uint64Hash(n uint64) uint64 {
	// Simple hash function for uint64
	return n * 11400714819323198485
}

func uint64Equal(a, b uint64) bool {
	return a == b
}

func TestHashMapBasicOperations(t *testing.T) {
	hm := NewHashMap[string, int](stringHash, stringEqual)

	// Test empty map
	if !hm.IsEmpty() {
		t.Error("New map should be empty")
	}
	if hm.Len() != 0 {
		t.Error("New map size should be 0")
	}

	// Test Set and Get
	hm.Set("key1", 100)
	hm.Set("key2", 200)
	hm.Set("key3", 300)

	if hm.Len() != 3 {
		t.Errorf("Expected size 3, got %d", hm.Len())
	}

	if val, ok := hm.Get("key1"); !ok || val != 100 {
		t.Errorf("Expected (100, true), got (%d, %t)", val, ok)
	}

	if val, ok := hm.Get("key2"); !ok || val != 200 {
		t.Errorf("Expected (200, true), got (%d, %t)", val, ok)
	}

	if val, ok := hm.Get("key3"); !ok || val != 300 {
		t.Errorf("Expected (300, true), got (%d, %t)", val, ok)
	}

	// Test non-existent key
	if val, ok := hm.Get("nonexistent"); ok {
		t.Errorf("Expected (0, false) for non-existent key, got (%d, %t)", val, ok)
	}
}

func TestHashMapUpdate(t *testing.T) {
	hm := NewHashMap[string, int](stringHash, stringEqual)

	// Set initial value
	hm.Set("key1", 100)
	if val, ok := hm.Get("key1"); !ok || val != 100 {
		t.Errorf("Expected (100, true), got (%d, %t)", val, ok)
	}

	// Update value
	hm.Set("key1", 500)
	if val, ok := hm.Get("key1"); !ok || val != 500 {
		t.Errorf("Expected (500, true), got (%d, %t)", val, ok)
	}

	// Len should remain same after update
	if hm.Len() != 1 {
		t.Errorf("Expected size 1 after update, got %d", hm.Len())
	}
}

func TestHashMapClear(t *testing.T) {
	hm := NewHashMap[string, int](stringHash, stringEqual)

	// Add some entries
	hm.Set("key1", 100)
	hm.Set("key2", 200)
	hm.Set("key3", 300)

	if hm.Len() != 3 {
		t.Errorf("Expected size 3, got %d", hm.Len())
	}

	// Clear the map
	hm.Clear()

	if !hm.IsEmpty() {
		t.Error("Map should be empty after clear")
	}
	if hm.Len() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", hm.Len())
	}

	// All keys should be gone
	if val, ok := hm.Get("key1"); ok {
		t.Errorf("Expected (0, false) after clear, got (%d, %t)", val, ok)
	}
}

func TestHashMapResize(t *testing.T) {
	hm := NewHashMapWithCapacity[string, int](4, stringHash, stringEqual)

	// Add enough entries to trigger resize
	for i := 0; i < 10; i++ {
		key := "key" + strconv.Itoa(i)
		hm.Set(key, i*100)
	}

	// Verify all entries are still accessible
	for i := 0; i < 10; i++ {
		key := "key" + strconv.Itoa(i)
		if val, ok := hm.Get(key); !ok || val != i*100 {
			t.Errorf("Expected (%d, true) for key %s, got (%d, %t)", i*100, key, val, ok)
		}
	}

	if hm.Len() != 10 {
		t.Errorf("Expected size 10, got %d", hm.Len())
	}
}

func TestHashMapCollisions(t *testing.T) {
	// Use a hash function that creates collisions
	badHash := func(s string) uint64 {
		return 42 // Always same hash
	}

	hm := NewHashMapWithCapacity[string, int](8, badHash, stringEqual)

	// Add multiple entries with same hash
	keys := []string{"a", "b", "c", "d", "e"}
	for i, key := range keys {
		hm.Set(key, i*100)
	}

	// Verify all entries are accessible despite collisions
	for i, key := range keys {
		if val, ok := hm.Get(key); !ok || val != i*100 {
			t.Errorf("Expected (%d, true) for key %s, got (%d, %t)", i*100, key, val, ok)
		}
	}
}

func TestHashMapWithUint64Keys(t *testing.T) {
	hm := NewHashMap[uint64, string](uint64Hash, uint64Equal)

	// Test with uint64 keys
	hm.Set(1, "one")
	hm.Set(2, "two")
	hm.Set(3, "three")

	if val, ok := hm.Get(1); !ok || val != "one" {
		t.Errorf("Expected ('one', true), got ('%s', %t)", val, ok)
	}
	if val, ok := hm.Get(2); !ok || val != "two" {
		t.Errorf("Expected ('two', true), got ('%s', %t)", val, ok)
	}
	if val, ok := hm.Get(3); !ok || val != "three" {
		t.Errorf("Expected ('three', true), got ('%s', %t)", val, ok)
	}
}

// Benchmarks comparing HashMap with built-in map
const benchmarkSize = 10000

func BenchmarkHashMapSet(b *testing.B) {
	hm := NewHashMap[string, uint64](stringHash, stringEqual)
	keys := make([]string, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		hm.Set(key, uint64(i))
	}
}

func BenchmarkBuiltinMapSet(b *testing.B) {
	m := make(map[string]uint64)
	keys := make([]string, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		m[key] = uint64(i)
	}
}

func BenchmarkHashMapGet(b *testing.B) {
	hm := NewHashMap[string, uint64](stringHash, stringEqual)
	keys := make([]string, benchmarkSize)

	// Prepare data
	for i := 0; i < benchmarkSize; i++ {
		key := "key" + strconv.Itoa(i)
		keys[i] = key
		hm.Set(key, uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		_, _ = hm.Get(key)
	}
}

func BenchmarkBuiltinMapGet(b *testing.B) {
	m := make(map[string]uint64)
	keys := make([]string, benchmarkSize)

	// Prepare data
	for i := 0; i < benchmarkSize; i++ {
		key := "key" + strconv.Itoa(i)
		keys[i] = key
		m[key] = uint64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		_, _ = m[key]
	}
}

func BenchmarkHashMapSetGet(b *testing.B) {
	hm := NewHashMap[string, uint64](stringHash, stringEqual)
	keys := make([]string, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		hm.Set(key, uint64(i))
		_, _ = hm.Get(key)
	}
}

func BenchmarkBuiltinMapSetGet(b *testing.B) {
	m := make(map[string]uint64)
	keys := make([]string, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		m[key] = uint64(i)
		_, _ = m[key]
	}
}

// Benchmark with different sizes
func BenchmarkHashMapSetSmall(b *testing.B) {
	benchmarkHashMapSetWithSize(b, 100)
}

func BenchmarkHashMapSetMedium(b *testing.B) {
	benchmarkHashMapSetWithSize(b, 1000)
}

func BenchmarkHashMapSetLarge(b *testing.B) {
	benchmarkHashMapSetWithSize(b, 100000)
}

func benchmarkHashMapSetWithSize(b *testing.B, size int) {
	hm := NewHashMap[string, uint64](stringHash, stringEqual)
	keys := make([]string, size)

	for i := 0; i < size; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		hm.Set(key, uint64(i))
	}
}

func BenchmarkBuiltinMapSetSmall(b *testing.B) {
	benchmarkBuiltinMapSetWithSize(b, 100)
}

func BenchmarkBuiltinMapSetMedium(b *testing.B) {
	benchmarkBuiltinMapSetWithSize(b, 1000)
}

func BenchmarkBuiltinMapSetLarge(b *testing.B) {
	benchmarkBuiltinMapSetWithSize(b, 100000)
}

func benchmarkBuiltinMapSetWithSize(b *testing.B, size int) {
	m := make(map[string]uint64)
	keys := make([]string, size)

	for i := 0; i < size; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		m[key] = uint64(i)
	}
}

// Tree benchmarks using modernc.org/b/v2
func BenchmarkTreeSet(b *testing.B) {
	tree := b2.TreeNew[string, uint64](
		func(a, b string) int {
			if a < b {
				return -1
			} else if a > b {
				return 1
			}
			return 0
		},
	)
	keys := make([]string, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		tree.Set(key, uint64(i))
	}
}

func BenchmarkTreeGet(b *testing.B) {
	tree := b2.TreeNew[string, uint64](
		func(a, b string) int {
			if a < b {
				return -1
			} else if a > b {
				return 1
			}
			return 0
		},
	)
	keys := make([]string, benchmarkSize)

	// Prepare data
	for i := 0; i < benchmarkSize; i++ {
		key := "key" + strconv.Itoa(i)
		keys[i] = key
		tree.Set(key, uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		_, _ = tree.Get(key)
	}
}

func BenchmarkTreeSetGet(b *testing.B) {
	tree := b2.TreeNew[string, uint64](
		func(a, b string) int {
			if a < b {
				return -1
			} else if a > b {
				return 1
			}
			return 0
		},
	)
	keys := make([]string, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		tree.Set(key, uint64(i))
		_, _ = tree.Get(key)
	}
}

// Tree benchmarks with different sizes
func BenchmarkTreeSetSmall(b *testing.B) {
	benchmarkTreeSetWithSize(b, 100)
}

func BenchmarkTreeSetMedium(b *testing.B) {
	benchmarkTreeSetWithSize(b, 1000)
}

func BenchmarkTreeSetLarge(b *testing.B) {
	benchmarkTreeSetWithSize(b, 100000)
}

func benchmarkTreeSetWithSize(b *testing.B, size int) {
	tree := b2.TreeNew[string, uint64](
		func(a, b string) int {
			if a < b {
				return -1
			} else if a > b {
				return 1
			}
			return 0
		},
	)
	keys := make([]string, size)

	for i := 0; i < size; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		tree.Set(key, uint64(i))
	}
}

func TestHashMapObjBasicOperations(t *testing.T) {
	hm := NewHashMap[Obj, uint64](objHash, objEqual)

	// Test empty map
	if !hm.IsEmpty() {
		t.Error("New map should be empty")
	}
	if hm.Len() != 0 {
		t.Error("New map size should be 0")
	}

	// Test Set and Get
	hm.Set(Obj{"key1"}, 100)
	hm.Set(Obj{"key2"}, 200)
	hm.Set(Obj{"key3"}, 300)

	if hm.Len() != 3 {
		t.Errorf("Expected size 3, got %d", hm.Len())
	}

	if val, ok := hm.Get(Obj{"key1"}); !ok || val != 100 {
		t.Errorf("Expected (100, true), got (%d, %t)", val, ok)
	}

	if val, ok := hm.Get(Obj{"key2"}); !ok || val != 200 {
		t.Errorf("Expected (200, true), got (%d, %t)", val, ok)
	}

	if val, ok := hm.Get(Obj{"key3"}); !ok || val != 300 {
		t.Errorf("Expected (300, true), got (%d, %t)", val, ok)
	}

	// Test non-existent key
	if val, ok := hm.Get(Obj{"nonexistent"}); ok {
		t.Errorf("Expected (0, false) for non-existent key, got (%d, %t)", val, ok)
	}
}

func BenchmarkHashMapSetObj(b *testing.B) {
	hm := NewHashMap[Obj, uint64](objHash, objEqual)
	keys := make([]Obj, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = Obj{"key" + strconv.Itoa(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		hm.Set(key, uint64(i))
	}
}

func BenchmarkHashMapGetObj(b *testing.B) {
	hm := NewHashMap[Obj, uint64](objHash, objEqual)
	keys := make([]Obj, benchmarkSize)

	// Prepare data
	for i := 0; i < benchmarkSize; i++ {
		key := Obj{"key" + strconv.Itoa(i)}
		keys[i] = key
		hm.Set(key, uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		_, _ = hm.Get(key)
	}
}

func BenchmarkHashMapSetGetObj(b *testing.B) {
	hm := NewHashMap[Obj, uint64](objHash, objEqual)
	keys := make([]Obj, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = Obj{"key" + strconv.Itoa(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		hm.Set(key, uint64(i))
		_, _ = hm.Get(key)
	}
}

// Tree benchmarks using modernc.org/b/v2 with Obj keys
func BenchmarkTreeSetObj(b *testing.B) {
	tree := b2.TreeNew[Obj, uint64](
		func(a, b Obj) int {
			if a.name < b.name {
				return -1
			} else if a.name > b.name {
				return 1
			}
			return 0
		},
	)
	keys := make([]Obj, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = Obj{"key" + strconv.Itoa(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		tree.Set(key, uint64(i))
	}
}

func BenchmarkTreeGetObj(b *testing.B) {
	tree := b2.TreeNew[Obj, uint64](
		func(a, b Obj) int {
			if a.name < b.name {
				return -1
			} else if a.name > b.name {
				return 1
			}
			return 0
		},
	)
	keys := make([]Obj, benchmarkSize)

	// Prepare data
	for i := 0; i < benchmarkSize; i++ {
		key := Obj{"key" + strconv.Itoa(i)}
		keys[i] = key
		tree.Set(key, uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		_, _ = tree.Get(key)
	}
}

func BenchmarkTreeSetGetObj(b *testing.B) {
	tree := b2.TreeNew[Obj, uint64](
		func(a, b Obj) int {
			if a.name < b.name {
				return -1
			} else if a.name > b.name {
				return 1
			}
			return 0
		},
	)
	keys := make([]Obj, benchmarkSize)

	// Prepare keys
	for i := 0; i < benchmarkSize; i++ {
		keys[i] = Obj{"key" + strconv.Itoa(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%benchmarkSize]
		tree.Set(key, uint64(i))
		_, _ = tree.Get(key)
	}
}

// Tree benchmarks with different sizes using Obj keys
func BenchmarkTreeSetSmallObj(b *testing.B) {
	benchmarkTreeSetWithSizeObj(b, 100)
}

func BenchmarkTreeSetMediumObj(b *testing.B) {
	benchmarkTreeSetWithSizeObj(b, 1000)
}

func BenchmarkTreeSetLargeObj(b *testing.B) {
	benchmarkTreeSetWithSizeObj(b, 100000)
}

func benchmarkTreeSetWithSizeObj(b *testing.B, size int) {
	tree := b2.TreeNew[Obj, uint64](
		func(a, b Obj) int {
			if a.name < b.name {
				return -1
			} else if a.name > b.name {
				return 1
			}
			return 0
		},
	)
	keys := make([]Obj, size)

	for i := 0; i < size; i++ {
		keys[i] = Obj{"key" + strconv.Itoa(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		tree.Set(key, uint64(i))
	}
}
