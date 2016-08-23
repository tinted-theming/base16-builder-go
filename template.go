package main

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/hoisie/mustache"

	yaml "gopkg.in/yaml.v2"
)

type template struct {
	Name      string `yaml:"-"`
	Dir       string `yaml:"-"`
	Extension string `yaml:"extension"`
	OutputDir string `yaml:"output"`
}

func templatesFromFile(templatesDir string) ([]*template, error) {
	data, err := ioutil.ReadFile(path.Join(templatesDir, "templates", "config.yaml"))
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
			fmt.Println("OutputDir missing from theme config block")
		}

		if t.Extension == "" {
			fmt.Println("Extension missing from theme config block")
		}

		ret = append(ret, t)
	}

	return ret, nil
}

func (t *template) Render(schemes []*scheme) error {
	m, err := mustache.ParseFile(path.Join(t.Dir, "templates", t.Name+".mustache"))
	if err != nil {
		return err
	}

	outputDir := path.Join(t.Dir, t.OutputDir)
	for _, scheme := range schemes {
		fileName := path.Join(outputDir, scheme.Name+t.Extension)
		rendered := m.Render(scheme.mustacheContext())
		err = ioutil.WriteFile(fileName, []byte(rendered), 0777)
		if err != nil {
			return err
		}
	}

	return nil
}
