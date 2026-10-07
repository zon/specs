package update

import (
	"strings"

	"github.com/zon/specs/internal/gitops"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

// Options selects what an update run renders: the scope of kinds to
// read, the source they come from, the target to write to, and whether
// a full update also renders agents and enables optional features.
type Options struct {
	Scope         source.Scope
	Agents        bool
	Orchestration bool
	Process       bool
	Source        string
	Target        string
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
	features := render.Enabled(opts.Orchestration, opts.Process)
	for _, p := range pairs(opts.Scope, opts.Target, opts.Agents) {
		if err := updatePair(root, sourceDir, sourceLabel, p, features); err != nil {
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
// named target. Docs write to docs/zpecs. A full update only renders
// agents when agents is true; the agents scope always renders them.
func pairs(s source.Scope, targetName string, agents bool) []pair {
	switch s {
	case source.ScopeSkills:
		return []pair{{target: targetName, kinds: []source.Kind{source.Skill}}}
	case source.ScopeAgents:
		return []pair{{target: targetName, kinds: []source.Kind{source.Agent}}}
	case source.ScopeDocs:
		return []pair{{target: source.Docs, kinds: []source.Kind{source.Doc}}}
	}
	targetKinds := []source.Kind{source.Skill}
	if agents {
		targetKinds = append(targetKinds, source.Agent)
	}
	return []pair{
		{target: targetName, kinds: targetKinds},
		{target: source.Docs, kinds: []source.Kind{source.Doc}},
	}
}

// updatePair renders a pair's definitions into its target under root.
// It then reports the run.
func updatePair(root, sourceDir, sourceLabel string, p pair, features render.Features) error {
	defs, err := source.ReadKinds(p.kinds, sourceDir)
	if err != nil {
		return err
	}
	owned, err := targetdir.Owned(root, p.target)
	if err != nil {
		return err
	}
	defs, texts, err := renderDefs(defs, render.ForTarget(p.target, features))
	if err != nil {
		return err
	}
	if _, err := targetdir.RemoveStale(root, p.target, owned, defs, p.kinds...); err != nil {
		return err
	}
	content := func(d source.Definition) (string, error) { return texts[d], nil }
	if err := targetdir.WriteAll(root, p.target, defs, content, owned); err != nil {
		return err
	}
	if err := targetdir.SaveOwned(root, p.target, owned); err != nil {
		return err
	}
	return report.Summary(p.kinds, p.target, sourceLabel, len(defs))
}

// renderDefs renders each definition with content and returns the ones
// whose rendered text is not blank, keyed by definition. A definition
// that renders to nothing is dropped, so the run neither writes it nor
// records it as owned, and a file it wrote before is removed as stale.
func renderDefs(defs []source.Definition, content func(source.Definition) (string, error)) ([]source.Definition, map[source.Definition]string, error) {
	texts := make(map[source.Definition]string, len(defs))
	kept := make([]source.Definition, 0, len(defs))
	for _, d := range defs {
		text, err := content(d)
		if err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		texts[d] = text
		kept = append(kept, d)
	}
	return kept, texts, nil
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
