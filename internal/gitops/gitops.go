package gitops

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Root returns the root of the git repository containing the process
// working directory. It walks up until a .git entry appears, or errors
// at the filesystem root.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		_, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil {
			return dir, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository")
		}
		dir = parent
	}
}

// IsRemote reports whether source names a git repository by URL.
func IsRemote(source string) bool {
	return strings.Contains(source, "://")
}

// Clone copies the git repository at url into a temp dir and returns a
// cleanup func that removes it.
func Clone(url string) (dir string, cleanup func(), err error) {
	dir, err = os.MkdirTemp("", "zpecs-source-*")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { os.RemoveAll(dir) }
	cmd := exec.Command("git", "clone", "--depth", "1", url, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("cloning %s: %w\n%s", url, err, strings.TrimSpace(string(out)))
	}
	return dir, cleanup, nil
}
