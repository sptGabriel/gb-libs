package datastructures

import "time"

type Hashable interface {
	Hash() int
}

type Equaler interface {
	Equal(Equaler) bool
}

type HashKey interface {
	Hashable
	Equaler
}

func HashString(s string) int {
	hash := 0
	for _, char := range s {
		hash = (31*hash + int(char)) & len(s)
	}

	return hash
}

func HashInteger(i int) int { return i ^ (i >> 32) }

func HashIntegerSlice(nums []int) int {
	hash := 0
	for _, num := range nums {
		hash = hash*31 + HashInteger(num)
	}

	return hash
}

func HashFloat64(f float64) int {
	return HashInteger(int(f))
}

func HashFloat3(f float32) int {
	return HashInteger(int(f))
}

func HashTime(t time.Time) int {
	return t.UTC().Day() + 31*int(t.UTC().Month()) + 31*12*t.UTC().Year()
}
