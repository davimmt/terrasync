package src

import (
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"
)

func FindTfDirs(rootWorkingDir string) []string {
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
