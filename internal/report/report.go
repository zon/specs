package report

import (
	"fmt"
	"io"
)

// Summary prints the line that reports an update run to w. When target
// is empty the line omits it (docs have no target).
func Summary(w io.Writer, scopeName, target, sourceLabel string, n int) {
	if target == "" {
		fmt.Fprintf(w, "updating %s from %s (%d files)\n", scopeName, sourceLabel, n)
		return
	}
	fmt.Fprintf(w, "updating %s for %s from %s (%d definitions)\n", scopeName, target, sourceLabel, n)
}
