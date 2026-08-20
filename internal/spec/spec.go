// Package spec parses a spec markdown file into a JSON document.
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

// read parses the spec file at path.
func read(path string) (Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	root := goldmark.New().Parser().Parse(text.NewReader(content))
	return fromAST(root, content)
}

// Convert reads the spec file at path and writes it to w as JSON.
func Convert(path string, w io.Writer) error {
	doc, err := read(path)
	if err != nil {
		return err
	}
	return write(w, doc)
}

// write prints doc as one indented JSON object.
func write(w io.Writer, doc Document) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

// fromAST walks the markdown AST into a Document. Headings drive the
// state. Prose and lists fill the current section.
func fromAST(root ast.Node, src []byte) (Document, error) {
	st := fromASTState{doc: Document{Requirements: []Requirement{}}, current: -1}
	for child := root.FirstChild(); child != nil; child = child.NextSibling() {
		var err error
		if child.Kind() == ast.KindHeading {
			err = st.applyHeading(child, src)
		} else {
			st.applyContent(child, src)
		}
		if err != nil {
			return Document{}, err
		}
	}
	if !st.titleSet {
		return Document{}, errors.New("no top-level heading")
	}
	st.doc.Purpose = strings.TrimSpace(st.doc.Purpose)
	for i := range st.doc.Requirements {
		st.doc.Requirements[i].Body = strings.TrimSpace(st.doc.Requirements[i].Body)
	}
	return st.doc, nil
}

// fromASTState is the walk state: the document under construction and
// where the next content block lands.
type fromASTState struct {
	doc       Document
	current   int
	inPurpose bool
	titleSet  bool
}

// applyContent adds one non-heading node to the purpose, the current
// scenario's steps, or the current requirement's body.
func (st *fromASTState) applyContent(child ast.Node, src []byte) {
	switch {
	case st.inPurpose:
		st.doc.Purpose = joinProse(st.doc.Purpose, proseBlock(child, src))
	case st.current >= 0 && len(st.doc.Requirements[st.current].Scenarios) > 0:
		if child.Kind() == ast.KindList {
			scenario := &st.doc.Requirements[st.current].Scenarios[len(st.doc.Requirements[st.current].Scenarios)-1]
			scenario.Steps = append(scenario.Steps, stepsBlock(child, src)...)
		}
	case st.current >= 0:
		st.doc.Requirements[st.current].Body = joinProse(st.doc.Requirements[st.current].Body, proseBlock(child, src))
	}
}

// applyHeading moves the state to the section the heading opens. A
// heading that matches none of the known shapes leaves the current
// section behind.
func (st *fromASTState) applyHeading(heading ast.Node, src []byte) error {
	level := heading.(*ast.Heading).Level
	text := renderNode(heading, src)
	switch {
	case level == 1 && !st.titleSet:
		st.doc.Title = text
		st.titleSet = true
		st.current = -1
		st.inPurpose = false
	case level == 2 && text == "Purpose":
		st.doc.Purpose = ""
		st.current = -1
		st.inPurpose = true
	case level == 3 && strings.HasPrefix(text, "Requirement:"):
		name := strings.TrimSpace(strings.TrimPrefix(text, "Requirement:"))
		if name == "" {
			return errors.New("requirement heading without a name")
		}
		st.doc.Requirements = append(st.doc.Requirements, Requirement{Name: name, Scenarios: []Scenario{}})
		st.current = len(st.doc.Requirements) - 1
		st.inPurpose = false
	case level == 4 && strings.HasPrefix(text, "Scenario:"):
		if st.current < 0 {
			return errors.New("scenario outside a requirement")
		}
		name := strings.TrimSpace(strings.TrimPrefix(text, "Scenario:"))
		st.doc.Requirements[st.current].Scenarios = append(st.doc.Requirements[st.current].Scenarios, Scenario{Name: name, Steps: []string{}})
		st.inPurpose = false
	default:
		st.current = -1
		st.inPurpose = false
	}
	return nil
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
// children. Strong, emph, and link drop their markers. Wrap code spans
// in backticks.
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
