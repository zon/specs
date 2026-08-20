// Package update orchestrates an update run's read, write, and report steps.
package update

import (
	"github.com/zon/specs/internal/clone"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/repo"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/targetdir"
)

// Options is the parsed input to an update run.
type Options struct {
	// Scope is the scope word, one of "all", "skills", "agents", or
	// "docs". "all" selects skills, agents, and docs.
	Scope string
	// Source is where the definitions come from.
	Source string
	// Target is the runner the skills and agents write to.
	Target string
}

// pair is one run of the sync pipeline: the scope word to report, the
// target to write to, and the kinds it selects.
type pair struct {
	scope  string
	target string
	kinds  []source.Kind
}

// pairs selects the runs for a scope. Skills and agents write to the
// runner target. Docs always write to docs/zpecs.
func pairs(scope, runner string) []pair {
	switch scope {
	case "skills":
		return []pair{{scope: "skills", target: runner, kinds: []source.Kind{source.Skill}}}
	case "agents":
		return []pair{{scope: "agents", target: runner, kinds: []source.Kind{source.Agent}}}
	case "docs":
		return []pair{{scope: "docs", target: target.Docs, kinds: []source.Kind{source.Doc}}}
	default:
		return []pair{
			{scope: "all", target: runner, kinds: []source.Kind{source.Skill, source.Agent}},
			{scope: "docs", target: target.Docs, kinds: []source.Kind{source.Doc}},
		}
	}
}

// Run reads the definitions for opts.Scope from opts.Source and writes
// them into the repository under opts.Target.
func Run(opts Options) error {
	root, err := repo.Root(".")
	if err != nil {
		return err
	}
	sourceDir, sourceLabel, cleanup, err := resolveSource(opts.Source)
	if err != nil {
		return err
	}
	defer cleanup()
	for _, p := range pairs(opts.Scope, opts.Target) {
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
		return err
	}
	if err := targetdir.WriteAll(root, p.target, defs, func(d source.Definition) (string, error) {
		return render.Definition(d, p.target)
	}, owned); err != nil {
		return err
	}
	if err := targetdir.SaveOwned(root, p.target, owned); err != nil {
		return err
	}
	report.Summary(p.scope, p.target, sourceLabel, len(defs))
	return nil
}

// resolveSource returns the directory the definitions come from, the
// label to report, and a cleanup func. A value with a scheme is a
// repository to clone. Anything else is a local directory read in
// place.
func resolveSource(source string) (dir, label string, cleanup func(), err error) {
	if clone.IsRemote(source) {
		dir, cleanup, err = clone.Clone(source)
		if err != nil {
			return "", "", nil, err
		}
		return dir, source, cleanup, nil
	}
	return source, source, func() {}, nil
}
