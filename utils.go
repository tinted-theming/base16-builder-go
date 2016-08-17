package main

import colorful "github.com/lucasb-eyer/go-colorful"

// color is a simple wrapper around colorful.Color which adds
// UnmarshalYAML
type color struct {
	colorful.Color
}

// UnmarshalYAML implements yaml.Unmarshaler
func (c *color) UnmarshalYAML(f func(interface{}) error) error {
	var in string
	err := f(&in)
	if err != nil {
		return err
	}

	c.Color, err = colorful.Hex("#" + in)
	return err
}
