package targetdir

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zon/specs/internal/source"
)

// Target names.
const (
	Claude   = "claude"
	Opencode = "opencode"
)

// manifestName is the file inside a target directory that records
// ownership.
const manifestName = ".zpecs"

// Path returns the path under root where a definition writes, keyed by
// the source name rather than any rendered field.
func Path(root, name string, d source.Definition) string {
	return filepath.Join(root, RelPath(name, d))
}

// RelPath returns the path of a definition's written file under its
// target, relative to the repository root.
func RelPath(name string, d source.Definition) string {
	target := targetDir(name)
	if d.Kind == source.Skill {
		return filepath.Join(target, "skills", d.Name, "SKILL.md")
	}
	return filepath.Join(target, "agents", d.Name+".md")
}

// targetDir returns the hidden directory a target writes to.
func targetDir(name string) string {
	if name == Claude {
		return ".claude"
	}
	return ".opencode"
}

// Owned returns the set of paths the system wrote under root for a
// target. It reads the target's manifest. A target without a manifest
// owns nothing.
func Owned(root, target string) (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(root, targetDir(target), manifestName))
	if os.IsNotExist(err) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	owned := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			owned[line] = true
		}
	}
	return owned, nil
}

// Write stores content at the definition's path under root, creating the
// directories it needs. It replaces a file only when the system wrote it
// before. A foreign file stays untouched. It records the written path in
// owned and reports whether it wrote the file.
func Write(root, name string, d source.Definition, content string, owned map[string]bool) (bool, error) {
	p := Path(root, name, d)
	rel := RelPath(name, d)
	if _, err := os.Stat(p); err == nil && !owned[rel] {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return false, err
	}
	owned[rel] = true
	return true, nil
}

// SaveOwned persists the owned paths for a target under root.
func SaveOwned(root, target string, owned map[string]bool) error {
	lines := make([]string, 0, len(owned))
	for p := range owned {
		lines = append(lines, p)
	}
	sort.Strings(lines)
	dir := filepath.Join(root, targetDir(target))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, manifestName), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
