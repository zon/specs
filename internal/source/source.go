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

// Definition is one skill, agent, or doc found in a source.
type Definition struct {
	Kind Kind
	Name string
	Path string
}

// ReadKinds reads the listed kinds from a local source, in that order.
func ReadKinds(kinds []Kind, dir string) ([]Definition, error) {
	if err := checkDir(dir); err != nil {
		return nil, err
	}
	var defs []Definition
	for _, kind := range kinds {
		var got []Definition
		var err error
		switch kind {
		case Skill:
			got, err = readSkills(dir)
		case Agent:
			got, err = readAgents(dir)
		case Doc:
			got, err = readDocs(dir)
		default:
			return nil, fmt.Errorf("unknown kind %d", kind)
		}
		if err != nil {
			return nil, err
		}
		defs = append(defs, got...)
	}
	return defs, nil
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
