package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var colorRegex = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

// color is a small utility type which parses colors in html format and drops
// them into a type which we can use for some basic conversions. It also adds
// UnmarshalYAML so it can be parsed directly by the yaml parser.
type color struct {
	R, G, B int
}

// UnmarshalYAML implements yaml.Unmarshaler
func (c *color) UnmarshalYAML(f func(interface{}) error) error {
	var in string
	var tmp int64
	err := f(&in)
	if err != nil {
		return err
	}

	if !colorRegex.MatchString(in) {
		return fmt.Errorf("Color %q is not formatted correctly", in)
	}

	in = strings.TrimPrefix(in, "#")

	tmp, err = strconv.ParseInt(in[0:2], 16, 32)
	if err != nil {
		return err
	}
	c.R = int(tmp)

	tmp, err = strconv.ParseInt(in[2:4], 16, 32)
	if err != nil {
		return err
	}
	c.G = int(tmp)

	tmp, err = strconv.ParseInt(in[4:6], 16, 32)
	if err != nil {
		return err
	}
	c.B = int(tmp)

	return err
}
