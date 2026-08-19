package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zon/specs/internal/clone"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/repo"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

type scope int

const (
	scopeAll scope = iota
	scopeSkills
	scopeAgents
	scopeDocs
)

func (s scope) String() string {
	switch s {
	case scopeSkills:
		return "skills"
	case scopeAgents:
		return "agents"
	case scopeDocs:
		return "docs"
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
	case "docs":
		return scopeDocs, nil
	default:
		return scopeAll, fmt.Errorf("unknown update scope %q", s)
	}
}

type target string

const (
	targetClaude   target = targetdir.Claude
	targetOpencode target = targetdir.Opencode
)

type options struct {
	scope  scope
	source string
	target target
}

// cli is the kong grammar for the whole application.
type cli struct {
	Update updateCmd `cmd:"" help:"renders skills and agents, or syncs docs"`
}

// updateCmd is the kong grammar for `zpecs update`.
type updateCmd struct {
	Scope  string `arg:"" optional:"" default:"" help:"render skills or agents, or sync docs (default: skills and agents)"`
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
	target := opts.target
	if opts.scope == scopeDocs {
		target = targetdir.Docs
	}
	owned, err := targetdir.Owned(root, string(target))
	if err != nil {
		return err
	}
	if _, err := targetdir.RemoveStale(root, string(target), owned, defs, scopeKinds(opts.scope)...); err != nil {
		return fmt.Errorf("removing stale definitions: %w", err)
	}
	if err := targetdir.WriteAll(root, string(target), defs, func(d source.Definition) (string, error) {
		return render.Definition(d, string(opts.target))
	}, owned); err != nil {
		return err
	}
	if err := targetdir.SaveOwned(root, string(target), owned); err != nil {
		return err
	}
	if opts.scope == scopeDocs {
		report.Summary(os.Stdout, opts.scope.String(), "", sourceLabel, len(defs))
	} else {
		report.Summary(os.Stdout, opts.scope.String(), string(opts.target), sourceLabel, len(defs))
	}
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
	case scopeDocs:
		return source.ReadDocs(sourceDir)
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
	case scopeDocs:
		return []source.Kind{source.Doc}
	default:
		return []source.Kind{source.Skill, source.Agent}
	}
}
