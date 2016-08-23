package main

import (
	"os"
	"path"

	"github.com/Masterminds/vcs"
	"github.com/Unknwon/com"
	"github.com/spf13/cobra"
)

var sourcesFile string

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVar(&sourcesFile, "sources", "sources.yaml", "File with base16 sources")
}

// buildCmd represents the build command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull in updates from the source repos",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Updating sources")
		dirs, err := downloadSourceList(sourcesFile, sourcesDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, dir := range []string{"schemes", "templates"} {
			if !com.IsSliceContainsStr(dirs, dir) {
				log.Fatalf("%q location is missing from sources file", dir)
			}
		}

		log.Info("Updating schemes")
		_, err = downloadSourceList(path.Join(sourcesDir, "schemes", "list.yaml"), schemesDir)
		if err != nil {
			handleVcsError(err)
		}

		log.Info("Updating templates")
		_, err = downloadSourceList(path.Join(sourcesDir, "templates", "list.yaml"), templatesDir)
		if err != nil {
			handleVcsError(err)
		}
	},
}

func handleVcsError(err error) {
	if lErr, ok := err.(*vcs.LocalError); ok {
		log.Error(lErr)
		log.Fatal(lErr.Original())
	}

	if rErr, ok := err.(*vcs.RemoteError); ok {
		log.Error(rErr)
		log.Fatal(rErr.Original())
	}

	log.Fatal(err)
}

func downloadSourceList(sourceFile, targetDir string) ([]string, error) {
	sources, err := readSourcesList(sourceFile)
	if err != nil {
		return nil, err
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
			log.Debugf("Cloning %q to %q", sourceLocation, sourceDir)
			err = repo.Get()
			if err != nil {
				return nil, err
			}
		} else {
			log.Debugf("Updating %q", sourceDir)
			err = repo.Update()
			if err != nil {
				return nil, err
			}
		}
	}

	return ret, nil
}
