package report

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

// Out receives summary lines.
var Out io.Writer = os.Stdout

// Summary prints the line reporting an update run. When targetName
// is the docs target, the line omits it.
func Summary(kinds []source.Kind, targetName, sourceLabel string, n int) error {
	label := scopeLabel(kinds)
	var err error
	if targetName == target.Docs {
		_, err = fmt.Fprintf(Out, "updating %s from %s (%d files)\n", label, sourceLabel, n)
		return err
	}
	_, err = fmt.Fprintf(Out, "updating %s for %s from %s (%d definitions)\n", label, targetName, sourceLabel, n)
	return err
}

// scopeLabel names the kinds a run selects, e.g. "skills and agents".
func scopeLabel(kinds []source.Kind) string {
	words := make([]string, len(kinds))
	for i, kind := range kinds {
		words[i] = kindPlurals[kind]
	}
	return strings.Join(words, " and ")
}

var kindPlurals = map[source.Kind]string{
	source.Skill: "skills",
	source.Agent: "agents",
	source.Doc:   "docs",
}
