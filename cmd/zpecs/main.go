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
	Version kong.VersionFlag `name:"version" help:"Print version information and quit"`
	Update  updateCmd        `cmd:"" help:"renders skills, agents, and docs"`
	Convert convertCmd       `cmd:"" help:"turns a spec markdown file into JSON"`
}

// updateCmd is the kong grammar for `zpecs update`.
type updateCmd struct {
	Scope  scope  `arg:"" default:"all" help:"render skills, agents, and docs, or one of them"`
	Source string `name:"source" env:"ZPECS_SOURCE" default:"${default_source}" help:"read definitions from a local directory, or clone it if it is a git repository"`
	Target target `name:"target" enum:"claude,opencode" default:"opencode" help:"render for claude or opencode"`
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

// printError writes err to stderr, then the usage text when err is a
// parse error. Runtime errors get no usage text.
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
	if err := updateScope(root, sourceDir, sourceLabel, opts.scope, string(opts.target)); err != nil {
		return err
	}
	if opts.scope == scopeAll {
		return updateScope(root, sourceDir, sourceLabel, scopeDocs, "")
	}
	return nil
}

// updateScope renders the definitions one scope selects into their target
// under root and reports the run. Docs have no target: they always write
// to docs/zpecs/.
func updateScope(root, sourceDir, sourceLabel string, s scope, target string) error {
	if s == scopeDocs {
		target = targetdir.Docs
	}
	defs, err := readDefinitions(s, sourceDir)
	if err != nil {
		return err
	}
	owned, err := targetdir.Owned(root, target)
	if err != nil {
		return err
	}
	if _, err := targetdir.RemoveStale(root, target, owned, defs, scopeKinds(s)...); err != nil {
		return fmt.Errorf("removing stale definitions: %w", err)
	}
	if err := targetdir.WriteAll(root, target, defs, func(d source.Definition) (string, error) {
		return render.Definition(d, target)
	}, owned); err != nil {
		return err
	}
	if err := targetdir.SaveOwned(root, target, owned); err != nil {
		return err
	}
	if s == scopeDocs {
		report.Summary(os.Stdout, s.String(), "", sourceLabel, len(defs))
		return nil
	}
	report.Summary(os.Stdout, s.String(), target, sourceLabel, len(defs))
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
