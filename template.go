package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cbroglie/mustache"
	yaml "gopkg.in/yaml.v3"
)

type template struct {
	Name      string `yaml:"-"`
	Dir       string `yaml:"-"`
	Extension string `yaml:"extension"`
	OutputDir string `yaml:"output"`
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

		if t.OutputDir == "" {
			log.Warn("OutputDir missing from theme config block")
		}

		if t.Extension == "" {
			log.Warn("Extension missing from theme config block")
		}

		log.Debugf("Found template %q in dir %q", t.Name, t.Dir)

		ret = append(ret, t)
	}

	return ret, nil
}

func (t *template) Render(schemes []*scheme) error {
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

	for _, scheme := range schemes {
		fileName := filepath.Join(outputDir, "base16-"+scheme.Slug+t.Extension)
		rendered, err := m.Render(scheme.mustacheContext())
		if err != nil {
			return err
		}

		// We use 666 as the filemode here rather than 777 because we don't want
		// it executable by default.
		err = os.WriteFile(fileName, []byte(rendered), 0666)
		if err != nil {
			return err
		}
	}

	return nil
}
