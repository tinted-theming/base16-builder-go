package main

import (
	"os"
	"path"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ignoreErrors    bool
	templatesSource string
	schemesSource   string
)

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&ignoreErrors, "ignore-errors", false, "Don't exit on error if possible to continue")
	updateCmd.Flags().StringVar(&templatesSource, "templates-source", "https://github.com/chriskempson/base16-templates-source.git", "Repo to grab templates from")
	updateCmd.Flags().StringVar(&schemesSource, "schemes-source", "https://github.com/chriskempson/base16-schemes-source.git", "Repo to grab schemes from")

}

// buildCmd represents the build command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull in updates from the source repos",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Updating sources")
		if !cloneRepo(templatesSource, path.Join(sourcesDir, "templates"), "templates") {
			log.Fatal("Failed to update templates sources")
		}
		if !cloneRepo(schemesSource, path.Join(sourcesDir, "schemes"), "schemes") {
			log.Fatal("Failed to update scheme sources")
		}

		var errored bool

		log.Info("Updating schemes")
		if !downloadSourceList(path.Join(sourcesDir, "schemes", "list.yaml"), schemesDir) {
			if !ignoreErrors {
				log.Fatal("Failed to update schemes")
			}

			errored = true
		}

		log.Info("Updating templates")
		if !downloadSourceList(path.Join(sourcesDir, "templates", "list.yaml"), templatesDir) {
			if !ignoreErrors {
				log.Fatal("Failed to update templates")
			}

			errored = true
		}

		if errored {
			log.Fatal("An error occured while updating")
		}
	},
}

func downloadSourceList(sourceFile, targetDir string) bool {
	sources, err := readSourcesList(sourceFile)
	if err != nil {
		log.Error(err)
		return false
	}

	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		log.Error(err)
		return false
	}

	ok := true
	for _, source := range sources {
		key := source.Key.(string)

		sourceDir := path.Join(targetDir, key)
		sourceLocation := source.Value.(string)

		ok = cloneRepo(sourceLocation, sourceDir, key) && ok
	}

	return ok
}

func cloneRepo(src, dest, key string) bool {
	logger := log.WithField("source", key)

	repo, err := vcs.NewRepo(src, dest)
	if err != nil {
		logger.Error(err)
		return false
	}

	if ok := repo.CheckLocal(); !ok {
		logger.Debugf("Cloning %q to %q", src, dest)
		err = repo.Get()
		if err != nil {
			handleVcsError(logger, err)
			return false
		}
	} else {
		logger.Debugf("Updating %q", dest)
		err = repo.Update()
		if err != nil {
			handleVcsError(logger, err)
			return false
		}
	}

	return true
}

func handleVcsError(logger *logrus.Entry, err error) {
	logger.Error(err)

	if lErr, ok := err.(*vcs.LocalError); ok {
		logger.Error(strings.TrimSpace(lErr.Out()))
		logger.Error(lErr.Original())
	}

	if rErr, ok := err.(*vcs.RemoteError); ok {
		logger.Error(strings.TrimSpace(rErr.Out()))
		logger.Error(rErr.Original())
	}
}
