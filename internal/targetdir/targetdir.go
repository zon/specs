package targetdir

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

// ownedPath records one written path and its kind.
type ownedPath struct {
	kind source.Kind
	// known is false for a path read from a manifest written before
	// kinds were stored.
	known bool
}

// manifestName is the file inside a target directory that records
// ownership, one "kind path" line per written file.
const manifestName = ".zpecs"

// Path returns the path under root where a definition writes, keyed by
// the source name rather than any rendered field.
func Path(root, name string, d source.Definition) string {
	return filepath.Join(root, RelPath(name, d))
}

// RelPath returns the path of a definition's written file under its
// target, relative to the repository root.
func RelPath(name string, d source.Definition) string {
	dir := targetDir(name)
	if d.Kind == source.Skill {
		return filepath.Join(dir, "skills", d.Name, "SKILL.md")
	}
	if d.Kind == source.Doc {
		return filepath.Join(dir, d.Name+".md")
	}
	return filepath.Join(dir, "agents", d.Name+".md")
}

// targetDir returns the directory a target writes to.
func targetDir(name string) string {
	if name == target.Claude {
		return ".claude"
	}
	if name == target.Docs {
		return filepath.Join("docs", "zpecs")
	}
	return ".opencode"
}

// Owned returns the paths the system wrote under root for a target,
// with each path's kind. It reads the target's manifest. A target
// without a manifest owns nothing.
func Owned(root, name string) (map[string]ownedPath, error) {
	data, err := os.ReadFile(filepath.Join(root, targetDir(name), manifestName))
	if os.IsNotExist(err) {
		return map[string]ownedPath{}, nil
	}
	if err != nil {
		return nil, err
	}
	owned := map[string]ownedPath{}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		kind, path, found := strings.Cut(line, " ")
		if k, ok := parseKind(kind); found && ok {
			owned[path] = ownedPath{kind: k, known: true}
			continue
		}
		owned[line] = ownedPath{}
	}
	return owned, nil
}

// Write stores content at the definition's path under root, creating the
// directories it needs. It replaces a file only when the system wrote it
// before. It records the written path in owned.
func Write(root, name string, d source.Definition, content string, owned map[string]ownedPath) error {
	p := Path(root, name, d)
	rel := RelPath(name, d)
	if _, err := os.Stat(p); err == nil {
		if _, ok := owned[rel]; !ok {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return err
	}
	owned[rel] = ownedPath{kind: d.Kind, known: true}
	return nil
}

// WriteAll writes every definition in defs under root for target, using
// content to produce each file's text. It follows the same owned-file
// rules as Write.
func WriteAll(root, name string, defs []source.Definition, content func(source.Definition) (string, error), owned map[string]ownedPath) error {
	for _, d := range defs {
		text, err := content(d)
		if err != nil {
			return err
		}
		if err := Write(root, name, d, text, owned); err != nil {
			return fmt.Errorf("writing %s: %w", Path(root, name, d), err)
		}
	}
	return nil
}

// SaveOwned persists the owned paths for a target under root.
func SaveOwned(root, name string, owned map[string]ownedPath) error {
	lines := make([]string, 0, len(owned))
	for p, op := range owned {
		if op.known {
			lines = append(lines, kindName(op.kind)+" "+p)
		} else {
			lines = append(lines, p)
		}
	}
	sort.Strings(lines)
	dir := filepath.Join(root, targetDir(name))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, manifestName), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// RemoveStale deletes for target under root the files owned records that
// no definition in current writes, limited to the selected kinds. An
// entry whose kind is unknown stays until a later write of the same
// path records its kind. RemoveStale drops the removed paths from owned
// and returns them.
func RemoveStale(root, name string, owned map[string]ownedPath, current []source.Definition, kinds ...source.Kind) ([]string, error) {
	written := make(map[string]bool, len(current))
	for _, d := range current {
		written[RelPath(name, d)] = true
	}
	selected := make(map[source.Kind]bool, len(kinds))
	for _, k := range kinds {
		selected[k] = true
	}
	var removed []string
	for rel, op := range owned {
		if written[rel] {
			continue
		}
		if !op.known || !selected[op.kind] {
			continue
		}
		if err := os.Remove(filepath.Join(root, rel)); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("removing stale %s: %w", filepath.Join(root, rel), err)
		}
		delete(owned, rel)
		removed = append(removed, rel)
	}
	return removed, nil
}

// kindNames maps each kind to its manifest word.
var kindNames = map[source.Kind]string{
	source.Skill: "skill",
	source.Agent: "agent",
	source.Doc:   "doc",
}

// kindName returns the manifest word for a kind.
func kindName(k source.Kind) string {
	return kindNames[k]
}

// parseKind returns the kind a manifest word names.
func parseKind(s string) (source.Kind, bool) {
	for k, word := range kindNames {
		if word == s {
			return k, true
		}
	}
	return 0, false
}
