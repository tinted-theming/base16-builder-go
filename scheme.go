package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	slugify "github.com/metal3d/go-slugify"
)

var bases = []string{
	"00", "01", "02", "03", "04", "05", "06", "07",
	"08", "09", "0A", "0B", "0C", "0D", "0E", "0F",
}

type scheme struct {
	Name string `yaml:"-"`
	Slug string `yaml:"-"`

	Scheme string `yaml:"scheme"`
	Author string `yaml:"author"`

	// Colors will hold all the "base*" variables.
	Colors map[string]color `yaml:",inline"`
}

func schemeFromFile(fileName string) (*scheme, error) {
	ret := &scheme{}

	ret.Slug = path.Base(fileName)

	if !strings.HasSuffix(ret.Slug, ".yaml") {
		return nil, errors.New("Scheme name must end in .yaml")
	}

	ret.Slug = ret.Slug[:len(ret.Slug)-5]

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}

	// Now that we have the data, we can sanitize it
	var errored bool
	if ret.Scheme == "" {
		fmt.Println(errors.New("Scheme name cannot be empty"))
		errored = true
	}

	// Author is a warning because there appear to be some themes
	// without them.
	if ret.Author == "" {
		fmt.Println(errors.New("Scheme author cannot be empty"))
	}

	if len(bases) != len(ret.Colors) {
		fmt.Println(errors.New("Wrong number of colors in scheme"))
		errored = true
	}

	for _, base := range bases {
		baseKey := "base" + base
		if _, ok := ret.Colors[baseKey]; !ok {
			fmt.Println(fmt.Errorf("Scheme missing %q", baseKey))
			continue
		}
	}

	if errored {
		return nil, errors.New("Failed to parse scheme")
	}

	return ret, nil
}

func (s *scheme) mustacheContext() map[string]interface{} {
	ret := map[string]interface{}{
		"scheme-name":   s.Scheme,
		"scheme-author": s.Author,
		"scheme-slug":   slugify.Marshal(s.Scheme),
	}

	for _, base := range bases {
		baseKey := "base" + base
		baseVal := s.Colors[baseKey]

		ret[baseKey+"-hex"] = baseVal.Hex()

		r, g, b := baseVal.RGB255()
		ret[baseKey+"-rgb-r"] = r
		ret[baseKey+"-rgb-g"] = g
		ret[baseKey+"-rgb-b"] = b
		ret[baseKey+"-hex-r"] = strconv.FormatUint(uint64(r), 16)
		ret[baseKey+"-hex-g"] = strconv.FormatUint(uint64(g), 16)
		ret[baseKey+"-hex-b"] = strconv.FormatUint(uint64(b), 16)

		h, c, l := baseVal.Hcl()
		ret[baseKey+"-hcl-h"] = h
		ret[baseKey+"-hcl-c"] = c
		ret[baseKey+"-hcl-l"] = l
	}

	return ret
}
