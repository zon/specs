package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zon/specs/internal/clone"
	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/repo"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

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

func parseScope(s string) (scope, error) {
	switch s {
	case "":
		return scopeAll, nil
	case "skills":
		return scopeSkills, nil
	case "agents":
		return scopeAgents, nil
	default:
		return scopeAll, fmt.Errorf("unknown update scope %q", s)
	}
}

type target string

const (
	targetClaude   target = "claude"
	targetOpencode target = "opencode"
)

type options struct {
	scope  scope
	source string
	target target
}

// cli is the kong grammar for the whole application.
type cli struct {
	Update updateCmd `cmd:"" help:"renders skills and agents"`
}

// updateCmd is the kong grammar for `zpecs update`.
type updateCmd struct {
	Scope  string `arg:"" optional:"" default:"" help:"render skills or agents (default: both)"`
	Source string `name:"source" help:"read definitions from the local directory DIR (default: GitHub)"`
	Target string `name:"target" enum:"claude,opencode" default:"opencode" help:"render for claude or opencode"`
}

func (u *updateCmd) Run() error {
	s, err := parseScope(u.Scope)
	if err != nil {
		return err
	}
	return update(options{
		scope:  s,
		source: u.Source,
		target: target(u.Target),
	})
}

// usageText returns the help kong generates for the whole application.
// The --help hook prints help and stops. With Exit overridden to a no-op,
// Parse continues and fails on the missing command.
func usageText() string {
	var sb strings.Builder
	var c cli
	parser, err := kong.New(&c, kong.Writers(&sb, &sb), kong.Exit(func(int) {}))
	if err != nil {
		return ""
	}
	_, _ = parser.Parse([]string{"--help"})
	return sb.String()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "zpecs:", err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, usageText())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usageText())
		return nil
	}
	var c cli
	parser, err := kong.New(&c)
	if err != nil {
		return err
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	return ctx.Run()
}

func update(opts options) error {
	root, err := repo.Root(".")
	if err != nil {
		return err
	}
	sourceDir, sourceLabel, cleanup, err := resolveSource(opts.source)
	if err != nil {
		return err
	}
	defer cleanup()
	defs, err := readDefinitions(opts.scope, sourceDir)
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
	fmt.Printf("updating %s for %s from %s (%d definitions)\n", opts.scope, opts.target, sourceLabel, len(defs))
	return nil
}

// defaultSource is the GitHub repository the CLI reads definitions from
// when run without a --source flag. ZPECS_SOURCE overrides it, mostly for
// tests.
func defaultSource() string {
	if v := os.Getenv("ZPECS_SOURCE"); v != "" {
		return v
	}
	return "https://github.com/zon/specs"
}

// resolveSource returns the directory definitions come from, the label
// to report, and a cleanup func. A --source flag names a local
// directory. Without one, the default source is cloned into a temp dir.
func resolveSource(flag string) (dir, label string, cleanup func(), err error) {
	if flag != "" {
		return flag, flag, func() {}, nil
	}
	url := defaultSource()
	dir, cleanup, err = clone.Clone(url)
	if err != nil {
		return "", "", nil, err
	}
	return dir, url, cleanup, nil
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
