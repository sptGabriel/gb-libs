package iterables

import (
	"errors"
)

var (
	ErrIndexOutOfBounds              = errors.New("index out of range")
	ErrNoSingleItemFound             = errors.New("no item found")
	ErrMultipleItemsFound            = errors.New("multiple items found")
	ErrPerPageIsNegativeOrZero       = errors.New("perPage argument cannot be negative or zero")
	ErrPaginateStartIndexOutOfBounds = errors.New("pagination start index is out of iterable bounds")
)

// Iterable represents any type that can be iterated over.
// This type alias serves as a generic constraint for functions and methods
// that work with iterable data structures such as slices, arrays, maps, and channels.
type Iterable any

// Filter returns a new slice containing only the elements from s for which
// the predicate function f returns true. The original slice is not modified.
// The returned slice has zero length but may have non-zero capacity.
func Filter[S ~[]E, E Iterable](s S, f func(E) bool) []E {
	s2 := make([]E, 0, len(s))
	for _, e := range s {
		if f(e) {
			s2 = append(s2, e)
		}
	}
	return s2
}

// TakeWhile, applied to a predicate p, returns the elements
// of s that satisfy p.
func TakeWhile[S ~[]E, E Iterable](s S, p func(E) bool) []E {
	for i, item := range s {
		if !p(item) {
			return s[0:i]
		}
	}
	return s
}

// Returns a copy of the first match in s that satisfies f.
// If no match found, returns ErrNoSingleItemFound.
// If multiple items found, returns ErrMultipleItemsFound.
func Single[S ~[]E, E Iterable](s S, f func(E) bool) (E, error) {
	s2 := Filter(s, f)

	if len(s2) == 0 {
		return *new(E), ErrNoSingleItemFound
	}

	if len(s2) > 1 {
		return *new(E), ErrMultipleItemsFound
	}

	return s2[0], nil
}

// Performs a linear search on s to find the first occurrence
// of an item that satisfies the predicate f. Returns a reference
// to the first item that satisfies the predicate, its index in s,
// and ErrMultipleItemsFound if multiple items satisfy the predicate.
// If no item is found, the index is -1 and no error is returned (nil).
func SingleOrDefault[S ~[]E, E Iterable](s S, f func(E) bool) (*E, int, error) {
	s2 := make([]E, 0, len(s))
	s2idx := -1

	for i, e := range s {
		if f(e) {
			s2 = append(s2, e)
			if s2idx == -1 {
				s2idx = i
			}
		}
	}

	if len(s2) == 0 {
		return new(E), s2idx, nil
	}

	if len(s2) > 1 {
		return new(E), s2idx, ErrMultipleItemsFound
	}

	return &s2[0], s2idx, nil
}

// Returns true if every element of the sequence s passes the test
// in the specified predicate p, or if the sequence is empty;
// otherwise, false.
func All[S ~[]E, E Iterable](s S, p func(E) bool) bool {
	for _, item := range s {
		if !p(item) {
			return false
		}
	}
	return true
}

// Returns true if any element of the sequence s passes the test
// in the specified predicate p, or if the sequence is empty;
// otherwise, false.
func Any[S ~[]E, E Iterable](s S, p func(E) bool) bool {
	for _, item := range s {
		if p(item) {
			return true
		}
	}
	return false
}

// Iterates through s applying the function f and returns
// the collection of elements returned by f. Same behavior
// as the .map from JavaScript/TypeScript.
func Map[S ~[]E, E, R Iterable](s S, f func(E) R) []R {
	s2 := make([]R, 0, len(s))
	for _, e := range s {
		s2 = append(s2, f(e))
	}
	return s2
}

// MapWithIndex applies the given function to each element of the slice along with its index,
// returning a new slice containing the results. The function f receives the index and the
// element as parameters and returns a value of type R.
func MapWithIndex[S ~[]E, E, R Iterable](s S, f func(int, E) R) []R {
	s2 := make([]R, 0, len(s))
	for i, e := range s {
		s2 = append(s2, f(i, e))
	}
	return s2
}

// Iterates through s, applying the function f to each element,
// and returns a collection of results. Similar to JavaScript/TypeScript's .map,
// but designed for functions that may return an error.
func TryMap[S ~[]E, E, R Iterable](s S, f func(E) (R, error)) ([]R, error) {
	s2 := make([]R, 0, len(s))
	for _, e := range s {
		r, err := f(e)
		if err != nil {
			return make([]R, 0), err
		}
		s2 = append(s2, r)
	}
	return s2, nil
}

