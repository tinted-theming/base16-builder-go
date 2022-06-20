package main

import (
	"embed"
	"flag"
	"io/fs"
	"os"

	"github.com/sirupsen/logrus"
)

//go:embed schemes/*.yaml
var schemesFS embed.FS

var (
	schemesDir  string
	templateDir string

	// Define the logger we'll be using
	logVerbose bool
	rawLog     = logrus.New()
	log        = logrus.NewEntry(rawLog)

	// Variables set by goreleaser
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func init() {
	flag.StringVar(&schemesDir, "schemes-dir", "-", "Target directory for scheme data. The default value uses internal schemes.")
	flag.StringVar(&templateDir, "template-dir", ".", "Target template directory to build.")
	flag.BoolVar(&logVerbose, "verbose", false, "Log all debug messages")

	rawLog.Level = logrus.InfoLevel
	if logVerbose {
		rawLog.Level = logrus.DebugLevel
	}
}

func main() {
	flag.Parse()

	log.WithFields(logrus.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("base16-builder-go")

	var targetFS fs.FS
	if schemesDir == "-" {
		log.Info("Using internal schemes")
		targetFS, _ = fs.Sub(schemesFS, "schemes")
	} else {
		log.Infof("Processing scheme dir %q", schemesDir)
		targetFS = os.DirFS(schemesDir)
	}

	colorSchemes, ok := loadSchemes(targetFS)
	if !ok {
		log.Fatal("Failed to load color schemes")
	}

	log.Infof("Found %d color schemes", len(colorSchemes))

	templates, err := templatesFromFile(templateDir)
	if err != nil {
		log.Panic(err)
	}

	for _, template := range templates {
		err = template.Render(colorSchemes)
		if err != nil {
			log.Panic(err)
		}
	}
}
