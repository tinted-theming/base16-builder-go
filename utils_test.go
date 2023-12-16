package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	var testCases = []struct {
		Input  string
		Output string
	}{
		{
			Input:  "Hello World",
			Output: "hello-world",
		},
		{
			Input:  "Rosé Pine",
			Output: "rose-pine",
		},
	}

	for _, testCase := range testCases {
		ret, err := Slugify(testCase.Input)
		assert.NoError(t, err)
		assert.Equal(t, testCase.Output, ret)
	}
}
