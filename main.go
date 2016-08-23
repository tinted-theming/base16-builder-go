package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	sourcesDir   string
	schemesDir   string
	templatesDir string
)

func init() {
	RootCmd.PersistentFlags().StringVar(&sourcesDir, "sources-dir", "./sources/", "Target directory for source repos")
	RootCmd.PersistentFlags().StringVar(&schemesDir, "schemes-dir", "./schemes/", "Target directory for scheme data")
	RootCmd.PersistentFlags().StringVar(&templatesDir, "templates-dir", "./templates/", "Target directory for template data")
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "base16-builder-go",
	Short: "A simple builder for base16 templates and schemes",
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
