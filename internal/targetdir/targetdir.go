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

// Path returns where a definition writes under the target directory, keyed
// by the definition's source name rather than by any rendered field.
func Path(name string, d source.Definition) string {
	root := ".opencode"
	if name == Claude {
		root = ".claude"
	}
	if d.Kind == source.Skill {
		return filepath.Join(root, "skills", d.Name, "SKILL.md")
	}
	return filepath.Join(root, "agents", d.Name+".md")
}

// Write writes content to a definition's path under the target directory,
// creating the directories it needs.
func Write(name string, d source.Definition, content string) error {
	p := Path(name, d)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), 0o644)
}