// TryMapWithIndex applies a function to each element of a slice along with its index,
// returning a new slice with the transformed elements or an error if any transformation fails.
// The function f receives the index and element as parameters and returns the transformed
// value and an error. If any invocation of f returns an error, TryMapWithIndex stops
// processing and returns an empty slice along with the error.
func TryMapWithIndex[S ~[]E, E, R Iterable](s S, f func(int, E) (R, error)) ([]R, error) {
	s2 := make([]R, 0, len(s))
	for i, e := range s {
		r, err := f(i, e)
		if err != nil {
			return make([]R, 0), err
		}
		s2 = append(s2, r)
	}
	return s2, nil
}

// Removes the element at index i from s by returning a copy of s.
// If i is out of bounds it returns ErrIndexOutOfBounds.
func RemoveAt[S ~[]E, E Iterable](s S, i int) ([]E, error) {
	if i > len(s)-1 || i < 0 {
		return nil, ErrIndexOutOfBounds
	}
	return append(s[:i], s[i+1:]...), nil
}

// Paginates s using pageNumber and pageSize when the pagination start index will
// be (pageNumber - 1) * pageSize, and the pagination end index will be
// the min value between the (startIndex + pageSize) and length of s.
// If the len of s is zero, it returns s.
// If pageNumber is negative or zero, it returns ErrPerPageIsNegativeOrZero.
// If start index is greater than or equal the length of s, it returns ErrPaginateStartIndexOutOfBounds.
func Paginate[S ~[]E, E Iterable](s S, pageIndex, perPage int) ([]E, error) {
	if len(s) == 0 {
		return s, nil
	}

	// If the pageIndex is zero or negative, the assumed value will be 1 (one)
	if pageIndex <= 0 {
		pageIndex = 0
	}

	if perPage <= 0 {
		return nil, ErrPerPageIsNegativeOrZero
	}

	startIndex := pageIndex * perPage

	if startIndex >= len(s) {
		return nil, ErrPaginateStartIndexOutOfBounds
	}

	endIndex := min(startIndex+perPage, len(s))
	return s[startIndex:endIndex], nil
}

// ForEach iterates over the slice s, applying the function f to each element.
// It does not return any value and does not handle errors.
// Use this when you want to perform side effects for each element in the slice.
func ForEach[S ~[]E, E Iterable](s S, f func(E)) {
	for _, v := range s {
		f(v)
	}
}

// ForEachWithIndex iterates over a slice s and calls function f for each element,
// passing both the index and the element value to f.
func ForEachWithIndex[S ~[]E, E Iterable](s S, f func(int, E)) {
	for i, v := range s {
		f(i, v)
	}
}

// TryForEach iterates over the slice s, applying the function f to each element.
// If f returns an error for any element, TryForEach stops processing and returns the error.
// Use this when you want to perform side effects for each element and handle potential errors.
func TryForEach[S ~[]E, E Iterable](s S, f func(E) error) error {
	for _, v := range s {
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}

// TryForEachWithIndex iterates over a slice of Iterable elements, calling the provided function
// for each element along with its index. If the function returns an error for any element,
// iteration stops immediately and that error is returned. If all elements are processed
// successfully, nil is returned.
func TryForEachWithIndex[S ~[]E, E Iterable](s S, f func(int, E) error) error {
	for i, v := range s {
		if err := f(i, v); err != nil {
			return err
		}
	}
	return nil
}

// SplitByCondition iterates over slice s and returns two slices. The first slice contains all
// elements that matched the predicate p criteria, the second slice contains the remaining elements.
// If s is empty, two empty slices will be returned.
func SplitByCondition[S ~[]E, E Iterable](s S, p func(E) bool) (matched, unmatched []E) {
	if len(s) == 0 {
		return []E{}, []E{}
	}

	for _, element := range s {
		if p(element) {
			matched = append(matched, element)
		} else {
			unmatched = append(unmatched, element)
		}
	}

	return matched, unmatched
}

// Count iterates through s and counts the number of elements
// that satisfy the predicate p.
func Count[Slice ~[]Element, Element Iterable](s Slice, p func(Element) bool) int {
	var count int
	for _, e := range s {
		if p(e) {
			count++
		}
	}
	return count
}
