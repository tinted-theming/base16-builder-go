package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func ToAscii(str string) (string, error) {
	// This works by normalizing the string to Unicode NFD form (which is the
	// decomposed version), and then dropping any combining characters.
	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), str)
	if err != nil {
		return "", err
	}

	// The previous nomalization should have been enough, butfFor good measure,
	// we drop any non-ascii characters.
	return strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, result), nil
}
