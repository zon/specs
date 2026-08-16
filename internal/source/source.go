package source

import (
	"os"
	"path/filepath"
	"strings"
)

// Kind distinguishes a skill definition from an agent definition.
type Kind int

const (
	// Skill is a definition at skills/<name>/SKILL.md.
	Skill Kind = iota
	// Agent is a definition at agents/<name>.md.
	Agent
)

// Definition is one skill or agent found in a source.
type Definition struct {
	Kind Kind
	Name string
	Path string
}

// ReadLocal reads definitions from a local source. The source follows the
// repository layout, so it ignores anything in other locations.
func ReadLocal(dir string) ([]Definition, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	defs, err := readSkills(dir)
	if err != nil {
		return nil, err
	}
	agents, err := readAgents(dir)
	if err != nil {
		return nil, err
	}
	return append(defs, agents...), nil
}

func readSkills(dir string) ([]Definition, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "skills", "*", "SKILL.md"))
	if err != nil {
		return nil, err
	}
	defs := make([]Definition, 0, len(matches))
	for _, path := range matches {
		defs = append(defs, Definition{
			Kind: Skill,
			Name: filepath.Base(filepath.Dir(path)),
			Path: path,
		})
	}
	return defs, nil
}

func readAgents(dir string) ([]Definition, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "agents", "*.md"))
	if err != nil {
		return nil, err
	}
	defs := make([]Definition, 0, len(matches))
	for _, path := range matches {
		defs = append(defs, Definition{
			Kind: Agent,
			Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Path: path,
		})
	}
	return defs, nil
}
