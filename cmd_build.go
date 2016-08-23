package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(buildCmd)
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all templates",
	Run: func(cmd *cobra.Command, args []string) {
		colorSchemes, err := loadSchemes(path.Join(sourcesDir, "schemes", "list.yaml"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		templates, err := loadTemplates(path.Join(sourcesDir, "templates", "list.yaml"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, template := range templates {
			err := template.Render(colorSchemes)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}

func loadSchemes(schemeFile string) ([]*scheme, error) {
	schemeItems, err := readSourcesList(schemeFile)
	if err != nil {
		return nil, err
	}

	schemes := make(map[string]*scheme)
	for _, item := range schemeItems {
		schemeName := item.Key.(string)
		fmt.Printf("Processing scheme dir %q\n", schemeName)

		schemeGroupPath := path.Join(schemesDir, schemeName, "*.yaml")

		schemePaths, err := filepath.Glob(schemeGroupPath)
		if err != nil {
			return nil, err
		}

		for _, schemePath := range schemePaths {
			scheme, err := schemeFromFile(schemePath)
			if err != nil {
				return nil, err
			}

			if _, ok := schemes[scheme.Slug]; ok {
				fmt.Printf("Conflicting scheme %q\n", scheme.Slug)
			}

			fmt.Printf("Found scheme %q\n", scheme.Slug)

			schemes[scheme.Slug] = scheme
		}
	}

	ret := []*scheme{}
	for _, scheme := range schemes {
		ret = append(ret, scheme)
	}

	return ret, nil
}

func loadTemplates(templateFile string) ([]*template, error) {
	templateItems, err := readSourcesList(templateFile)
	if err != nil {
		return nil, err
	}

	ret := []*template{}
	for _, item := range templateItems {
		templateName := item.Key.(string)
		fmt.Printf("Processing templates dir %q\n", templateName)

		templateDir := path.Join(templatesDir, templateName)
		templates, err := templatesFromFile(templateDir)
		if err != nil {
			return nil, err
		}

		ret = append(ret, templates...)
	}

	return ret, nil
}
