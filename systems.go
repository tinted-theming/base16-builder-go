package main

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/hashicorp/go-multierror"
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
		return LoadLegacyScheme(data)
	}

	return LoadUniversalScheme(data)
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
	Variant     string           `yaml:"variant"`
	Slug        string           `yaml:"slug"`
	Palette     map[string]color `yaml:",inline"`
}

func LoadLegacyScheme(data []byte) (*ColorScheme, error) {
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
		Slug:        scheme.Slug,
		Description: scheme.Description,
		Variant:     scheme.Variant,
		Palette:     scheme.Palette,
	}

	if ret.Slug == "" {
		ret.Slug, err = Slugify(ret.Name)
		if err != nil {
			return nil, fmt.Errorf("Failed to slugify name: %e", err)
		}
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
	Slug        string            `yaml:"slug"`
	Name        string            `yaml:"name"`
	Author      string            `yaml:"author"`
	System      string            `yaml:"system"`
	Description string            `yaml:"description"`
	Variant     string            `yaml:"variant"`
	Palette     map[string]color  `yaml:"palette"`
	Mappings    map[string]string `yaml:"mappings"`
}

func LoadUniversalScheme(data []byte) (*ColorScheme, error) {
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
		Variant:     scheme.Variant,
		Palette:     scheme.Palette,
	}

	// Missing the author field is a warning, not an error because there appear
	// to be some pre-existing themes without them.
	if ret.Author == "" {
		log.Warn("Scheme author should not be empty")
	}

	// If we have an empty slug, we need to infer it from the scheme name. This
	// involves normalizing any unicode, lower-casing it, replacing spaces with
	// dashes.
	if ret.Slug == "" {
		ret.Slug, err = Slugify(ret.Name)
		if err != nil {
			return nil, err
		}
	}

	if ret.Description == "" {
		ret.Description = ret.Name
	}

	if ret.Name == "" {
		return nil, errors.New("scheme name cannot be empty")
	}

	merr := &multierror.Error{}

	// Copy any mappings into the palette
	for key, alias := range scheme.Mappings {
		if _, ok := ret.Palette[key]; ok {
			merr = AppendErrorf(merr, "duplicate key in palette and mappings: %s", key)
			continue
		}

		if _, ok := ret.Palette[alias]; !ok {
			merr = AppendErrorf(merr, "missing referenced alias: %s", alias)
			continue
		}

		ret.Palette[key] = ret.Palette[alias]
	}

	return ret, merr.ErrorOrNil()
}
