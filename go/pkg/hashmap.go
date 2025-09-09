package pkg

import (
	"fmt"
)

// HashFunc is a function type that computes hash for a key
type HashFunc[K any] func(K) uint64

// EqualFunc is a function type that checks equality between two keys
type EqualFunc[K any] func(K, K) bool

// entryType represents a slot in the hash table
type entryType[K any, V any] struct {
	key    K
	value  V
	exists bool
}

// HashMap implements a generic hash map using open addressing with linear probing
type HashMap[K any, V any] struct {
	buckets      []entryType[K, V]
	size         int
	capacity     int
	hashFunc     HashFunc[K]
	equalFunc    EqualFunc[K]
	maxElemCount int
	//loadFactor float64
}

const (
	defaultCapacity   = 16
	defaultLoadFactor = 0.75
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

	buckets := make([]entryType[K, V], capacity)

	return &HashMap[K, V]{
		buckets:      buckets,
		size:         0,
		capacity:     capacity,
		hashFunc:     hashFunc,
		equalFunc:    equalFunc,
		maxElemCount: int(float64(capacity) * defaultLoadFactor),
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
func (m *HashMap[K, V]) hash(key K) uint64 {
	return m.hashFunc(key) & uint64(m.capacity-1)
}

// findSlot finds the slot for a given key using linear probing
func (m *HashMap[K, V]) findSlot(key K) (int, bool) {
	index := int(m.hash(key))

	for {
		entry := &m.buckets[index]

		if !entry.exists {
			// Found empty slot (never used)
			return index, false
		}

		if m.equalFunc(entry.key, key) {
			return index, true
		}

		// Linear probing
		index = (index + 1) & (m.capacity - 1)
	}
}

// resize grows the hash map when load factor exceeds threshold
func (m *HashMap[K, V]) resize() {
	oldBuckets := m.buckets
	oldCapacity := m.capacity

	m.capacity = m.capacity * 2
	m.buckets = make([]entryType[K, V], m.capacity)
	m.size = 0
	m.maxElemCount = int(float64(m.capacity) * defaultLoadFactor)

	// Rehash all existing entries
	for i := 0; i < oldCapacity; i++ {
		entry := &oldBuckets[i]
		if entry.exists {
			m.Set(entry.key, entry.value)
		}
	}
}

// Set inserts or updates a key-value pair
func (m *HashMap[K, V]) Set(key K, value V) {
	// Check if resize is needed
	if m.size+1 > m.maxElemCount {
		m.resize()
	}

	index, found := m.findSlot(key)

	entry := &m.buckets[index]

	if found {
		// Update existing key
		entry.value = value
	} else {
		// Insert new key
		entry.key = key
		entry.value = value
		entry.exists = true
		m.size++
	}
}

// Get retrieves the value for a given key
func (m *HashMap[K, V]) Get(key K) (V, bool) {
	var zero V

	index, found := m.findSlot(key)
	if !found {
		return zero, false
	}

	return m.buckets[index].value, true
}

// Clear removes all key-value pairs
func (m *HashMap[K, V]) Clear() {
	for i := range m.buckets {
		m.buckets[i].exists = false
	}
	m.size = 0
}

// Len returns the number of key-value pairs
func (m *HashMap[K, V]) Len() int {
	return m.size
}

// IsEmpty returns true if the map is empty
func (m *HashMap[K, V]) IsEmpty() bool {
	return m.size == 0
}

// String returns a string representation of the HashMap
func (m *HashMap[K, V]) String() string {
	return fmt.Sprintf(
		"HashMap{size: %d, capacity: %d, loadFactor: %.2f}",
		m.size, m.capacity, float64(m.size)/float64(m.capacity),
	)
}
