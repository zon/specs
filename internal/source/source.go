package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Kind marks a definition as a skill, an agent, or a doc.
type Kind int

const (
	// Skill is a definition at skills/<name>/SKILL.md.
	Skill Kind = iota
	// Agent is a definition at agents/<name>.md.
	Agent
	// Doc is a definition at docs/zpecs/<name>.md.
	Doc
)

// Scope selects the kinds of definitions to read from a source.
type Scope int

const (
	// ScopeAll selects every kind of definition.
	ScopeAll Scope = iota
	// ScopeSkills selects skill definitions.
	ScopeSkills
	// ScopeAgents selects agent definitions.
	ScopeAgents
	// ScopeDocs selects doc definitions.
	ScopeDocs
)

// UnmarshalText maps a scope token to its constant.
func (s *Scope) UnmarshalText(text []byte) error {
	switch string(text) {
	case "all":
		*s = ScopeAll
	case "skills":
		*s = ScopeSkills
	case "agents":
		*s = ScopeAgents
	case "docs":
		*s = ScopeDocs
	default:
		return fmt.Errorf("unknown scope %q", string(text))
	}
	return nil
}

// Definition is one skill, agent, or doc found in a source.
type Definition struct {
	Kind Kind
	Name string
	Path string
}

// kindSpec locates and names the definitions of a kind in a source.
type kindSpec struct {
	pattern string
	name    func(path string) string
}

// kindSpecs maps each kind to its spec.
var kindSpecs = map[Kind]kindSpec{
	Skill: {pattern: filepath.Join("skills", "*", "SKILL.md"), name: skillName},
	Agent: {pattern: filepath.Join("agents", "*.md"), name: baseName},
	Doc:   {pattern: filepath.Join("docs", "zpecs", "*.md"), name: baseName},
}

// RelPath returns a definition's path in its kind's layout.
// Unknown kinds fall back to the agent path.
func RelPath(d Definition) string {
	switch d.Kind {
	case Skill:
		return filepath.Join("skills", d.Name, "SKILL.md")
	case Doc:
		return filepath.Join("docs", "zpecs", d.Name+".md")
	default:
		return filepath.Join("agents", d.Name+".md")
	}
}

// ReadKinds reads the listed kinds from a local source, in that order.
func ReadKinds(kinds []Kind, dir string) ([]Definition, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	var defs []Definition
	for _, kind := range kinds {
		spec, ok := kindSpecs[kind]
		if !ok {
			return nil, fmt.Errorf("unknown kind %d", kind)
		}
		got, err := readKind(kind, spec.pattern, spec.name, dir)
		if err != nil {
			return nil, err
		}
		defs = append(defs, got...)
	}
	return defs, nil
}

// readKind finds the definitions matching a pattern in a source and names
// each one with the name function.
func readKind(kind Kind, pattern string, name func(string) string, dir string) ([]Definition, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}
	defs := make([]Definition, 0, len(matches))
	for _, path := range matches {
		defs = append(defs, Definition{
			Kind: kind,
			Name: name(path),
			Path: path,
		})
	}
	return defs, nil
}

// skillName is the skill directory's name below skills/.
func skillName(path string) string {
	return filepath.Base(filepath.Dir(path))
}

// baseName is the file's name without its extension.
func baseName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}
