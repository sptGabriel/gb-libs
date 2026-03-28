package datastructures

import (
	"math"
)

type KVPair[K HashKey, V any] struct {
	Key   K
	Value V
}

func NewKVP[K HashKey, V any](key K, val V) KVPair[K, V] {
	return KVPair[K, V]{Key: key, Value: val}
}

// HashTable implementation that uses linear probing.
//
// Its intended use is to provide fine-grained control of equality
// comparison of complex structures where the default Go's map
// implementation lacks (e.g. you want to compare equality with
// only a selection of fields of a struct).
//
// (Source: Sedgewick, Algorithms - 4th edition, chapter 3)
type HashMap[K HashKey, V any] struct {
	values []*V
	keys   []*K
	size   int
}

func NewHashMap[K HashKey, V any](items ...KVPair[K, V]) *HashMap[K, V] {
	var hmap *HashMap[K, V]
	if len(items) == 0 {
		hmap = newHashMap[K, V](16)
	} else {
		hmap = newHashMap[K, V](int(float64(len(items)) * 1.5))
	}

	for _, kvp := range items {
		hmap.Set(kvp.Key, kvp.Value)
	}

	return hmap
}

func newHashMap[K HashKey, V any](capacity int) *HashMap[K, V] {
	return &HashMap[K, V]{
		values: make([]*V, capacity),
		keys:   make([]*K, capacity),
		size:   0,
	}
}

func (m *HashMap[K, V]) Len() int {
	return m.size
}

func (m *HashMap[K, V]) IsEmpty() bool {
	return m.Len() == 0
}

// Returns the object and a boolean indicating whether or
// not the key was found.
func (m *HashMap[K, V]) Get(key K) (V, bool) {
	i := m.hash(key)
	k := m.keys[i]
	ok := k != nil
	for ok {
		if (*k).Equal(key) {
			return *m.values[i], true
		}

		i = (i + 1) % m.capacity()
		k = m.keys[i]
		ok = k != nil
	}

	return *new(V), false
}

func (m *HashMap[K, V]) Set(key K, value V) {
	if m.Len() >= m.capacity()/2 {
		m.resize(2 * m.capacity())
	}

	i := m.hash(key)
	k := m.keys[i]
	ok := k != nil
	for ok {
		if (*k).Equal(key) {
			m.values[i] = &value
			return
		}

		i = (i + 1) % m.capacity()
		k = m.keys[i]
		ok = k != nil
	}

	m.keys[i] = &key
	m.values[i] = &value
	m.size++
}

func (m *HashMap[K, V]) Delete(key K) {
	i := m.hash(key)
	k := m.keys[i]
	if k == nil {
		return
	}

	for !key.Equal(*k) {
		i = (i + 1) % m.capacity()
		k = m.keys[i]
	}

	m.keys[i] = nil
	m.values[i] = nil
	i = (i + 1) % m.capacity()

	for m.keys[i] != nil {
		keyToRedo := *m.keys[i]
		valToRedo := *m.values[i]
		m.keys[i] = nil
		m.values[i] = nil
		m.size--
		m.Set(keyToRedo, valToRedo)
		i = (i + 1) % m.capacity()
	}

	m.size--
	if m.Len() > 0 && m.Len() == m.capacity()/8 {
		m.resize(m.capacity() / 2)
	}
}

func (m *HashMap[K, V]) Keys() []K {
	keys := []K{}
	for _, key := range m.keys {
		if key == nil {
			continue
		}
		keys = append(keys, *key)
	}
	return keys
}

func (m *HashMap[K, V]) Values() []V {
	values := []V{}
	for _, value := range m.values {
		if value == nil {
			continue
		}
		values = append(values, *value)
	}
	return values
}

func (m *HashMap[K, V]) KeyValuePairs() []KVPair[K, V] {
	kvps := []KVPair[K, V]{}
	for _, key := range m.Keys() {
		value, _ := m.Get(key)
		kvps = append(kvps, KVPair[K, V]{Key: key, Value: value})
	}
	return kvps
}

func (m *HashMap[K, V]) Clone() *HashMap[K, V] {
	newMap := newHashMap[K, V](m.capacity())
	for _, key := range m.Keys() {
		value, _ := m.Get(key)
		newMap.Set(key, value)
	}
	return newMap
}

func (m *HashMap[K, V]) hash(key K) int {
	return key.Hash() & math.MaxInt % m.capacity()
}

func (m *HashMap[K, V]) capacity() int {
	return cap(m.values)
}

func (m *HashMap[K, V]) resize(capacity int) {
	newMap := newHashMap[K, V](capacity)
	for i := 0; i < m.capacity(); i++ {
		if m.keys[i] != nil {
			newMap.Set(*m.keys[i], *m.values[i])
		}
	}
	*m = *newMap
}
