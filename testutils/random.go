package testutils

import "math/rand"

var (
	upperCasePool = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	runePool      = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

func RandomStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runePool[rand.Intn(len(upperCasePool))]
	}

	return string(b)
}

func RandomLowercaseStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runePool[rand.Intn(len(runePool))]
	}

	return string(b)
}
