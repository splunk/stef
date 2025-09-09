package pkg

import (
	"fmt"
)

// HashFunc is a function type that computes hash for a key
type HashFunc[K any] func(K) uint64

// EqualFunc is a function type that checks equality between two keys
type EqualFunc[K any] func(K, K) bool

// entry represents a slot in the hash table
type entry[K any, V any] struct {
	key     K
	value   V
	deleted bool
	empty   bool
}

// HashMap implements a generic hash map using open addressing with linear probing
type HashMap[K any, V any] struct {
	buckets    []entry[K, V]
	size       int
	capacity   int
	hashFunc   HashFunc[K]
	equalFunc  EqualFunc[K]
	loadFactor float64
}

const (
	defaultCapacity   = 16
	defaultLoadFactor = 0.75
	maxLoadFactor     = 0.9
)

// NewHashMap creates a new HashMap with the given hash and equality functions
func NewHashMap[K any, V any](hashFunc HashFunc[K], equalFunc EqualFunc[K]) *HashMap[K, V] {
	return NewHashMapWithCapacity[K, V](defaultCapacity, hashFunc, equalFunc)
}

// NewHashMapWithCapacity creates a new HashMap with specified initial capacity
func NewHashMapWithCapacity[K any, V any](capacity int, hashFunc HashFunc[K], equalFunc EqualFunc[K]) *HashMap[K, V] {
	if capacity < 1 {
		capacity = defaultCapacity
	}

	// Ensure capacity is power of 2 for better performance
	capacity = nextPowerOfTwo(capacity)

	buckets := make([]entry[K, V], capacity)
	for i := range buckets {
		buckets[i].empty = true
	}

	return &HashMap[K, V]{
		buckets:    buckets,
		size:       0,
		capacity:   capacity,
		hashFunc:   hashFunc,
		equalFunc:  equalFunc,
		loadFactor: defaultLoadFactor,
	}
}

// nextPowerOfTwo returns the next power of 2 greater than or equal to n
func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}

// hash computes the hash index for a key
func (hm *HashMap[K, V]) hash(key K) uint64 {
	return hm.hashFunc(key) & uint64(hm.capacity-1)
}

// findSlot finds the slot for a given key using linear probing
func (hm *HashMap[K, V]) findSlot(key K) (int, bool) {
	index := int(hm.hash(key))
	startIndex := index

	for {
		entry := &hm.buckets[index]

		// Found empty slot (never used)
		if entry.empty {
			return index, false
		}

		// Found deleted slot or matching key
		if entry.deleted || hm.equalFunc(entry.key, key) {
			return index, !entry.deleted && hm.equalFunc(entry.key, key)
		}

		// Linear probing
		index = (index + 1) % hm.capacity

		// Full circle - should not happen if load factor is maintained
		if index == startIndex {
			break
		}
	}

	return -1, false
}

// resize grows the hash map when load factor exceeds threshold
func (hm *HashMap[K, V]) resize() {
	oldBuckets := hm.buckets
	oldCapacity := hm.capacity

	hm.capacity = hm.capacity * 2
	hm.buckets = make([]entry[K, V], hm.capacity)
	for i := range hm.buckets {
		hm.buckets[i].empty = true
	}
	hm.size = 0

	// Rehash all existing entries
	for i := 0; i < oldCapacity; i++ {
		entry := &oldBuckets[i]
		if !entry.empty && !entry.deleted {
			hm.Set(entry.key, entry.value)
		}
	}
}

// Set inserts or updates a key-value pair
func (hm *HashMap[K, V]) Set(key K, value V) {
	// Check if resize is needed
	if float64(hm.size+1)/float64(hm.capacity) > hm.loadFactor {
		hm.resize()
	}

	index, found := hm.findSlot(key)
	if index == -1 {
		panic("HashMap is full - this should not happen")
	}

	entry := &hm.buckets[index]

	if found {
		// Update existing key
		entry.value = value
	} else {
		// Insert new key
		entry.key = key
		entry.value = value
		entry.deleted = false
		entry.empty = false
		hm.size++
	}
}

// Get retrieves the value for a given key
func (hm *HashMap[K, V]) Get(key K) (V, bool) {
	var zero V

	index, found := hm.findSlot(key)
	if index == -1 || !found {
		return zero, false
	}

	entry := &hm.buckets[index]
	return entry.value, true
}

// Delete removes a key-value pair
func (hm *HashMap[K, V]) Delete(key K) bool {
	index, found := hm.findSlot(key)
	if index == -1 || !found {
		return false
	}

	entry := &hm.buckets[index]
	entry.deleted = true
	hm.size--
	return true
}

// Clear removes all key-value pairs
func (hm *HashMap[K, V]) Clear() {
	for i := range hm.buckets {
		hm.buckets[i].empty = true
		hm.buckets[i].deleted = false
	}
	hm.size = 0
}

// Size returns the number of key-value pairs
func (hm *HashMap[K, V]) Size() int {
	return hm.size
}

// IsEmpty returns true if the map is empty
func (hm *HashMap[K, V]) IsEmpty() bool {
	return hm.size == 0
}

// String returns a string representation of the HashMap
func (hm *HashMap[K, V]) String() string {
	return fmt.Sprintf(
		"HashMap{size: %d, capacity: %d, loadFactor: %.2f}",
		hm.size, hm.capacity, float64(hm.size)/float64(hm.capacity),
	)
}
