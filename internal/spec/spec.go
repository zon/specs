// Package spec parses and writes the spec markdown format.
package spec

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Document is one parsed spec file.
type Document struct {
	Title        string        `json:"title"`
	Purpose      string        `json:"purpose"`
	Requirements []Requirement `json:"requirements"`
}

// Requirement is one `### Requirement:` section.
type Requirement struct {
	Name      string     `json:"name"`
	Body      string     `json:"body"`
	Scenarios []Scenario `json:"scenarios"`
}

// Scenario is one `#### Scenario:` section.
type Scenario struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

// Read parses the spec file at path.
func Read(path string) (Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	root := goldmark.New().Parser().Parse(text.NewReader(content))
	return fromAST(root, content)
}

// Write prints doc as one indented JSON object.
func Write(w io.Writer, doc Document) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

// fromAST walks the markdown AST into a Document. Headings drive the
// state; prose and lists fill the current section.
func fromAST(root ast.Node, src []byte) (Document, error) {
	doc := Document{Requirements: []Requirement{}}
	current := -1
	inPurpose := false
	titleSet := false
	for child := root.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != ast.KindHeading {
			switch {
			case inPurpose:
				doc.Purpose = joinProse(doc.Purpose, proseBlock(child, src))
			case current >= 0 && len(doc.Requirements[current].Scenarios) > 0:
				if child.Kind() == ast.KindList {
					scenario := &doc.Requirements[current].Scenarios[len(doc.Requirements[current].Scenarios)-1]
					scenario.Steps = append(scenario.Steps, stepsBlock(child, src)...)
				}
			case current >= 0:
				doc.Requirements[current].Body = joinProse(doc.Requirements[current].Body, proseBlock(child, src))
			}
			continue
		}
		level := child.(*ast.Heading).Level
		text := renderNode(child, src)
		switch {
		case level == 1 && !titleSet:
			doc.Title = text
			titleSet = true
			current = -1
			inPurpose = false
		case level == 2 && text == "Purpose":
			doc.Purpose = ""
			current = -1
			inPurpose = true
		case level == 3 && strings.HasPrefix(text, "Requirement:"):
			name := strings.TrimSpace(strings.TrimPrefix(text, "Requirement:"))
			if name == "" {
				return Document{}, errors.New("requirement heading without a name")
			}
			doc.Requirements = append(doc.Requirements, Requirement{Name: name, Scenarios: []Scenario{}})
			current = len(doc.Requirements) - 1
			inPurpose = false
		case level == 4 && strings.HasPrefix(text, "Scenario:"):
			if current < 0 {
				return Document{}, errors.New("scenario outside a requirement")
			}
			name := strings.TrimSpace(strings.TrimPrefix(text, "Scenario:"))
			doc.Requirements[current].Scenarios = append(doc.Requirements[current].Scenarios, Scenario{Name: name, Steps: []string{}})
			inPurpose = false
		default:
			current = -1
			inPurpose = false
		}
	}
	if !titleSet {
		return Document{}, errors.New("no top-level heading")
	}
	doc.Purpose = strings.TrimSpace(doc.Purpose)
	for i := range doc.Requirements {
		doc.Requirements[i].Body = strings.TrimSpace(doc.Requirements[i].Body)
	}
	return doc, nil
}

// proseBlock renders one prose block. Only paragraphs carry prose.
func proseBlock(n ast.Node, src []byte) string {
	if n.Kind() != ast.KindParagraph {
		return ""
	}
	return strings.TrimSpace(renderNode(n, src))
}

// joinProse joins a prose block onto an existing section with a blank line.
func joinProse(prose, block string) string {
	if block == "" {
		return prose
	}
	if prose == "" {
		return block
	}
	return prose + "\n\n" + block
}

// stepsBlock renders a list's items as plain step lines. The list marker
// is markdown syntax and does not belong in the JSON output.
func stepsBlock(n ast.Node, src []byte) []string {
	steps := []string{}
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		if item.Kind() != ast.KindListItem {
			continue
		}
		line := strings.TrimSpace(renderNode(item, src))
		if line != "" {
			steps = append(steps, line)
		}
	}
	return steps
}

// renderNode renders an AST node as text. Containers concatenate their
// children. Strong, emph, and link drop their markers. Code spans get
// wrapped in backticks.
func renderNode(n ast.Node, src []byte) string {
	switch n.Kind() {
	case ast.KindText:
		t := n.(*ast.Text)
		value := string(t.Value(src))
		if t.SoftLineBreak() || t.HardLineBreak() {
			value += "\n"
		}
		return value
	case ast.KindCodeSpan:
		// CodeSpan carries no opening and closing marker lengths, so
		// single backticks are the closest faithful rendering.
		return "`" + string(n.Text(src)) + "`"
	case ast.KindDocument, ast.KindHeading, ast.KindParagraph, ast.KindTextBlock,
		ast.KindList, ast.KindListItem, ast.KindEmphasis, ast.KindLink:
		return joinChildren(n, src)
	default:
		if n.HasChildren() {
			return joinChildren(n, src)
		}
		if n.Type() == ast.TypeInline {
			return strings.TrimRight(string(n.Text(src)), "\n")
		}
		return ""
	}
}

// joinChildren renders an AST node's children in order.
func joinChildren(n ast.Node, src []byte) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		sb.WriteString(renderNode(c, src))
	}
	return sb.String()
}
