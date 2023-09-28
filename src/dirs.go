package src

import (
	"io/fs"
	"log"
	"path/filepath"
	"slices"
)

func FindTfDirs(rootWorkingDir string) []string {
	dirs := []string{}
	filepath.WalkDir(rootWorkingDir, func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf("Error running WalDir: %s", err)
		}

		dir := filepath.Dir(s)
		if filepath.Ext(d.Name()) == ".tf" && !slices.Contains(dirs, dir) {
			dirs = append(dirs, dir)
		}

		return nil
	})

	return dirs
}
