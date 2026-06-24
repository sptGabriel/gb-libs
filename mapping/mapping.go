package mapping

import "errors"

type (
	Key   comparable
	Value any
)

var (
	ErrNoSingleItemFound        = errors.New("no item found")
	ErrMultipleItemsFound       = errors.New("multiple items found")
	ErrDuplicateKeysInInversion = errors.New("duplicate keys found during inversion")
)

// Filter return a new map containg only the key value paris from m
// for which the predicate fucntion f returns true
func Filter[K Key, V Value](m map[K]V, f func(K, V) bool) map[K]V {
	m2 := make(map[K]V)
	for k, v := range m {
		if f(k, v) {
			m2[k] = v
		}
	}

	return m2
}

// Single returns the only key-value pair in m that satisfies the predicate f.
// It returns ErrNoSingleItemFound if no items match, or ErrMultipleItemsFound
// if more than one item matches
func Single[K Key, V Value](m map[K]V, f func(K, V) bool) (K, V, error) {
	var (
		key   K
		value V
		found bool
	)

	for k, v := range m {
		if !f(k, v) {
			continue
		}

		if found {
			return *new(K), *new(V), ErrMultipleItemsFound
		}

		key = k
		value = v
		found = true
	}

	if !found {
		return *new(K), *new(V), ErrNoSingleItemFound
	}

	return key, value, nil
}

// SingleOrDefault returns the only key-value pair in m that satisfies the predicate f.
// If no items match, it returns zero values with found set to false and nil error.
// If exactly one item matches, it returns the key-value pair with found set to true.
// If multiple items match, it returns zero values with found set to false and ErrMultipleItemsFound.
func SingleOrDefault[K Key, V Value](m map[K]V, f func(K, V) bool) (K, V, bool, error) {
	var (
		found bool
		key   K
		value V
	)

	for k, v := range m {
		if f(k, v) {
			if found {
				return *new(K), *new(V), false, ErrMultipleItemsFound
			}
			key = k
			value = v
			found = true
		}
	}

	return key, value, found, nil
}

// All returns true if all key-value pairs in m satisfy the predicate p.
// Returns true for an empty map.
func All[K Key, V Value](m map[K]V, p func(K, V) bool) bool {
	for k, v := range m {
		if !p(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one key-value pair in m satisfies the predicate p.
// Returns false for an empty map.
func Any[K Key, V Value](m map[K]V, p func(K, V) bool) bool {
	for k, v := range m {
		if p(k, v) {
			return true
		}
	}
	return false
}

// It returns ErrDuplicateKeysInInversion if any key exists in more than one map.
func TryMerge[K Key, V Value](m ...map[K]V) (map[K]V, error) {
	result := make(map[K]V)
	for _, m2 := range m {
		for k, v := range m2 {
			if _, exists := result[k]; exists {
				return nil, ErrDuplicateKeysInInversion
			}
			result[k] = v
		}
	}
	return result, nil
}

// Invert returns a new map where keys become values and values become keys.
// If multiple keys have the same value, only one will be preserved (non-deterministic).
func Invert[K, V Key](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// TryInvert returns a new map where keys become values and values become keys.
// It returns ErrDuplicateKeysInInversion if multiple keys have the same value.
func TryInvert[K, V Key](m map[K]V) (map[V]K, error) {
	result := make(map[V]K, len(m))
	for k, v := range m {
		if _, exists := result[v]; exists {
			return nil, ErrDuplicateKeysInInversion
		}
		result[v] = k
	}
	return result, nil
}

// FirstOrDefault returns the first key-value pair in m that satisfies the predicate f.
// If no items match, it returns zero values with found set to false.
//
// Note: Since Go maps are unordered, the result may vary between executions.
func FirstOrDefault[K Key, V Value](m map[K]V, f func(K, V) bool) (K, V, bool) {
	for k, v := range m {
		if f(k, v) {
			return k, v, true
		}
	}
	return *new(K), *new(V), false
}

// Count returns the number of key-value pairs in m that satisfy the predicate f.
func Count[K Key, V Value](m map[K]V, f func(K, V) bool) int {
	count := 0
	for k, v := range m {
		if f(k, v) {
			count++
		}
	}
	return count
}

// Merge combines multiple maps into a single map.
// If the same key exists in multiple maps, the value from the later map takes precedence.
func Merge[K Key, V Value](m ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m2 := range m {
		for k, v := range m2 {
			result[k] = v
		}
	}
	return result
}

// TryForEach executes the function f for each key-value pair in m.
// If f returns an error for any pair, TryForEach immediately returns that error.
// Otherwise, it returns nil after processing all pairs.
func TryForEach[K Key, V Value](m map[K]V, f func(K, V) error) error {
	for k, v := range m {
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Keys returns a slice containing all keys from the map m.
func Keys[K Key, V Value](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice containing all values from the map m.
func Values[K Key, V Value](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// First returns the first key-value pair in m that satisfies the predicate f.
// It returns ErrNoSingleItemFound if no items match.
//
// Note: Since Go maps are unordered, the result may vary between executions.
func First[K Key, V Value](m map[K]V, f func(K, V) bool) (K, V, error) {
	for k, v := range m {
		if f(k, v) {
			return k, v, nil
		}
	}
	return *new(K), *new(V), ErrNoSingleItemFound
}

// Map applies the transformation function f to each key-value pair in m
// and returns a new map with the same keys and transformed values.
func Map[K Key, V, R Value](m map[K]V, f func(K, V) R) map[K]R {
	m2 := make(map[K]R)
	for k, v := range m {
		m2[k] = f(k, v)
	}
	return m2
}

// TryMap applies the transformation function f to each key-value pair in m.
// If f returns an error for any pair, TryMap immediately returns nil and that error.
// Otherwise, it returns a new map with the same keys and transformed values.
func TryMap[K Key, V, R Value](m map[K]V, f func(K, V) (R, error)) (map[K]R, error) {
	m2 := make(map[K]R)
	for k, v := range m {
		r, err := f(k, v)
		if err != nil {
			return nil, err
		}
		m2[k] = r
	}
	return m2, nil
}

// ForEach executes the function f for each key-value pair in m.
// The function f is called for its side effects; return values are ignored.
func ForEach[K Key, V Value](m map[K]V, f func(K, V)) {
	for k, v := range m {
		f(k, v)
	}
}
