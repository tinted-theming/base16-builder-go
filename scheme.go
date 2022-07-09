package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

var base16Bases = []string{
	"base00", "base01", "base02", "base03", "base04", "base05", "base06", "base07",
	"base08", "base09", "base0A", "base0B", "base0C", "base0D", "base0E", "base0F",
}

var base24Bases = []string{
	"base10", "base11", "base12", "base13", "base14", "base15", "base16", "base17",
}

// legacyScheme contains only the fields which are different between a base16
// scheme and our universal scheme format.
type legacyScheme struct {
	Scheme string `yaml:"scheme"`

	// Colors will hold all the "base*" variables.
	Colors map[string]color `yaml:",inline"`
}

type universalScheme struct {
	Slug string `yaml:"-"`

	Name        string           `yaml:"name"`
	Author      string           `yaml:"author"`
	System      string           `yaml:"system"`
	Description string           `yaml:"description"`
	Palette     map[string]color `yaml:"palette"`
}

func schemeFromFile(schemesFS fs.FS, fileName string) (*universalScheme, bool) {
	ret := &universalScheme{}

	logger := log.WithField("file", fileName)

	if !strings.HasSuffix(fileName, ".yaml") {
		logger.Error("Scheme must end in .yaml")
		return nil, false
	}

	data, err := fs.ReadFile(schemesFS, fileName)
	if err != nil {
		logger.Error(err)
		return nil, false
	}

	err = yaml.Unmarshal(data, ret)
	if err != nil {
		logger.Error(err)
		return nil, false
	}

	// If there's no scheme system defined, we assume this is either a base16 or
	// base24 style scheme, so we need to parse it again as the legacy format
	// and convert it.
	if ret.System == "" {
		var legacy legacyScheme
		err = yaml.Unmarshal(data, &legacy)
		if err != nil {
			logger.Error(err)
			return nil, false
		}

		// The name was previously called "scheme".
		ret.Name = legacy.Scheme

		var missingColors []string

		for _, baseKey := range base16Bases {
			if val, ok := legacy.Colors[baseKey]; ok {
				ret.Palette[baseKey] = val
			} else {
				missingColors = append(missingColors, baseKey)
			}
		}

		// At this point we've checked all of the original 16 bases (which are
		// included in both base16 and base24), so if we don't have all 16
		// colors, it's an error.
		if len(ret.Palette) != 16 {
			logger.Errorf("Missing colors from base16 pallete: %s", strings.Join(missingColors, ", "))
			return nil, false
		}

		for _, baseKey := range base24Bases {
			if val, ok := legacy.Colors[baseKey]; ok {
				ret.Palette[baseKey] = val
			} else {
				missingColors = append(missingColors, baseKey)
			}
		}

		// Infer the palette based on how many colors we ended up with.
		if len(ret.Palette) == 16 {
			ret.System = "base16"
		} else if len(ret.Palette) == 24 {
			ret.System = "base24"
		} else {
			logger.Errorf("Missing colors from base24 pallete: %s", strings.Join(missingColors, ", "))
			return nil, false
		}
	}

	// Now that we have the data, we can sanitize it
	ok := true
	if ret.Name == "" {
		logger.Error("Scheme name cannot be empty")
		ok = false
	}

	// Author is a warning because there appear to be some themes
	// without them.
	if ret.Author == "" {
		logger.Warn("Scheme author should not be empty")
	}

	// Sanitize any fields which were added later
	if ret.Description == "" {
		ret.Description = ret.Name
	}

	// Take the last path component and chop off .yaml
	ret.Slug = filepath.Base(strings.TrimSuffix(fileName, ".yaml"))

	return ret, ok
}

func (s *universalScheme) mustacheContext() map[string]interface{} {
	ret := map[string]interface{}{
		"scheme-name":             s.Name,
		"scheme-author":           s.Author,
		"scheme-slug":             s.Slug,
		"scheme-system":           s.System,
		"scheme-description":      s.Description,
		"scheme-slug-underscored": strings.Replace(s.Slug, "-", "_", -1),
	}

	for colorKey, colorVal := range s.Palette {
		// Note that we only lowercase the output of this to match the reference
		// repo.
		ret[colorKey+"-hex"] = fmt.Sprintf("%02x%02x%02x", colorVal.R, colorVal.G, colorVal.B)
		ret[colorKey+"-hex-bgr"] = fmt.Sprintf("%02x%02x%02x", colorVal.B, colorVal.G, colorVal.R)

		ret[colorKey+"-rgb-r"] = colorVal.R
		ret[colorKey+"-rgb-g"] = colorVal.G
		ret[colorKey+"-rgb-b"] = colorVal.B
		ret[colorKey+"-dec-r"] = float32(colorVal.R) / 255
		ret[colorKey+"-dec-g"] = float32(colorVal.G) / 255
		ret[colorKey+"-dec-b"] = float32(colorVal.B) / 255
		ret[colorKey+"-hex-r"] = fmt.Sprintf("%02x", colorVal.R)
		ret[colorKey+"-hex-g"] = fmt.Sprintf("%02x", colorVal.G)
		ret[colorKey+"-hex-b"] = fmt.Sprintf("%02x", colorVal.B)
	}

	return ret
}

func loadSchemes(schemesFS fs.FS) ([]*universalScheme, bool) {
	schemes := make(map[string]map[string]*universalScheme)

	// Pre-create some of our special cases to make it easier later
	schemes["base16"] = make(map[string]*universalScheme)
	schemes["base17"] = make(map[string]*universalScheme)

	schemePaths, err := fs.Glob(schemesFS, "*.yaml")
	if err != nil {
		log.Error(err)
		return nil, false
	}

	additionalSchemePaths, err := fs.Glob(schemesFS, "*/*.yaml")
	if err != nil {
		log.Error(err)
		return nil, false
	}

	schemePaths = append(schemePaths, additionalSchemePaths...)

	for _, schemePath := range schemePaths {
		scheme, ok := schemeFromFile(schemesFS, schemePath)
		if !ok {
			log.Errorf("Failed to load scheme")
			return nil, false
		}

		if _, ok := schemes[scheme.System]; !ok {
			schemes[scheme.System] = make(map[string]*universalScheme)
		}

		if _, ok := schemes[scheme.System][scheme.Slug]; ok {
			log.WithField("scheme", scheme.Slug).Warnf("Conflicting scheme")
		}

		log.Debugf("Found scheme %q", scheme.Slug)

		schemes[scheme.System][scheme.Slug] = scheme
	}

	// Copy all base17 schemes to base16 which are missing
	for _, scheme := range schemes["base17"] {
		if _, ok := schemes["base16"][scheme.Slug]; ok {
			continue
		}

		// Copy the scheme and update the "system"
		var newScheme universalScheme = *scheme
		newScheme.System = "base16"
		schemes["base16"][scheme.Slug] = &newScheme
	}

	// Copy all base16 schemes to base17 which are missing
	for _, scheme := range schemes["base16"] {
		if _, ok := schemes["base17"][scheme.Slug]; ok {
			continue
		}

		// Copy the scheme and update the "system"
		var newScheme universalScheme = *scheme
		newScheme.System = "base17"
		schemes["base17"][scheme.Slug] = &newScheme
	}

	var ret []*universalScheme
	for _, system := range schemes {
		for _, scheme := range system {
			ret = append(ret, scheme)
		}
	}

	return ret, true
}
