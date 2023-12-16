package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Slugify takes an input string, drops all non-alphanumeric ASCII characters or
// spaces/dashes and lower cases it.
func Slugify(str string) (string, error) {
	// This works by normalizing the string to Unicode NFD form (which is the
	// decomposed version), and then dropping any combining characters.
	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), str)
	if err != nil {
		return "", err
	}

	// The previous nomalization should have been enough, but for good measure,
	// we drop any non-ascii characters.
	result = strings.Map(func(r rune) rune {
		// Drop all unicode
		if r > unicode.MaxASCII {
			return -1
		}

		// Replace spaces with dash, keep existing dashes.
		if r == ' ' || r == '-' {
			return '-'
		}

		// Keep alpha-numeric characters
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}

		// Drop everything else
		return -1
	}, result)

	return strings.ToLower(result), nil
}

func AppendErrorf(err error, format string, args ...interface{}) *multierror.Error {
	return multierror.Append(err, fmt.Errorf(format, args...))
}

func AppendError(err error, errs ...error) *multierror.Error {
	return multierror.Append(err, errs...)
}
