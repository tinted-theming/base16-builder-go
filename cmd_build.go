package main

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(buildCmd)
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build templates",
	Run: func(cmd *cobra.Command, args []string) {
		colorSchemes, ok := loadSchemes(filepath.Join(sourcesDir, "schemes", "list.yaml"))
		if !ok {
			log.Fatal("Failed to load color schemes")
		}

		log.Infof("Found %d color schemes", len(colorSchemes))

		templates, ok := loadTemplates(filepath.Join(sourcesDir, "templates", "list.yaml"), args)
		if !ok {
			log.Fatal("Failed to load templates")
		}

		log.Infof("Found %d templates", len(templates))

		for _, template := range templates {
			log.Infof("Rendering template %s in %s", template.Name, template.Dir)
			err := template.Render(colorSchemes)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}
