package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	slugify "github.com/metal3d/go-slugify"
)

var bases = []string{
	"00", "01", "02", "03", "04", "05", "06", "07",
	"08", "09", "0A", "0B", "0C", "0D", "0E", "0F",
}

type colorScheme struct {
	Slug string `yaml:"-"`

	Scheme string `yaml:"scheme"`
	Author string `yaml:"author"`

	// Colors will hold all the "base*" variables.
	Colors map[string]color `yaml:",inline"`
}

func schemeFromFile(fileName string) (*colorScheme, error) {
	ret := &colorScheme{}

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

	return ret, nil
}

func (s colorScheme) exportToContext() (map[string]interface{}, error) {
	var errs []error
	if s.Scheme == "" {
		errs = append(errs, errors.New("Scheme name cannot be empty"))
	}

	if s.Author == "" {
		//errs = append(errs, errors.New("Scheme author cannot be empty"))
	}

	ret := map[string]interface{}{
		"scheme-name":   s.Scheme,
		"scheme-author": s.Author,
		"scheme-slug":   slugify.Marshal(s.Scheme),
	}

	if len(bases) != 16 {
		errs = append(errs, errors.New("Wrong number of colors"))
	}

	for _, base := range bases {
		baseKey := "base" + base
		baseVal, ok := s.Colors[baseKey]
		if !ok {
			errs = append(errs, fmt.Errorf("Missing %s in scheme", baseKey))
			continue
		}

		ret[baseKey+"-hex"] = baseVal.Hex()

		r, g, b := baseVal.RGB255()
		ret[baseKey+"-rgb-r"] = r
		ret[baseKey+"-rgb-g"] = g
		ret[baseKey+"-rgb-b"] = b
		ret[baseKey+"-hex-r"] = strconv.FormatUint(uint64(r), 16)
		ret[baseKey+"-hex-g"] = strconv.FormatUint(uint64(g), 16)
		ret[baseKey+"-hex-b"] = strconv.FormatUint(uint64(b), 16)

		// TODO: Missing hsl
	}

	if len(errs) != 0 {
		for _, innerErr := range errs {
			logrus.Println(innerErr)
		}

		return nil, errors.New("Failed to parse scheme")
	}

	return ret, nil
}
