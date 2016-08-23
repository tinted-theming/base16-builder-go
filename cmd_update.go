package main

import (
	"fmt"
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
		fmt.Println("Updating sources")
		dirs, err := downloadSourceList(sourcesFile, sourcesDir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, dir := range []string{"schemes", "templates"} {
			if !com.IsSliceContainsStr(dirs, dir) {
				fmt.Printf("%q location is missing from sources file", dir)
				os.Exit(1)
			}
		}

		fmt.Println("Updating schemes")
		_, err = downloadSourceList(path.Join(sourcesDir, "schemes", "list.yaml"), schemesDir)
		if err != nil {
			handleVcsError(err)
		}

		fmt.Println("Updating templates")
		_, err = downloadSourceList(path.Join(sourcesDir, "templates", "list.yaml"), templatesDir)
		if err != nil {
			handleVcsError(err)
		}
	},
}

func handleVcsError(err error) {
	if lErr, ok := err.(*vcs.LocalError); ok {
		fmt.Println(lErr.Original())
	} else if rErr, ok := err.(*vcs.RemoteError); ok {
		fmt.Println(rErr.Original())
	} else {
		fmt.Println(err)
	}

	os.Exit(1)
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
			fmt.Printf("Cloning %q from %q\n", sourceLocation, sourceDir)
			err = repo.Get()
			if err != nil {
				return nil, err
			}
		} else {
			fmt.Printf("Updating %q\n", sourceDir)
			err = repo.Update()
			if err != nil {
				return nil, err
			}
		}
	}

	return ret, nil
}
