package validation

import (
	"cmp"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/sptGabriel/gb-libs/iterables"
)

var ErrInvalidFileSignature = errors.New("invalid file signature")

type allowedExtension string

const (
	pdfExtension  allowedExtension = ".pdf"
	xlsExtension  allowedExtension = ".xls"
	csvExtension  allowedExtension = ".csv"
	bmpExtension  allowedExtension = ".bmp"
	jpgExtension  allowedExtension = ".jpg"
	pngExtension  allowedExtension = ".png"
	xlsxExtension allowedExtension = ".xlsx"
)

type mimetypeHeader string

type fileSignature struct {
	header mimetypeHeader
	values [][]byte
}

var fileSignatures = map[allowedExtension]fileSignature{
	pdfExtension: {
		header: "application/pdf",
		values: [][]byte{
			{'%', 'P', 'D', 'F', '-'},
		},
	},
	xlsExtension: {
		header: "application/vnd.ms-excel",
		values: [][]byte{
			{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1},
		},
	},
	xlsxExtension: {
		header: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		values: [][]byte{
			{0x50, 0x4B, 0x03, 0x04},
		},
	},
	pngExtension: {
		header: "image/png",
		values: [][]byte{
			{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		},
	},
	jpgExtension: {
		header: "image/jpeg",
		values: [][]byte{
			{0xFF, 0xD8, 0xFF, 0xE0},
			{0xFF, 0xD8, 0xFF, 0xDB},
			{0xFF, 0xD8, 0xFF, 0xEE},
			{0xFF, 0xD8, 0xFF, 0xE1},
		},
	},
	csvExtension: {
		header: "text/csv",
		values: [][]byte{},
	},
	bmpExtension: {
		header: "image/bmp",
		values: [][]byte{
			{0x42, 0x4D},
		},
	},
}

func ValidateFile(name string, content []byte) error {
	strExtension := filepath.Ext(name)
	if strExtension == "" {
		return fmt.Errorf("invalid file extension in '%s'", name)
	}

	values, exists := fileSignatures[allowedExtension(strExtension)]
	if !exists {
		return fmt.Errorf("invalid file extension '%s'", strExtension)
	}

	signatures := values.values

	if len(signatures) == 0 {
		return nil
	}

	headerBytesLen := len(
		slices.MaxFunc(signatures, func(a, b []byte) int {
			return cmp.Compare(len(a), len(b))
		}),
	)

	if len(content) < headerBytesLen {
		return ErrInvalidFileSignature
	}

	headerBytes := content[:headerBytesLen]

	isValid := iterables.Any(signatures, func(s []byte) bool {
		return slices.Equal(headerBytes[:len(s)], s)
	})

	if !isValid {
		return ErrInvalidFileSignature
	}

	return nil
}
