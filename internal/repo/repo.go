package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

// Root returns the root of the git repository containing the process
// working directory. It walks up until a .git entry appears, and errors
// when it reaches the filesystem root without finding one.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository")
		}
		dir = parent
	}
}
