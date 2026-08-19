package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zon/specs/internal/clone"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/repo"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/spec"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/targetdir"
)

// version is the CLI version. Set it at build time with
// -ldflags "-X main.version=...".
var version = "dev"

// defaultSourceURL is the repository the CLI reads when neither a
// --source flag nor ZPECS_SOURCE names another source.
const defaultSourceURL = "https://github.com/zon/specs"

// cliVars feeds kong's ${...} interpolation in the grammar.
var cliVars = kong.Vars{
	"version":        version,
	"default_source": defaultSourceURL,
}

type scope int

const (
	scopeAll scope = iota
	scopeSkills
	scopeAgents
	scopeDocs
)

func (s *scope) UnmarshalText(text []byte) error {
	switch string(text) {
	case "all":
		*s = scopeAll
	case "skills":
		*s = scopeSkills
	case "agents":
		*s = scopeAgents
	case "docs":
		*s = scopeDocs
	default:
		return fmt.Errorf("unknown scope %q", string(text))
	}
	return nil
}

type targetName string

const (
	targetClaude   targetName = target.Claude
	targetOpencode targetName = target.Opencode
)

type options struct {
	scope  scope
	source string
	target targetName
}

// pair is one run of the sync pipeline: the name to report, the target
// to write to, and the kinds it selects.
type pair struct {
	name   string
	target string
	kinds  []source.Kind
}

// pairs selects the runs for a scope. Skills and agents write to the
// command's target. Docs always write to docs/zpecs.
func pairs(s scope, runner string) []pair {
	switch s {
	case scopeSkills:
		return []pair{{name: "skills", target: runner, kinds: []source.Kind{source.Skill}}}
	case scopeAgents:
		return []pair{{name: "agents", target: runner, kinds: []source.Kind{source.Agent}}}
	case scopeDocs:
		return []pair{{name: "docs", target: target.Docs, kinds: []source.Kind{source.Doc}}}
	default:
		return []pair{
			{name: "skills and agents", target: runner, kinds: []source.Kind{source.Skill, source.Agent}},
			{name: "docs", target: target.Docs, kinds: []source.Kind{source.Doc}},
		}
	}
}

// cli is the kong grammar for the whole application.
type cli struct {
	Version kong.VersionFlag `name:"version" help:"Print version information and quit"`
	Update  updateCmd        `cmd:"" help:"renders skills, agents, and docs"`
	Convert convertCmd       `cmd:"" help:"turns a spec markdown file into JSON"`
}

// updateCmd is the kong grammar for `zpecs update`.
type updateCmd struct {
	Scope  scope      `arg:"" default:"all" help:"render skills, agents, and docs, or one of them"`
	Source string     `name:"source" env:"ZPECS_SOURCE" default:"${default_source}" help:"read definitions from a local directory, or clone it if it is a git repository"`
	Target targetName `name:"target" enum:"claude,opencode" default:"opencode" help:"render for claude or opencode"`
}

func (u *updateCmd) Run() error {
	return update(options{
		scope:  u.Scope,
		source: u.Source,
		target: u.Target,
	})
}

// convertCmd is the kong grammar for `zpecs convert`.
type convertCmd struct {
	Path string `arg:"" help:"path to a spec markdown file"`
}

func (c *convertCmd) Run() error {
	doc, err := spec.Read(c.Path)
	if err != nil {
		return err
	}
	return spec.Write(os.Stdout, doc)
}

// usageText returns the help kong generates for the whole application.
// The --help hook prints it and calls Exit. Here Exit is a no-op, so
// Parse continues and fails on the missing command.
func usageText() string {
	var sb strings.Builder
	var c cli
	parser, err := kong.New(&c, cliVars, kong.Writers(&sb, &sb), kong.Exit(func(int) {}))
	if err != nil {
		return ""
	}
	_, _ = parser.Parse([]string{"--help"})
	return sb.String()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		printError(err)
		os.Exit(1)
	}
}

// printError writes err to stderr, then the usage text for parse errors.
func printError(err error) {
	fmt.Fprintln(os.Stderr, "zpecs:", err)
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, usageText())
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usageText())
		return nil
	}
	var c cli
	parser, err := kong.New(&c, cliVars)
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
	for _, p := range pairs(opts.scope, string(opts.target)) {
		if err := updatePair(root, sourceDir, sourceLabel, p); err != nil {
			return err
		}
	}
	return nil
}

// updatePair renders a pair's definitions into its target under root.
// It then reports the run.
func updatePair(root, sourceDir, sourceLabel string, p pair) error {
	defs, err := source.ReadKinds(p.kinds, sourceDir)
	if err != nil {
		return err
	}
	owned, err := targetdir.Owned(root, p.target)
	if err != nil {
		return err
	}
	if _, err := targetdir.RemoveStale(root, p.target, owned, defs, p.kinds...); err != nil {
		return fmt.Errorf("removing stale definitions: %w", err)
	}
	if err := targetdir.WriteAll(root, p.target, defs, func(d source.Definition) (string, error) {
		return render.Definition(d, p.target)
	}, owned); err != nil {
		return err
	}
	if err := targetdir.SaveOwned(root, p.target, owned); err != nil {
		return err
	}
	report.Summary(os.Stdout, p.name, p.target, sourceLabel, len(defs))
	return nil
}

// resolveSource returns the directory the definitions come from, the
// label to report, and a cleanup func. Kong supplies the source: the
// default repository URL, a ZPECS_SOURCE override, or a --source path.
// A value with a scheme is a repository to clone. Anything else is a
// local directory to read in place.
func resolveSource(source string) (dir, label string, cleanup func(), err error) {
	if strings.Contains(source, "://") {
		dir, cleanup, err = clone.Clone(source)
		if err != nil {
			return "", "", nil, err
		}
		return dir, source, cleanup, nil
	}
	return source, source, func() {}, nil
}
