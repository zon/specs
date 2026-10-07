package targetdir

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zon/specs/internal/source"
)

// ownedPath records one written path and its kind.
type ownedPath struct {
	kind source.Kind
	// known is false for a path read from a manifest written before
	// kinds were stored.
	known bool
}

// manifestName is the file inside a target directory that records
// ownership, one "kind path" line per written file, and the enabled
// features, one "feature name" line each.
const manifestName = ".zpecs"

// featureKeyword marks a manifest line that records an enabled feature.
const featureKeyword = "feature"

// Path returns the path under root where a definition writes, keyed by
// the source name rather than any rendered field.
func Path(root, name string, d source.Definition) string {
	return filepath.Join(root, RelPath(name, d))
}

// RelPath returns a definition's path relative to the repository root,
// joining the target's directory with the source's layout. A doc writes
// as-is, since the doc layout already names the docs directory.
func RelPath(name string, d source.Definition) string {
	if d.Kind == source.Doc {
		return source.RelPath(d)
	}
	return filepath.Join(targetDir(name), source.RelPath(d))
}

// targetDir returns the directory a target writes to.
func targetDir(name string) string {
	if name == source.Claude {
		return ".claude"
	}
	if name == source.Docs {
		return filepath.Join("docs", "zpecs")
	}
	return ".opencode"
}

// Owned returns the paths the system wrote under root for a target,
// with each path's kind. It reads the target's manifest. A target
// without a manifest owns nothing.
func Owned(root, name string) (map[string]ownedPath, error) {
	owned, _, err := readManifest(root, name)
	return owned, err
}

// Features returns the feature names a target's manifest records. A
// target without a manifest records none.
func Features(root, name string) ([]string, error) {
	_, features, err := readManifest(root, name)
	return features, err
}

// readManifest reads the owned paths and feature names from a target's
// manifest. A target without a manifest has neither.
func readManifest(root, name string) (map[string]ownedPath, []string, error) {
	data, err := os.ReadFile(filepath.Join(root, targetDir(name), manifestName))
	if os.IsNotExist(err) {
		return map[string]ownedPath{}, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	owned := map[string]ownedPath{}
	var features []string
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		if feature, ok := featureName(line); ok {
			features = append(features, feature)
			continue
		}
		kind, path, found := strings.Cut(line, " ")
		if k, ok := parseKind(kind); found && ok {
			owned[path] = ownedPath{kind: k, known: true}
			continue
		}
		owned[line] = ownedPath{}
	}
	return owned, features, nil
}

// featureName returns the feature name a manifest line records. A line
// records a feature when it is "feature" and one single-word name.
func featureName(line string) (string, bool) {
	keyword, name, found := strings.Cut(line, " ")
	if !found || keyword != featureKeyword || name == "" || strings.Contains(name, " ") {
		return "", false
	}
	return name, true
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

// SaveOwned persists the owned paths for a target under root, keeping
// the features its manifest already records.
func SaveOwned(root, name string, owned map[string]ownedPath) error {
	_, features, err := readManifest(root, name)
	if err != nil {
		return err
	}
	return SaveManifest(root, name, owned, features)
}

// SaveManifest persists the owned paths and enabled features for a
// target under root. When both are empty, it removes the manifest
// instead of writing an empty one.
func SaveManifest(root, name string, owned map[string]ownedPath, features []string) error {
	lines := ownedLines(owned)
	lines = append(lines, featureLines(features)...)
	if len(lines) == 0 {
		err := os.Remove(filepath.Join(root, targetDir(name), manifestName))
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	sort.Strings(lines)
	dir := filepath.Join(root, targetDir(name))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, manifestName), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// ownedLines returns one manifest line per owned path.
func ownedLines(owned map[string]ownedPath) []string {
	lines := make([]string, 0, len(owned))
	for p, op := range owned {
		if op.known {
			lines = append(lines, kindName(op.kind)+" "+p)
		} else {
			lines = append(lines, p)
		}
	}
	return lines
}

// featureLines returns one manifest line per feature name, once each.
func featureLines(features []string) []string {
	seen := make(map[string]bool, len(features))
	lines := make([]string, 0, len(features))
	for _, f := range features {
		if f == "" || seen[f] {
			continue
		}
		seen[f] = true
		lines = append(lines, featureKeyword+" "+f)
	}
	return lines
}

// RemoveStale deletes the files the system wrote under root for target
// that no current definition writes, limited to the selected kinds. An
// entry whose kind is unknown stays until a later write of the same
// path records its kind. It drops the removed paths from owned and
// returns them.
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
			return nil, fmt.Errorf("removing stale definitions: %w", err)
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
