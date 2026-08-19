package source

import (
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

// Definition is one skill, agent, or doc found in a source.
type Definition struct {
	Kind Kind
	Name string
	Path string
}

// ReadLocal reads definitions from a local source that follows the
// repository layout.
func ReadLocal(dir string) ([]Definition, error) {
	if err := checkDir(dir); err != nil {
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

// ReadSkills reads skill definitions from a local source.
func ReadSkills(dir string) ([]Definition, error) {
	if err := checkDir(dir); err != nil {
		return nil, err
	}
	return readSkills(dir)
}

// ReadAgents reads agent definitions from a local source.
func ReadAgents(dir string) ([]Definition, error) {
	if err := checkDir(dir); err != nil {
		return nil, err
	}
	return readAgents(dir)
}

// ReadDocs reads doc definitions from a local source.
func ReadDocs(dir string) ([]Definition, error) {
	if err := checkDir(dir); err != nil {
		return nil, err
	}
	return readDocs(dir)
}

func checkDir(dir string) error {
	_, err := os.Stat(dir)
	return err
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

func readDocs(dir string) ([]Definition, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "docs", "zpecs", "*.md"))
	if err != nil {
		return nil, err
	}
	defs := make([]Definition, 0, len(matches))
	for _, path := range matches {
		defs = append(defs, Definition{
			Kind: Doc,
			Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Path: path,
		})
	}
	return defs, nil
}
