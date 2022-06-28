package main

import (
	"compress/gzip"
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"

	"github.com/nlepage/go-tarfs"
	"github.com/sirupsen/logrus"
)

//go:embed schemes/*.yaml
var schemesFS embed.FS

var (
	schemesDir  string
	templateDir string
	onlineMode  bool

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
	flag.BoolVar(&onlineMode, "online", false, "Run in online mode and pull schemes directly from GitHub")

	rawLog.Level = logrus.InfoLevel
	if logVerbose {
		rawLog.Level = logrus.DebugLevel
	}
}

func getSchemesFromGithub() (fs.FS, error) {
	log.Info("Attempting to load schemes from GitHub")

	r, err := http.Get("https://github.com/base16-project/base16-schemes/archive/refs/heads/main.tar.gz")
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}

	targetFS, err := tarfs.New(gzipReader)
	if err != nil {
		return nil, err
	}

	// The archive has a subfolder containing all the schemes, so we return a
	// subfs of the folder.
	return fs.Sub(targetFS, "base16-schemes-main")
}

func main() {
	var err error

	flag.Parse()

	log.WithFields(logrus.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("base16-builder-go")

	var targetFS fs.FS
	if schemesDir == "-" {
		// If we're in online mode, the default is to pull from GitHub,
		// otherwise use the embedded schemes.
		if onlineMode {
			targetFS, err = getSchemesFromGithub()
			if err != nil {
				log.WithError(err).Fatal("Failed to load schemes from GitHub")
			}
		} else {
			log.Info("Using internal schemes")
			targetFS, _ = fs.Sub(schemesFS, "schemes")
		}
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
