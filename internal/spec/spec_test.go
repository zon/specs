package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadParsesTitlePurposeRequirementsAndScenarios(t *testing.T) {
	path := write(t, "# Convert\n\n## Purpose\nTurn a spec markdown file into a JSON document.\n\n### Requirement: Command Form\nThe system SHALL accept a path to a spec file as its only argument.\n\n#### Scenario: Path argument\n- GIVEN the path `specs/cli/sync.md`\n- WHEN the command runs\n- THEN it reads the spec at that path\n\n### Requirement: Output\nThe system SHALL print one JSON object to stdout.\n\n#### Scenario: Whole spec\n- GIVEN a spec file with a title, a purpose, and one requirement\n- WHEN the command runs\n- THEN it prints one JSON object\n- AND the object carries the title, purpose, and requirements\n")
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

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Read = %+v, want %+v", got, want)
	}
}

func TestRequirementWithoutScenariosHasEmptyNonNilSlice(t *testing.T) {
	path := write(t, "# Title\n\n### Requirement: One\nBody.\n")
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	req := got.Requirements[0]
	if req.Scenarios == nil {
		t.Fatal("Scenarios = nil, want an empty slice")
	}
	if len(req.Scenarios) != 0 {
		t.Fatalf("len(Scenarios) = %d, want 0", len(req.Scenarios))
	}
}

func TestStepLineKeepsInlineCodeBackticks(t *testing.T) {
	path := write(t, "# Convert\n\n### Requirement: Command Form\nBody.\n\n#### Scenario: Path argument\n- GIVEN the path `specs/cli/sync.md`\n")
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	steps := got.Requirements[0].Scenarios[0].Steps
	if len(steps) != 1 {
		t.Fatalf("len(steps) = %d, want 1", len(steps))
	}
	if want := "GIVEN the path `specs/cli/sync.md`"; steps[0] != want {
		t.Fatalf("step = %q, want %q", steps[0], want)
	}
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
			path := write(t, tt.content)
			if _, err := Read(path); err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Read error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestReadErrorsOnMissingFile(t *testing.T) {
	if _, err := Read(filepath.Join(t.TempDir(), "missing.md")); err == nil {
		t.Fatal("expected an error for a missing file")
	}
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
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write: %v", err)
	}
	var got Document
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, doc) {
		t.Fatalf("round trip = %+v, want %+v", got, doc)
	}
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
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write: %v", err)
	}
	for _, want := range []string{`"scenarios": []`, `"steps": []`} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("output = %s, want %s", buf.String(), want)
		}
	}
}
