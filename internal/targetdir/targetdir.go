package targetdir

import (
	"os"
	"path/filepath"

	"github.com/zon/specs/internal/source"
)

// Target names.
const (
	Claude   = "claude"
	Opencode = "opencode"
)

// Path returns the path under root where a definition writes, keyed by
// the source name rather than any rendered field.
func Path(root, name string, d source.Definition) string {
	target := ".opencode"
	if name == Claude {
		target = ".claude"
	}
	if d.Kind == source.Skill {
		return filepath.Join(root, target, "skills", d.Name, "SKILL.md")
	}
	return filepath.Join(root, target, "agents", d.Name+".md")
}

// Write stores content at the definition's path under root, creating the
// directories it needs.
func Write(root, name string, d source.Definition, content string) error {
	p := Path(root, name, d)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), 0o644)
}
