package xtext

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/sptGabriel/libs/text"
)

func TestRemoveDiacritics(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{"accent", "São", "Sao"},
		{"accent 2", "São Paulo", "Sao Paulo"},
		{"accent 3", "Invisível", "Invisivel"},
		{"cedilla", "Manhuaçu", "Manhuacu"},
		{"cedilla 2", "Paçoca", "Pacoca"},
		{"cedilla 3", "Açougue", "Acougue"},
		{"cedilla with accent", "Guaraçá", "Guaraca"},
		{"cedilla with accent 2", "Jacaré-açu", "Jacare-acu"},
		{"cedilla with accent 3", "Aço Inoxidável", "Aco Inoxidavel"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := text.RemoveDiacritics(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
