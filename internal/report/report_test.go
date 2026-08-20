package report

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSummary(t *testing.T) {
	old := sink
	var out strings.Builder
	sink = &out
	t.Cleanup(func() { sink = old })

	cases := []struct {
		name   string
		scope  string
		target string
		source string
		n      int
		want   string
	}{
		{name: "skills with target", scope: "skills", target: "claude", source: "/tmp/src", n: 2, want: "updating skills for claude from /tmp/src (2 definitions)\n"},
		{name: "docs write target", scope: "docs", target: "docs", source: "/tmp/src", n: 3, want: "updating docs from /tmp/src (3 files)\n"},
		{name: "all maps to skills and agents", scope: "all", target: "opencode", source: "/tmp/src", n: 4, want: "updating skills and agents for opencode from /tmp/src (4 definitions)\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out.Reset()
			Summary(tc.scope, tc.target, tc.source, tc.n)
			require.Equal(t, tc.want, out.String())
		})
	}
}
