package clone

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitRepo(t *testing.T, files ...string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("content\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-qm", "seed")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCloneCopiesRepository(t *testing.T) {
	src := gitRepo(t, filepath.Join("skills", "seed", "SKILL.md"))

	dir, cleanup, err := Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	defer cleanup()

	if _, err := os.Stat(filepath.Join(dir, "skills", "seed", "SKILL.md")); err != nil {
		t.Fatalf("clone missing the file: %v", err)
	}
}

func TestCloneCleanupRemovesDirectory(t *testing.T) {
	src := gitRepo(t, filepath.Join("skills", "seed", "SKILL.md"))

	dir, cleanup, err := Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	cleanup()

	if _, err := os.Stat(dir); err == nil {
		t.Fatal("cleanup left the clone directory behind")
	}
}
