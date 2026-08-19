package report

import (
	"fmt"
	"io"

	targetpkg "github.com/zon/specs/internal/target"
)

// Summary prints the line reporting an update run to w. When target
// is empty or is the docs target, the line omits it.
func Summary(w io.Writer, scopeName, target, sourceLabel string, n int) {
	if target == "" || target == targetpkg.Docs {
		fmt.Fprintf(w, "updating %s from %s (%d files)\n", scopeName, sourceLabel, n)
		return
	}
	fmt.Fprintf(w, "updating %s for %s from %s (%d definitions)\n", scopeName, target, sourceLabel, n)
}
