package iterables

import (
	"iter"
	"slices"
)

// Seq converts a slice to a lazy iter.Seq[E] iterator (Go 1.23).
func Seq[S ~[]E, E any](s S) iter.Seq[E] {
	return slices.Values(s)
}

// Collect materializes an iter.Seq[E] into a slice.
func Collect[E any](seq iter.Seq[E]) []E {
	return slices.Collect(seq)
}

// FilterSeq returns a lazy iterator of elements from seq that satisfy f.
// No intermediate slice is allocated.
func FilterSeq[E any](seq iter.Seq[E], f func(E) bool) iter.Seq[E] {
	return func(yield func(E) bool) {
		for e := range seq {
			if f(e) && !yield(e) {
				return
			}
		}
	}
}

// MapSeq transforms each element of seq using f, returning a new lazy iterator.
// No intermediate slice is allocated.
func MapSeq[E, R any](seq iter.Seq[E], f func(E) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		for e := range seq {
			if !yield(f(e)) {
				return
			}
		}
	}
}

// TakeSeq returns a lazy iterator of the first n elements of seq.
func TakeSeq[E any](seq iter.Seq[E], n int) iter.Seq[E] {
	return func(yield func(E) bool) {
		i := 0
		for e := range seq {
			if i >= n || !yield(e) {
				return
			}
			i++
		}
	}
}

// ReduceSeq applies f cumulatively to elements of seq, starting with initial.
func ReduceSeq[E, R any](seq iter.Seq[E], initial R, f func(R, E) R) R {
	acc := initial
	for e := range seq {
		acc = f(acc, e)
	}
	return acc
}
