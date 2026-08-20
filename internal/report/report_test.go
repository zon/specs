package report_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/testutil"
)

func TestSummary(t *testing.T) {
	cases := []struct {
		name   string
		kinds  []source.Kind
		target string
		source string
		n      int
		want   string
	}{
		{name: "skills with target", kinds: []source.Kind{source.Skill}, target: target.Claude, source: "/tmp/src", n: 2, want: "updating skills for claude from /tmp/src (2 definitions)\n"},
		{name: "agents with target", kinds: []source.Kind{source.Agent}, target: target.Opencode, source: "/tmp/src", n: 2, want: "updating agents for opencode from /tmp/src (2 definitions)\n"},
		{name: "skills and agents with target", kinds: []source.Kind{source.Skill, source.Agent}, target: target.Claude, source: "/tmp/src", n: 4, want: "updating skills and agents for claude from /tmp/src (4 definitions)\n"},
		{name: "docs write target", kinds: []source.Kind{source.Doc}, target: target.Docs, source: "/tmp/src", n: 3, want: "updating docs from /tmp/src (3 files)\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reported := testutil.CaptureReport(t)
			err := report.Summary(tc.kinds, tc.target, tc.source, tc.n)
			require.NoError(t, err)
			require.Equal(t, tc.want, reported())
		})
	}
}

// failWriter fails every write.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestSummaryWriteError(t *testing.T) {
	prev := report.Out
	report.Out = failWriter{}
	t.Cleanup(func() { report.Out = prev })

	err := report.Summary([]source.Kind{source.Skill}, target.Claude, "/tmp/src", 2)
	require.Error(t, err)
	require.ErrorContains(t, err, "write failed")
}
