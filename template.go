package main

import (
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

func loadTemplates(templateFile string) ([]*template, bool) {
	templateItems, err := readSourcesList(templateFile)
	if err != nil {
		log.Error(err)
		return nil, false
	}

	ok := true
	ret := []*template{}
	for _, item := range templateItems {
		templateName := item.Key.(string)
		log.Infof("Processing templates dir %q", templateName)

		templateDir := path.Join(templatesDir, templateName)
		templates, err := templatesFromFile(templateDir)
		if err != nil {
			log.Error(err)
			ok = false
			continue
		}

		ret = append(ret, templates...)
	}

	return ret, ok
}
