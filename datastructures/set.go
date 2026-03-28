package datastructures

type Set[E comparable] struct {
	container map[E]struct{}
}

func NewSet[E comparable](items ...E) *Set[E] {
	set := &Set[E]{
		container: map[E]struct{}{},
	}
	for _, item := range items {
		set.Add(item)
	}

	return set
}

func (s *Set[E]) Len() int {
	return len(s.container)
}

func (s *Set[E]) IsEmpty() bool {
	return len(s.container) == 0
}

func (s *Set[E]) Add(e E) {
	s.container[e] = struct{}{}
}

func (s *Set[E]) Remove(e E) {
	delete(s.container, e)
}

func (s *Set[E]) Contains(e E) bool {
	_, ok := s.container[e]
	return ok
}

// Iterates over all items in set applying the function f on each element
func (s *Set[E]) ForEach(f func(e E)) {
	for item := range s.container {
		f(item)
	}
}

func (s *Set[E]) Clone() *Set[E] {
	set := NewSet[E]()
	s.ForEach(func(e E) { set.Add(e) })
	return set
}

func (s *Set[E]) ToSlice() []E {
	result := []E{}
	s.ForEach(func(e E) { result = append(result, e) })
	return result
}

// Returns a new set with all items from both sets
func (s *Set[E]) Union(other *Set[E]) *Set[E] {
	set := s.Clone()
	other.ForEach(func(e E) { set.Add(e) })
	return set
}

// Returns a new set with common items from both sets
func (s *Set[E]) Intersection(other *Set[E]) *Set[E] {
	set := NewSet[E]()
	for item := range s.container {
		if other.Contains(item) {
			set.Add(item)
		}
	}

	return set
}

// Returns a new set of all the items in the current set that are not in the other
func (s *Set[E]) Difference(other *Set[E]) *Set[E] {
	set := NewSet[E]()

	for item := range s.container {
		if !other.Contains(item) {
			set.Add(item)
		}
	}

	return set
}

// Returns all the items that exist in only one of the sets.
// Opposite of intersection.
func (s *Set[E]) SymmetricDifference(other *Set[E]) *Set[E] {
	set := NewSet[E]()

	for item := range s.container {
		if !other.Contains(item) {
			set.Add(item)
		}
	}

	for item := range other.container {
		if !s.Contains(item) {
			set.Add(item)
		}
	}

	return set
}

// Returns whether all the items of the current set exist in the other one
func (s *Set[E]) IsSubset(other *Set[E]) bool {
	intersection := s.Intersection(other)
	return intersection.Len() == s.Len()
}

// Returns whether all the items of the other set exist in the current one
func (s *Set[E]) IsSuperset(other *Set[E]) bool {
	return other.IsSubset(s)
}

// Returns whether the two sets have no item in common
func (s *Set[E]) IsDisjoint(other *Set[E]) bool {
	intersection := s.Intersection(other)
	return intersection.IsEmpty()
}
