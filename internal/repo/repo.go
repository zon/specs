package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

// Root returns the root of the git repository containing dir. It walks
// up until a .git entry appears, and errors at the filesystem root
// first.
func Root(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, ".git")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("not inside a git repository")
		}
		abs = parent
	}
}
