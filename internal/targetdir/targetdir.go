package targetdir

import (
	"fmt"
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
	Docs     = "docs"
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
	if d.Kind == source.Doc {
		return filepath.Join(target, d.Name+".md")
	}
	return filepath.Join(target, "agents", d.Name+".md")
}

// targetDir returns the directory a target writes to.
func targetDir(name string) string {
	if name == Claude {
		return ".claude"
	}
	if name == Docs {
		return filepath.Join("docs", "zpecs")
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

// WriteAll writes every definition in defs under root for target, using
// content to produce each file's text. It follows the same owned-file
// rules as Write and records what it writes in owned.
func WriteAll(root, name string, defs []source.Definition, content func(source.Definition) (string, error), owned map[string]bool) error {
	for _, d := range defs {
		text, err := content(d)
		if err != nil {
			return err
		}
		if _, err := Write(root, name, d, text, owned); err != nil {
			return fmt.Errorf("writing %s: %w", Path(root, name, d), err)
		}
	}
	return nil
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

// RemoveStale deletes for target under root the files owned records that
// no definition in current writes, limited to the kinds kinds selects. It
// drops the removed paths from owned and returns them.
func RemoveStale(root, target string, owned map[string]bool, current []source.Definition, kinds ...source.Kind) ([]string, error) {
	written := make(map[string]bool, len(current))
	for _, d := range current {
		written[RelPath(target, d)] = true
	}
	selected := make(map[source.Kind]bool, len(kinds))
	for _, k := range kinds {
		selected[k] = true
	}
	var removed []string
	for rel := range owned {
		if written[rel] {
			continue
		}
		kind, ok := pathKind(target, rel)
		if !ok || !selected[kind] {
			continue
		}
		if err := os.Remove(filepath.Join(root, rel)); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		delete(owned, rel)
		removed = append(removed, rel)
	}
	return removed, nil
}

// pathKind returns the kind of a relative path under a target, reporting
// false when the system would not write the path.
func pathKind(target, rel string) (source.Kind, bool) {
	dir := targetDir(target) + string(filepath.Separator)
	switch {
	case strings.HasPrefix(rel, dir+"skills"+string(filepath.Separator)) &&
		strings.HasSuffix(rel, string(filepath.Separator)+"SKILL.md"):
		return source.Skill, true
	case strings.HasPrefix(rel, dir+"agents"+string(filepath.Separator)) &&
		strings.HasSuffix(rel, ".md"):
		return source.Agent, true
	case strings.HasPrefix(rel, dir) &&
		strings.HasSuffix(rel, ".md"):
		return source.Doc, true
	}
	return 0, false
}
