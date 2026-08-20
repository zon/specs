package report

import (
	"fmt"
	"io"
	"os"

	targetpkg "github.com/zon/specs/internal/target"
)

// sink is where Summary prints. The caller does not choose the output.
// report does.
var sink io.Writer = os.Stdout

// Summary prints the line reporting an update run. The scope word labels
// the run. "all" reads as "skills and agents". When target is the docs
// target, the line omits it.
func Summary(scope, target, sourceLabel string, n int) {
	scopeLabel := scope
	if scope == "all" {
		scopeLabel = "skills and agents"
	}
	if target == targetpkg.Docs {
		fmt.Fprintf(sink, "updating %s from %s (%d files)\n", scopeLabel, sourceLabel, n)
		return
	}
	fmt.Fprintf(sink, "updating %s for %s from %s (%d definitions)\n", scopeLabel, target, sourceLabel, n)
}
