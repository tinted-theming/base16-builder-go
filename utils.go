package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

var colorRegex = regexp.MustCompile(`^[0-9a-fA-F]{6}$`)

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

func readSourcesList(fileName string) (yaml.MapSlice, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var sources yaml.MapSlice
	err = yaml.Unmarshal(data, &sources)
	if err != nil {
		return nil, err
	}

	err = validateMapSlice(sources)
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func validateMapSlice(sources yaml.MapSlice) error {
	// Run through all the values and sanitize them
	dupeSet := make(map[string]struct{})
	for _, item := range sources {
		key, ok := item.Key.(string)
		if !ok {
			return fmt.Errorf("Failed to decode key %q as string", item.Key)
		}

		if _, ok := item.Value.(string); !ok {
			return fmt.Errorf("Failed to decode value %q for key %q as string", item.Value, item.Key)
		}

		if _, ok := dupeSet[key]; ok {
			return fmt.Errorf("Duplicate key %q", key)
		}

		dupeSet[key] = struct{}{}
	}

	return nil
}
