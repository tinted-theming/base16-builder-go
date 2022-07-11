package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type baseScheme struct {
	System string `yaml:"system"`
}

func LoadScheme(schemesFS fs.FS, filename string) (*ColorScheme, error) {
	var baseScheme baseScheme

	data, err := fs.ReadFile(schemesFS, filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &baseScheme)
	if err != nil {
		return nil, err
	}

	// If no system is specified, it can be loaded as a legacy scheme
	// (base16/base24).
	if baseScheme.System == "" {
		return LoadLegacyScheme(filename, data)
	}

	return LoadUniversalScheme(filename, data)
}

var base16Bases = []string{
	"base00", "base01", "base02", "base03", "base04", "base05", "base06", "base07",
	"base08", "base09", "base0A", "base0B", "base0C", "base0D", "base0E", "base0F",
}

var base24Bases = []string{
	"base10", "base11", "base12", "base13", "base14", "base15", "base16", "base17",
}

type legacyScheme struct {
	Name        string           `yaml:"scheme"`
	Author      string           `yaml:"author"`
	Description string           `yaml:"description"`
	Palette     map[string]color `yaml:",inline"`
}

func LoadLegacyScheme(filename string, data []byte) (*ColorScheme, error) {
	var scheme legacyScheme
	err := yaml.Unmarshal(data, &scheme)
	if err != nil {
		return nil, err
	}

	var missingColors []string

	for _, baseName := range base16Bases {
		if _, ok := scheme.Palette[baseName]; !ok {
			missingColors = append(missingColors, baseName)
		}
	}

	if len(missingColors) != 0 {
		return nil, fmt.Errorf("Missing colors from base16 palette: %s", strings.Join(missingColors, ", "))
	}

	for _, baseName := range base24Bases {
		if _, ok := scheme.Palette[baseName]; !ok {
			missingColors = append(missingColors, baseName)
		}
	}

	// If there were more than 16 colors and there were missing colors, we know
	// they were from the base24 palette.
	if len(scheme.Palette) > 16 && len(missingColors) != 0 {
		return nil, fmt.Errorf("Missing colors from base24 palette: %s", strings.Join(missingColors, ", "))
	}

	// Now that we've validated the data, we can convert it to the internal
	// ColorScheme format.
	ret := &ColorScheme{
		Name:        scheme.Name,
		Author:      scheme.Author,
		Slug:        filepath.Base(strings.TrimSuffix(filename, ".yaml")),
		Description: scheme.Description,
		Palette:     scheme.Palette,
	}

	if len(ret.Palette) == 16 {
		ret.System = "base16"
	} else if len(ret.Palette) == 24 {
		ret.System = "base24"
	} else {
		return nil, fmt.Errorf("Unexpected number of palette colors: %d", len(ret.Palette))
	}

	// Description isn't technically in base16, so we fall back to the name.
	if ret.Description == "" {
		ret.Description = ret.Name
	}

	return ret, nil
}

type universalScheme struct {
	Slug        string           `yaml:"slug"`
	Name        string           `yaml:"name"`
	Author      string           `yaml:"author"`
	System      string           `yaml:"system"`
	Description string           `yaml:"description"`
	Palette     map[string]color `yaml:"palette"`
}

func LoadUniversalScheme(filename string, data []byte) (*ColorScheme, error) {
	logger := log.WithField("file", filename)

	var scheme universalScheme

	err := yaml.Unmarshal(data, &scheme)
	if err != nil {
		return nil, err
	}

	ret := &ColorScheme{
		Slug:        scheme.Slug,
		Name:        scheme.Name,
		Author:      scheme.Author,
		System:      scheme.System,
		Description: scheme.Description,
		Palette:     scheme.Palette,
	}

	// Author is a warning because there appear to be some themes
	// without them.
	if ret.Author == "" {
		logger.Warn("Scheme author should not be empty")
	}

	// If we have an empty slug, we need to infer it from the scheme name. This
	// involves normalizing any unicode, lower-casing it, replacing spaces with
	// dashes.
	if ret.Slug == "" {
		slug, err := ToAscii(ret.Name)
		if err != nil {
			return nil, err
		}

		// Replace spaces with dashes and drop everything else that isn't
		// alphanumeric.
		ret.Slug = strings.Map(func(c rune) rune {
			if c == ' ' || c == '-' {
				return '-'
			}

			if unicode.IsLetter(c) || unicode.IsNumber(c) {
				return c
			}

			return -1
		}, strings.ToLower(slug))
	}

	if ret.Description == "" {
		ret.Description = ret.Name
	}

	if ret.Name == "" {
		return nil, errors.New("Scheme name cannot be empty")
	}

	return ret, nil
}
