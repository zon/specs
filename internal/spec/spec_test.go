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

func TestToGenericJSONProducesANodeTree(t *testing.T) {
	src := []byte("# Title\n\n## Purpose\nProse `code` here.\n\n- step one\n")
	root, err := toGenericJSON(src)
	if err != nil {
		t.Fatalf("toGenericJSON: %v", err)
	}
	if root.Type != "document" {
		t.Fatalf("root type = %q, want %q", root.Type, "document")
	}

	title := root.Children[0]
	if title.Type != "heading" || title.Level != 1 {
		t.Fatalf("first child = %+v, want a level 1 heading", title)
	}
	if title.Children[0].Text != "Title" {
		t.Fatalf("title text = %q, want %q", title.Children[0].Text, "Title")
	}

	purpose := root.Children[1]
	if purpose.Type != "heading" || purpose.Level != 2 {
		t.Fatalf("second child = %+v, want a level 2 heading", purpose)
	}
	if purpose.Children[0].Text != "Purpose" {
		t.Fatalf("purpose text = %q, want %q", purpose.Children[0].Text, "Purpose")
	}

	paragraph := root.Children[2]
	if paragraph.Type != "paragraph" {
		t.Fatalf("third child type = %q, want %q", paragraph.Type, "paragraph")
	}
	if paragraph.Children[0].Text != "Prose " {
		t.Fatalf("paragraph text = %q, want %q", paragraph.Children[0].Text, "Prose ")
	}
	if code := paragraph.Children[1]; code.Type != "code" || code.Text != "`code`" {
		t.Fatalf("paragraph child = %+v, want a code node with text %q", code, "`code`")
	}
	if paragraph.Children[2].Text != " here." {
		t.Fatalf("paragraph text = %q, want %q", paragraph.Children[2].Text, " here.")
	}

	list := root.Children[3]
	if list.Type != "list" {
		t.Fatalf("fourth child type = %q, want %q", list.Type, "list")
	}
	item := list.Children[0]
	if item.Type != "list_item" {
		t.Fatalf("list child type = %q, want %q", item.Type, "list_item")
	}
	block := item.Children[0]
	if block.Type != "text_block" {
		t.Fatalf("list item child type = %q, want %q", block.Type, "text_block")
	}
	if block.Children[0].Text != "step one" {
		t.Fatalf("list item text = %q, want %q", block.Children[0].Text, "step one")
	}
}

func TestFromGenericJSONMapsTheTree(t *testing.T) {
	root := genericNode{Type: "document", Children: []genericNode{
		{Type: "heading", Level: 1, Children: []genericNode{{Type: "text", Text: "Convert"}}},
		{Type: "heading", Level: 2, Children: []genericNode{{Type: "text", Text: "Purpose"}}},
		{Type: "paragraph", Children: []genericNode{{Type: "text", Text: "Turn a file into JSON."}}},
		{Type: "heading", Level: 3, Children: []genericNode{{Type: "text", Text: "Requirement: Command Form"}}},
		{Type: "paragraph", Children: []genericNode{{Type: "text", Text: "The system accepts a path."}}},
		{Type: "heading", Level: 4, Children: []genericNode{{Type: "text", Text: "Scenario: Path argument"}}},
		{Type: "list", Children: []genericNode{
			{Type: "list_item", Children: []genericNode{
				{Type: "text_block", Children: []genericNode{{Type: "text", Text: "GIVEN a path"}}},
			}},
		}},
	}}
	doc, err := fromGenericJSON(root)
	if err != nil {
		t.Fatalf("fromGenericJSON: %v", err)
	}
	if doc.Title != "Convert" {
		t.Fatalf("Title = %q, want %q", doc.Title, "Convert")
	}
	if doc.Purpose != "Turn a file into JSON." {
		t.Fatalf("Purpose = %q, want %q", doc.Purpose, "Turn a file into JSON.")
	}
	req := doc.Requirements[0]
	if req.Name != "Command Form" || req.Body != "The system accepts a path." {
		t.Fatalf("requirement = %+v, want name %q with body %q", req, "Command Form", "The system accepts a path.")
	}
	scenario := req.Scenarios[0]
	if scenario.Name != "Path argument" {
		t.Fatalf("scenario name = %q, want %q", scenario.Name, "Path argument")
	}
	if want := []string{"GIVEN a path"}; !reflect.DeepEqual(scenario.Steps, want) {
		t.Fatalf("steps = %v, want %v", scenario.Steps, want)
	}
}
