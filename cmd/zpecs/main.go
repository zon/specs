package main

import (
	"fmt"
	"os"
)

const usage = `zpecs renders skill and agent definitions for the claude and opencode targets

usage:
  zpecs update            renders skills and agents
  zpecs update skills     renders skills only
  zpecs update agents     renders agents only
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
	s, err := scopeFromArgs(args)
	if err != nil {
		return err
	}
	fmt.Println("updating", s)
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
