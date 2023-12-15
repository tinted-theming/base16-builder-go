package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cbroglie/mustache"
	"golang.org/x/exp/slices"
	yaml "gopkg.in/yaml.v3"
)

type template struct {
	Name             string   `yaml:"-"`
	Dir              string   `yaml:"-"`
	Filename         string   `yaml:"filename"`
	Extension        string   `yaml:"extension"`
	OutputDir        string   `yaml:"output"`
	SupportedSystems []string `yaml:"supported-systems"`
}

func templatesFromFile(templatesDir string) ([]*template, error) {
	data, err := ioutil.ReadFile(filepath.Join(templateDir, "templates", "config.yaml"))
	if err != nil {
		return nil, err
	}

	out := make(map[string]*template)
	err = yaml.Unmarshal(data, out)
	if err != nil {
		return nil, err
	}

	ret := []*template{}
	for k, t := range out {
		t.Name = k
		t.Dir = templatesDir

		if t.Filename == "" {
			log.Info("Filename missing from theme config block, inferring from OutputDir and Extension")

			if t.OutputDir == "" {
				log.Warn("OutputDir missing from theme config block")
				t.OutputDir = "."
			}

			if t.Extension == "" {
				return nil, errors.New("Extension missing from theme config block")
			}

			t.Filename = fmt.Sprintf("%s/{{ scheme-system }}-{{ scheme-slug }}%s", t.OutputDir, t.Extension)
		}

		if len(t.SupportedSystems) == 0 {
			log.Warn("Systems not set in theme config block, inferring base16")
			t.SupportedSystems = []string{"base16"}
		}

		log.Debugf("Found template %q in dir %q", t.Name, t.Dir)

		ret = append(ret, t)
	}

	return ret, nil
}

func (t *template) Render(schemes []*ColorScheme) error {
	m, err := mustache.ParseFile(filepath.Join(t.Dir, "templates", t.Name+".mustache"))
	if err != nil {
		return err
	}

	outputDir := filepath.Join(t.Dir, t.OutputDir)

	stat, err := os.Stat(outputDir)
	if err != nil {
		log.Warnf("Directory %s does not exist. Creating.", outputDir)
		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("Output dir %s is not a dir", outputDir)
	}

	var templateRendered bool

	for _, scheme := range schemes {
		templateVariables := scheme.TemplateVariables()

		// If the scheme's system wasn't in this template's supported systems
		// list, we skip it.
		if !slices.Contains(t.SupportedSystems, scheme.System) {
			continue
		}

		filenameTemplate, err := mustache.ParseString(t.Filename)
		if err != nil {
			return err
		}

		fileName, err := filenameTemplate.Render(templateVariables)
		if err != nil {
			return err
		}

		rendered, err := m.Render(templateVariables)
		if err != nil {
			return err
		}

		// We use 666 as the filemode here rather than 777 because we don't want
		// it executable by default.
		err = os.WriteFile(fileName, []byte(rendered), 0666)
		if err != nil {
			return err
		}

		templateRendered = true
	}

	// We want to ensure at least 1 valid scheme exists for each template being
	// rendered.
	if !templateRendered {
		return errors.New("No valid schemes for template")
	}

	return nil
}
