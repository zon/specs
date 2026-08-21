package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zon/specs/internal/opencode"
	"github.com/zon/specs/internal/review"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/spec"
	"github.com/zon/specs/internal/update"
)

// version is the CLI version. Set it at build time with
// -ldflags "-X main.version=...".
var version = "dev"

// defaultSourceURL is the repository the CLI reads when neither a
// --source flag nor ZPECS_SOURCE names another source.
const defaultSourceURL = "https://github.com/zon/specs"

// cliVars feeds kong's ${...} interpolation in the grammar.
var cliVars = kong.Vars{
	"version":         version,
	"default_source":  defaultSourceURL,
	"default_model":   opencode.DefaultModel,
	"default_variant": opencode.DefaultVariant,
}

// cli is the kong grammar for the whole application.
type cli struct {
	Version kong.VersionFlag `name:"version" help:"Print version information and quit"`
	Update  updateCmd        `cmd:"" help:"renders skills, agents, and docs"`
	Review  reviewCmd        `cmd:"" help:"reviews the repository against the code, architecture, or prose guidelines"`
	Convert convertCmd       `cmd:"" help:"turns a spec markdown file into JSON"`
}

// updateCmd is the kong grammar for `zpecs update`.
type updateCmd struct {
	Scope  source.Scope `arg:"" default:"all" help:"render skills, agents, and docs, or one of them"`
	Source string       `name:"source" env:"ZPECS_SOURCE" default:"${default_source}" help:"read definitions from a local directory, or clone it if it is a git repository"`
	Target string       `name:"target" enum:"claude,opencode" default:"opencode" help:"render for claude or opencode"`
}

func (u *updateCmd) Run() error {
	return update.Run(update.Options{Scope: u.Scope, Source: u.Source, Target: u.Target})
}

// reviewCmd is the kong grammar for `zpecs review`.
type reviewCmd struct {
	Scope   opencode.Scope `arg:"" help:"code, architecture, or prose"`
	Model   *string        `name:"model" help:"model opencode uses. Defaults to ${default_model}."`
	Variant *string        `name:"variant" help:"variant opencode uses. Defaults to ${default_variant} when no model or variant is given."`
}

// options resolves the parsed flags into review options.
func (r *reviewCmd) options() review.Options {
	model := opencode.DefaultModel
	if r.Model != nil {
		model = *r.Model
	}
	variant := ""
	if r.Variant != nil {
		variant = *r.Variant
	}
	if r.Model == nil && r.Variant == nil {
		variant = opencode.DefaultVariant
	}
	return review.Options{Scope: r.Scope, Model: model, Variant: variant}
}

func (r *reviewCmd) Run() error {
	return review.Run(r.options())
}

// convertCmd is the kong grammar for `zpecs convert`.
type convertCmd struct {
	Path string `arg:"" help:"path to a spec markdown file"`
}

func (c *convertCmd) Run() error {
	return spec.Convert(c.Path, os.Stdout)
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
