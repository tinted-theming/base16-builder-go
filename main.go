package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Masterminds/vcs"
	"github.com/Sirupsen/logrus"
	"github.com/hoisie/mustache"
	yaml "gopkg.in/yaml.v2"
)

var (
	sourcesDir   = "./sources"
	schemesDir   = "./schemes"
	templatesDir = "./templates"
)

type confBlock struct {
	Extension string `yaml:"extension"`
	OutputDir string `yaml:"output"`
}

func handleVcsError(err error) {
	if lErr, ok := err.(*vcs.LocalError); ok {
		logrus.Fatalln(lErr.Original())
	}

	if rErr, ok := err.(*vcs.RemoteError); ok {
		logrus.Fatalln(rErr.Original())
	}

	logrus.Fatalln(err)
}

func main() {
	//logrus.SetLevel(logrus.DebugLevel)

	logrus.Println("Updating sources")
	_, err := downloadSourceList("sources.yaml", sourcesDir)
	if err != nil {
		handleVcsError(err)
	}

	logrus.Println("Updating schemes")
	schemeNames, err := downloadSourceList(path.Join(sourcesDir, "schemes", "list.yaml"), schemesDir)
	if err != nil {
		handleVcsError(err)
	}

	logrus.Println("Updating templates")
	templateNames, err := downloadSourceList(path.Join(sourcesDir, "templates", "list.yaml"), templatesDir)
	if err != nil {
		handleVcsError(err)
	}

	schemes := make(map[string]*colorScheme)

	for _, schemeName := range schemeNames {
		logrus.Printf("Processing scheme dir %q", schemeName)

		schemeGroupPath := path.Join(schemesDir, schemeName, "*.yaml")

		schemePaths, err := filepath.Glob(schemeGroupPath)
		if err != nil {
			logrus.Fatalln(err)
		}

		for _, schemePath := range schemePaths {
			scheme, err := schemeFromFile(schemePath)
			if err != nil {
				logrus.Fatalln(err)
			}

			if _, ok := schemes[scheme.Slug]; ok {
				logrus.Warnf("Conflicting scheme %q", scheme.Slug)
			}

			logrus.Debugf("Found scheme %q", scheme.Slug)

			schemes[scheme.Slug] = scheme
		}
	}

	logrus.Infof("Loaded %d schemes", len(schemes))

	for _, templateName := range templateNames {
		dirPath := path.Join(templatesDir, templateName)

		languageTemplatesDir := path.Join(dirPath, "templates")

		logrus.Printf("Processing template dir %q", dirPath)

		data, err := ioutil.ReadFile(path.Join(languageTemplatesDir, "config.yaml"))
		if err != nil {
			logrus.Fatalln(err)
		}

		out := make(map[string]confBlock)
		err = yaml.Unmarshal(data, out)
		if err != nil {
			logrus.Fatalln(err)
		}

		for name, block := range out {
			if block.OutputDir == "" {
				logrus.Warnf("OutputDir missing from theme config block")
			}

			if block.Extension == "" {
				logrus.Warnf("Extension missing from theme config block")
			}

			template, err := mustache.ParseFile(path.Join(languageTemplatesDir, name+".mustache"))
			if err != nil {
				logrus.Fatalln(err)
			}

			outputDir := path.Join(dirPath, block.OutputDir)

			for schemeName, schemeData := range schemes {
				fileName := path.Join(outputDir, schemeName+block.Extension)

				logrus.Debugf("Rendering %q", fileName)

				context, err := schemeData.exportToContext()
				if err != nil {
					logrus.Fatalln(err)
				}

				rendered := template.Render(context)
				err = ioutil.WriteFile(fileName, []byte(rendered), 0777)
				if err != nil {
					logrus.Fatalln(err)
				}
			}
		}
	}

	logrus.Printf("Rendered %d files", len(templateNames)*len(schemes))
}

func downloadSourceList(sourceFile, targetDir string) ([]string, error) {
	// TODO: ENSURE NO DUPES
	data, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return nil, err
	}

	// Decode into a MapSlice so we can maintain order.
	var sources yaml.MapSlice
	err = yaml.Unmarshal(data, &sources)
	if err != nil {
		return nil, err
	}

	// Run through the MapSlice and ensure everything is a string.
	for _, item := range sources {
		if _, ok := item.Key.(string); !ok {
			return nil, fmt.Errorf("Failed to decode key %q as string", item.Key)
		}

		if _, ok := item.Value.(string); !ok {
			return nil, fmt.Errorf("Failed to decode value %q for key %q as string", item.Value, item.Key)
		}
	}

	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, source := range sources {
		ret = append(ret, source.Key.(string))

		sourceDir := path.Join(targetDir, source.Key.(string))
		sourceLocation := source.Value.(string)

		repo, err := vcs.NewRepo(sourceLocation, sourceDir)
		if err != nil {
			return nil, err
		}

		if ok := repo.CheckLocal(); !ok {
			logrus.Debugf("Cloning %q from %q", sourceLocation, sourceDir)
			err = repo.Get()
			if err != nil {
				return nil, err
			}
		} else {
			logrus.Debugf("Updating %q", sourceDir)
			err = repo.Update()
			if err != nil {
				return nil, err
			}
		}
	}

	return ret, nil
}
