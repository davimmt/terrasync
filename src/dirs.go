package src

import (
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
)

func FindTfDirs(gitRepoUrl string, rootWorkingDir string) []string {
	if gitRepoUrl != "" {
		_, err := git.PlainClone(rootWorkingDir, false, &git.CloneOptions{
			URL:           gitRepoUrl,
			SingleBranch:  true,
			ReferenceName: "HEAD",
			Depth:         1,
		})
		if err != nil {
			log.Fatalf("Error running git.PlainClone: %s", err)
		}
	}

	dirs := []string{}
	filepath.WalkDir(rootWorkingDir, func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf("Error running WalDir: %s", err)
		}

		dir := filepath.Dir(s)
		// Dir must contain .tf files
		// Dir must be unique within slice
		// Excluding dirs with ".terraform/" in their path, because they are download modules
		if filepath.Ext(d.Name()) == ".tf" && !slices.Contains(dirs, dir) && !strings.Contains(dir, ".terraform/") {
			dirs = append(dirs, dir)
		}

		return nil
	})

	return dirs
}
