package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeSpec(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestReadParsesTitlePurposeRequirementsAndScenarios(t *testing.T) {
	path := writeSpec(t, "# Convert\n\n## Purpose\nTurn a spec markdown file into a JSON document.\n\n### Requirement: Command Form\nThe system SHALL accept a path to a spec file as its only argument.\n\n#### Scenario: Path argument\n- GIVEN the path `specs/cli/sync.md`\n- WHEN the command runs\n- THEN it reads the spec at that path\n\n### Requirement: Output\nThe system SHALL print one JSON object to stdout.\n\n#### Scenario: Whole spec\n- GIVEN a spec file with a title, a purpose, and one requirement\n- WHEN the command runs\n- THEN it prints one JSON object\n- AND the object carries the title, purpose, and requirements\n")
	want := Document{
		Title:   "Convert",
		Purpose: "Turn a spec markdown file into a JSON document.",
		Requirements: []Requirement{
			{
				Name: "Command Form",
				Body: "The system SHALL accept a path to a spec file as its only argument.",
				Scenarios: []Scenario{
					{
						Name: "Path argument",
						Steps: []string{
							"GIVEN the path `specs/cli/sync.md`",
							"WHEN the command runs",
							"THEN it reads the spec at that path",
						},
					},
				},
			},
			{
				Name: "Output",
				Body: "The system SHALL print one JSON object to stdout.",
				Scenarios: []Scenario{
					{
						Name: "Whole spec",
						Steps: []string{
							"GIVEN a spec file with a title, a purpose, and one requirement",
							"WHEN the command runs",
							"THEN it prints one JSON object",
							"AND the object carries the title, purpose, and requirements",
						},
					},
				},
			},
		},
	}

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRequirementWithoutScenariosHasEmptyNonNilSlice(t *testing.T) {
	path := writeSpec(t, "# Title\n\n### Requirement: One\nBody.\n")
	got, err := read(path)
	require.NoError(t, err)
	req := got.Requirements[0]
	require.NotNil(t, req.Scenarios)
	require.Len(t, req.Scenarios, 0)
}

func TestStepLineKeepsInlineCodeBackticks(t *testing.T) {
	path := writeSpec(t, "# Convert\n\n### Requirement: Command Form\nBody.\n\n#### Scenario: Path argument\n- GIVEN the path `specs/cli/sync.md`\n")
	got, err := read(path)
	require.NoError(t, err)
	steps := got.Requirements[0].Scenarios[0].Steps
	require.Len(t, steps, 1)
	require.Equal(t, "GIVEN the path `specs/cli/sync.md`", steps[0])
}

func TestReadErrorsOnMalformedContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "no top-level heading",
			content: "## Purpose\nNothing.\n",
			wantErr: "no top-level heading",
		},
		{
			name:    "requirement without a name",
			content: "# Title\n\n### Requirement:\nBody.\n",
			wantErr: "requirement heading without a name",
		},
		{
			name:    "scenario outside a requirement",
			content: "# Title\n\n#### Scenario: Lost\n- GIVEN nothing\n",
			wantErr: "scenario outside a requirement",
		},
		{
			name:    "scenario after an unrelated heading",
			content: "# Title\n\n### Requirement: A\nBody.\n\n## Notes\n#### Scenario: Lost\n- GIVEN nothing\n",
			wantErr: "scenario outside a requirement",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeSpec(t, tt.content)
			_, err := read(path)
			require.Error(t, err)
			require.Equal(t, tt.wantErr, err.Error())
		})
	}
}

func TestReadErrorsOnMissingFile(t *testing.T) {
	_, err := read(filepath.Join(t.TempDir(), "missing.md"))
	require.Error(t, err)
}

func TestWriteRoundTripsTheDocument(t *testing.T) {
	doc := Document{
		Title:   "Convert",
		Purpose: "Turn a spec file into JSON.",
		Requirements: []Requirement{
			{
				Name:      "Command Form",
				Body:      "The system accepts a path.",
				Scenarios: []Scenario{},
			},
			{
				Name: "Scenarios",
				Body: "Each scenario keeps its steps.",
				Scenarios: []Scenario{
					{
						Name:  "Steps in order",
						Steps: []string{"GIVEN a scenario", "THEN it prints steps"},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, write(&buf, doc))
	var got Document
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Equal(t, doc, got)
}

func TestWritePrintsEmptyArrays(t *testing.T) {
	doc := Document{
		Title: "Convert",
		Requirements: []Requirement{
			{
				Name:      "One",
				Scenarios: []Scenario{},
			},
			{
				Name: "Two",
				Scenarios: []Scenario{
					{
						Name:  "Empty",
						Steps: []string{},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, write(&buf, doc))
	for _, want := range []string{`"scenarios": []`, `"steps": []`} {
		require.Contains(t, buf.String(), want)
	}
}

func TestConvertWritesTheSpecAsJSON(t *testing.T) {
	path := writeSpec(t, "# Convert\n\n## Purpose\nTurn a spec markdown file into a JSON document.\n\n### Requirement: Command Form\nThe system SHALL accept a path to a spec file as its only argument.\n")
	var buf bytes.Buffer
	require.NoError(t, Convert(path, &buf))
	var got Document
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got), "convert output is not one JSON object\n%s", buf.String())
	require.Equal(t, "Convert", got.Title)
	require.Equal(t, "Turn a spec markdown file into a JSON document.", got.Purpose)
	require.Len(t, got.Requirements, 1)
	require.Equal(t, "Command Form", got.Requirements[0].Name)
}

func TestConvertErrorsOnMalformedContentAndWritesNothing(t *testing.T) {
	path := writeSpec(t, "## Purpose\nNothing.\n")
	var buf bytes.Buffer
	err := Convert(path, &buf)
	require.Error(t, err)
	require.Equal(t, "no top-level heading", err.Error())
	require.Empty(t, buf.Bytes())
}

func TestConvertErrorsOnMissingFileAndWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	err := Convert(filepath.Join(t.TempDir(), "missing.md"), &buf)
	require.Error(t, err)
	require.Empty(t, buf.Bytes())
}
