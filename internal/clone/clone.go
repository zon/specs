package clone

import (
	"fmt"
	"os"
	"os/exec"
)

// Clone copies the git repository at url into a temp dir and returns
// that dir with a cleanup func that removes it.
func Clone(url string) (dir string, cleanup func(), err error) {
	dir, err = os.MkdirTemp("", "zpecs-source-*")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { os.RemoveAll(dir) }
	cmd := exec.Command("git", "clone", "--depth", "1", url, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("cloning %s: %v\n%s", url, err, out)
	}
	return dir, cleanup, nil
}
