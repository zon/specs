package update

import (
	"fmt"
	"os"

	"github.com/zon/specs/internal/clone"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/repo"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/targetdir"
)

// Options selects what an update run renders: the scope of kinds to
// read, the source they come from, and the target to write to.
type Options struct {
	Scope  source.Scope
	Source string
	Target string
}

// Run renders the selected kinds from the source into the target.
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

// pair is one run of the sync pipeline: the name to report, the target
// to write to, and the kinds it selects.
type pair struct {
	name   string
	target string
	kinds  []source.Kind
}

// pairs selects the runs for a scope. Skills and agents write to the
// named target. Docs always write to docs/zpecs.
func pairs(s source.Scope, targetName string) []pair {
	switch s {
	case source.ScopeSkills:
		return []pair{{name: "skills", target: targetName, kinds: []source.Kind{source.Skill}}}
	case source.ScopeAgents:
		return []pair{{name: "agents", target: targetName, kinds: []source.Kind{source.Agent}}}
	case source.ScopeDocs:
		return []pair{{name: "docs", target: target.Docs, kinds: []source.Kind{source.Doc}}}
	default:
		return []pair{
			{name: "skills and agents", target: targetName, kinds: []source.Kind{source.Skill, source.Agent}},
			{name: "docs", target: target.Docs, kinds: []source.Kind{source.Doc}},
		}
	}
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
// label to report, and a cleanup func. A value with a scheme is a
// repository to clone. Anything else is a local directory to read in
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
