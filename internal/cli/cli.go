// Package cli parses the command line and dispatches each command's work.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zon/specs/internal/spec"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/update"
)

// version is the CLI version. Set it at build time with
// -ldflags "-X github.com/zon/specs/internal/cli.version=...".
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

// scopeWord returns the scope's word for the update module.
func scopeWord(s scope) string {
	switch s {
	case scopeSkills:
		return "skills"
	case scopeAgents:
		return "agents"
	case scopeDocs:
		return "docs"
	default:
		return "all"
	}
}

type targetName string

const (
	targetClaude   targetName = target.Claude
	targetOpencode targetName = target.Opencode
)

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
	return update.Run(update.Options{
		Scope:  scopeWord(u.Scope),
		Source: u.Source,
		Target: string(u.Target),
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

// Main runs the CLI with args, excluding the program name. It prints
// errors and returns the process exit code.
func Main(args []string) int {
	if err := run(args); err != nil {
		printError(err)
		return 1
	}
	return 0
}
