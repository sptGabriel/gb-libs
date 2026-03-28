package datastructures

// See HashMap implementation for more details.
type HashSet[E HashKey] struct {
	values *HashMap[E, struct{}]
}

func NewHashSet[E HashKey](items ...E) *HashSet[E] {
	kvps := []KVPair[E, struct{}]{}
	for _, item := range items {
		kvps = append(kvps, NewKVP(item, struct{}{}))
	}

	set := &HashSet[E]{
		values: NewHashMap(kvps...),
	}

	return set
}

func (s *HashSet[E]) Len() int {
	return s.values.Len()
}

func (s *HashSet[E]) IsEmpty() bool {
	return s.Len() == 0
}

func (s *HashSet[E]) Add(e E) {
	s.values.Set(e, struct{}{})
}

func (s *HashSet[E]) Remove(e E) {
	s.values.Delete(e)
}

func (s *HashSet[E]) Contains(e E) bool {
	_, ok := s.values.Get(e)
	return ok
}

// Iterates over all items in set applying the function f on each element
func (s *HashSet[E]) ForEach(f func(e E)) {
	for _, item := range s.values.Keys() {
		f(item)
	}
}

func (s *HashSet[E]) Clone() *HashSet[E] {
	return NewHashSet(s.values.Keys()...)
}

func (s *HashSet[E]) ToSlice() []E {
	result := []E{}
	s.ForEach(func(e E) { result = append(result, e) })
	return result
}

// Returns a new set with all items from both sets
func (s *HashSet[E]) Union(other *HashSet[E]) *HashSet[E] {
	set := s.Clone()
	other.ForEach(func(e E) { set.Add(e) })
	return set
}

// Returns a new set with common items from both sets
func (s *HashSet[E]) Intersection(other *HashSet[E]) *HashSet[E] {
	set := NewHashSet[E]()
	for _, item := range s.values.Keys() {
		if other.Contains(item) {
			set.Add(item)
		}
	}
	return set
}

// Returns a new set of all the items in the current set that are not in the other
func (s *HashSet[E]) Difference(other *HashSet[E]) *HashSet[E] {
	set := NewHashSet[E]()
	for _, item := range s.values.Keys() {
		if !other.Contains(item) {
			set.Add(item)
		}
	}
	return set
}

// Returns all the items that exist in only one of the sets,
// Opposite of intersection.
func (s *HashSet[E]) SymmetricDifference(other *HashSet[E]) *HashSet[E] {
	set := NewHashSet[E]()

	for _, item := range s.values.Keys() {
		if !other.Contains(item) {
			set.Add(item)
		}
	}

	for _, item := range other.values.Keys() {
		if !s.Contains(item) {
			set.Add(item)
		}
	}

	return set
}

// Returns whether all the items of the current set exist in the other one
func (s *HashSet[E]) IsSubset(other *HashSet[E]) bool {
	intersection := s.Intersection(other)
	return intersection.Len() == s.Len()
}

// Returns whether all the items of the other set exist in the current one
func (s *HashSet[E]) IsSuperset(other *HashSet[E]) bool {
	return other.IsSubset(s)
}

// Returns whether the two sets have no item in common
func (s *HashSet[E]) IsDisjoint(other *HashSet[E]) bool {
	intersection := s.Intersection(other)
	return intersection.IsEmpty()
}
