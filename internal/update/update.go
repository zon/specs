package update

import (
	"github.com/zon/specs/internal/gitops"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
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
	root, err := gitops.Root()
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

// pair is one update run: the target to write to and the kinds it
// selects.
type pair struct {
	target string
	kinds  []source.Kind
}

// pairs selects the runs for a scope. Skills and agents write to the
// named target. Docs write to docs/zpecs.
func pairs(s source.Scope, targetName string) []pair {
	switch s {
	case source.ScopeSkills:
		return []pair{{target: targetName, kinds: []source.Kind{source.Skill}}}
	case source.ScopeAgents:
		return []pair{{target: targetName, kinds: []source.Kind{source.Agent}}}
	case source.ScopeDocs:
		return []pair{{target: source.Docs, kinds: []source.Kind{source.Doc}}}
	default:
		return []pair{
			{target: targetName, kinds: []source.Kind{source.Skill, source.Agent}},
			{target: source.Docs, kinds: []source.Kind{source.Doc}},
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
		return err
	}
	if err := targetdir.WriteAll(root, p.target, defs, render.ForTarget(p.target), owned); err != nil {
		return err
	}
	if err := targetdir.SaveOwned(root, p.target, owned); err != nil {
		return err
	}
	return report.Summary(p.kinds, p.target, sourceLabel, len(defs))
}

// resolveSource returns the directory the definitions come from, the
// label to report, and a cleanup func. A value with a scheme is a
// repository to clone. Anything else is a local directory to read in
// place.
func resolveSource(source string) (dir, label string, cleanup func(), err error) {
	if gitops.IsRemote(source) {
		dir, cleanup, err = gitops.Clone(source)
		if err != nil {
			return "", "", nil, err
		}
		return dir, source, cleanup, nil
	}
	return source, source, func() {}, nil
}
