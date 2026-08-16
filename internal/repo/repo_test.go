package repo

import (
	"os"
	"path/filepath"
	"testing"
)

// gitRoot returns a temp dir that looks like a git repository root.
func gitRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRootAtRepositoryRoot(t *testing.T) {
	root := gitRoot(t)
	got, err := Root(root)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if got != root {
		t.Fatalf("Root = %q, want %q", got, root)
	}
}

func TestRootFindsRepositoryFromSubdirectory(t *testing.T) {
	root := gitRoot(t)
	sub := filepath.Join(root, "docs", "specs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Root(sub)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if got != root {
		t.Fatalf("Root = %q, want %q", got, root)
	}
}

func TestRootErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	if _, err := Root(dir); err == nil {
		t.Fatal("Root should error outside a repository")
	}
}
