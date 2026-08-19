package report

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSummary(t *testing.T) {
	cases := []struct {
		name      string
		scopeName string
		target    string
		source    string
		n         int
		want      string
	}{
		{name: "skills with target", scopeName: "skills", target: "claude", source: "/tmp/src", n: 2, want: "updating skills for claude from /tmp/src (2 definitions)\n"},
		{name: "docs without target", scopeName: "docs", source: "/tmp/src", n: 3, want: "updating docs from /tmp/src (3 files)\n"},
		{name: "docs write target", scopeName: "docs", target: "docs", source: "/tmp/src", n: 3, want: "updating docs from /tmp/src (3 files)\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			Summary(&out, tc.scopeName, tc.target, tc.source, tc.n)
			require.Equal(t, tc.want, out.String())
		})
	}
}
