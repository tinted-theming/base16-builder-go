package main

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	slugify "github.com/metal3d/go-slugify"
	yaml "gopkg.in/yaml.v2"
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

func schemeFromFile(fileName string) (*scheme, bool) {
	ret := &scheme{}

	logger := log.WithField("file", fileName)

	if !strings.HasSuffix(fileName, ".yaml") {
		logger.Error("Scheme must end in .yaml")
		return nil, false
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.Error(err)
		return nil, false
	}

	err = yaml.Unmarshal(data, ret)
	if err != nil {
		logger.Error(err)
		return nil, false
	}

	// Now that we have the data, we can sanitize it
	ok := true
	if ret.Scheme == "" {
		logger.Error("Scheme name cannot be empty")
		ok = false
	}

	// Author is a warning because there appear to be some themes
	// without them.
	if ret.Author == "" {
		logger.Warn("Scheme author should not be empty")
	}

	if len(bases) != len(ret.Colors) {
		logger.Error("Wrong number of colors in scheme")
		ok = false
	}

	// Now that we've got all that out of the way, we can start
	// processing stuff.
	ret.Slug = strings.ToLower(slugify.Marshal(ret.Scheme))

	for _, base := range bases {
		baseKey := "base" + base
		if _, ok := ret.Colors[baseKey]; !ok {
			logger.Errorf("Scheme missing %q", baseKey)
			ok = false
			continue
		}
	}

	return ret, ok
}

func (s *scheme) mustacheContext() map[string]interface{} {
	ret := map[string]interface{}{
		"scheme-name":   s.Scheme,
		"scheme-author": s.Author,
		"scheme-slug":   s.Slug,
	}

	for _, base := range bases {
		baseKey := "base" + base
		baseVal := s.Colors[baseKey]

		// Note that we only lowercase the output of this to match the
		// reference repo.
		ret[baseKey+"-hex"] = strings.ToLower(strings.TrimLeft(baseVal.Hex(), "#"))

		r, g, b := baseVal.RGB255()
		ret[baseKey+"-rgb-r"] = r
		ret[baseKey+"-rgb-g"] = g
		ret[baseKey+"-rgb-b"] = b
		ret[baseKey+"-hex-r"] = strconv.FormatUint(uint64(r), 16)
		ret[baseKey+"-hex-g"] = strconv.FormatUint(uint64(g), 16)
		ret[baseKey+"-hex-b"] = strconv.FormatUint(uint64(b), 16)
	}

	return ret
}

func loadSchemes(schemeFile string) ([]*scheme, bool) {
	schemeItems, err := readSourcesList(schemeFile)
	if err != nil {
		log.Error(err)
		return nil, false
	}

	ok := true
	schemes := make(map[string]*scheme)
	for _, item := range schemeItems {
		schemeName := item.Key.(string)
		log.Infof("Processing scheme dir %q", schemeName)

		schemeGroupPath := path.Join(schemesDir, schemeName, "*.yaml")

		schemePaths, err := filepath.Glob(schemeGroupPath)
		if err != nil {
			log.Error(err)
			ok = false
			continue
		}

		for _, schemePath := range schemePaths {
			scheme, ok := schemeFromFile(schemePath)
			if !ok {
				log.Errorf("Failed to load scheme")
				ok = false
				continue
			}

			if _, ok := schemes[scheme.Slug]; ok {
				log.WithField("scheme", scheme.Slug).Warnf("Conflicting scheme")
			}

			log.Debugf("Found scheme %q", scheme.Slug)

			schemes[scheme.Slug] = scheme
		}
	}

	ret := []*scheme{}
	for _, scheme := range schemes {
		ret = append(ret, scheme)
	}

	return ret, ok
}
