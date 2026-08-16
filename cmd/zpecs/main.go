package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/repo"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

const usage = `zpecs renders skills and agents for claude and opencode

usage:
  zpecs update                    renders skills and agents
  zpecs update skills             renders skills only
  zpecs update agents             renders agents only

flags:
  --source DIR    read definitions from the local directory DIR (default: GitHub)
  --target NAME   render for claude or opencode (default: opencode)
`

type scope int

const (
	scopeAll scope = iota
	scopeSkills
	scopeAgents
)

func (s scope) String() string {
	switch s {
	case scopeSkills:
		return "skills"
	case scopeAgents:
		return "agents"
	default:
		return "skills and agents"
	}
}

type target string

const (
	targetClaude   target = "claude"
	targetOpencode target = "opencode"
)

func parseTarget(s string) (target, error) {
	switch target(s) {
	case targetClaude, targetOpencode:
		return target(s), nil
	default:
		return "", fmt.Errorf("unknown target %q (want claude or opencode)", s)
	}
}

type options struct {
	scope  scope
	source string
	target target
}

func parseOptions(args []string) (options, error) {
	var (
		opts options
		pos  []string
	)
	opts.scope = scopeAll
	opts.target = targetOpencode
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--source":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--source needs a directory path")
			}
			i++
			opts.source = args[i]
		case strings.HasPrefix(arg, "--source="):
			opts.source = strings.TrimPrefix(arg, "--source=")
		case arg == "--target":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--target needs a target name")
			}
			i++
			t, err := parseTarget(args[i])
			if err != nil {
				return opts, err
			}
			opts.target = t
		case strings.HasPrefix(arg, "--target="):
			t, err := parseTarget(strings.TrimPrefix(arg, "--target="))
			if err != nil {
				return opts, err
			}
			opts.target = t
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown flag %q", arg)
		default:
			pos = append(pos, arg)
		}
	}
	s, err := scopeFromArgs(pos)
	if err != nil {
		return opts, err
	}
	opts.scope = s
	return opts, nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "zpecs:", err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}
	if args[0] != "update" {
		return fmt.Errorf("unknown command %q", args[0])
	}
	return update(args[1:])
}

func update(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	root, err := repo.Root(".")
	if err != nil {
		return err
	}
	if opts.source == "" {
		fmt.Printf("updating %s for %s\n", opts.scope, opts.target)
		return nil
	}
	defs, err := readDefinitions(opts.scope, opts.source)
	if err != nil {
		return err
	}
	owned, err := targetdir.Owned(root, string(opts.target))
	if err != nil {
		return err
	}
	if _, err := targetdir.RemoveStale(root, string(opts.target), owned, defs, scopeKinds(opts.scope)...); err != nil {
		return fmt.Errorf("removing stale definitions: %w", err)
	}
	for _, d := range defs {
		content, err := rendered(d, opts.target)
		if err != nil {
			return err
		}
		if _, err := targetdir.Write(root, string(opts.target), d, content, owned); err != nil {
			return fmt.Errorf("writing %s: %w", targetdir.Path(root, string(opts.target), d), err)
		}
	}
	if err := targetdir.SaveOwned(root, string(opts.target), owned); err != nil {
		return err
	}
	fmt.Printf("updating %s for %s from %s (%d definitions)\n", opts.scope, opts.target, opts.source, len(defs))
	return nil
}

// readDefinitions returns the definitions a scope selects from the source.
func readDefinitions(s scope, sourceDir string) ([]source.Definition, error) {
	switch s {
	case scopeSkills:
		return source.ReadSkills(sourceDir)
	case scopeAgents:
		return source.ReadAgents(sourceDir)
	default:
		return source.ReadLocal(sourceDir)
	}
}

// scopeKinds returns the definition kinds a scope selects.
func scopeKinds(s scope) []source.Kind {
	switch s {
	case scopeSkills:
		return []source.Kind{source.Skill}
	case scopeAgents:
		return []source.Kind{source.Agent}
	default:
		return []source.Kind{source.Skill, source.Agent}
	}
}

// rendered returns a skill unchanged or an agent built from its frontmatter.
func rendered(d source.Definition, t target) (string, error) {
	if d.Kind == source.Skill {
		raw, err := os.ReadFile(d.Path)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", d.Path, err)
		}
		return string(raw), nil
	}
	content, err := frontmatter.Read(d.Path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", d.Path, err)
	}
	if t == targetClaude {
		return render.ClaudeAgent(content.Fields, content.Body), nil
	}
	return render.OpencodeAgent(content.Fields, content.Body), nil
}

func scopeFromArgs(args []string) (scope, error) {
	switch len(args) {
	case 0:
		return scopeAll, nil
	case 1:
		switch args[0] {
		case "skills":
			return scopeSkills, nil
		case "agents":
			return scopeAgents, nil
		}
		return scopeAll, fmt.Errorf("unknown update scope %q", args[0])
	default:
		return scopeAll, fmt.Errorf("update takes at most one scope")
	}
}
