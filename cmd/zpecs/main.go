package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/source"
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
	if opts.source == "" {
		fmt.Printf("updating %s for %s\n", opts.scope, opts.target)
		return nil
	}
	defs, err := source.ReadLocal(opts.source)
	if err != nil {
		return err
	}
	for _, d := range defs {
		if _, err := frontmatter.Read(d.Path); err != nil {
			return fmt.Errorf("reading %s: %w", d.Path, err)
		}
	}
	fmt.Printf("updating %s for %s from %s (%d definitions)\n", opts.scope, opts.target, opts.source, len(defs))
	return nil
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
