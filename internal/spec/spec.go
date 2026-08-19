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
	generic, err := toGenericJSON(content)
	if err != nil {
		return Document{}, err
	}
	return fromGenericJSON(generic)
}

// toGenericJSON parses markdown with goldmark into a generic JSON tree.
func toGenericJSON(src []byte) (genericNode, error) {
	md := goldmark.New()
	root := md.Parser().Parse(text.NewReader(src))
	return convertNode(root, src), nil
}

// Write prints doc as one indented JSON object.
func Write(w io.Writer, doc Document) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

// genericNode is one node of the generic JSON tree. Containers carry
// children, headings carry a level, and leaves carry text.
type genericNode struct {
	Type     string        `json:"type"`
	Level    int           `json:"level,omitempty"`
	Children []genericNode `json:"children,omitempty"`
	Text     string        `json:"text,omitempty"`
}

// convertNode maps one AST node to the generic JSON shape.
func convertNode(n ast.Node, src []byte) genericNode {
	switch n.Kind() {
	case ast.KindDocument:
		return genericNode{Type: "document", Children: convertChildren(n, src)}
	case ast.KindHeading:
		return genericNode{Type: "heading", Level: n.(*ast.Heading).Level, Children: convertChildren(n, src)}
	case ast.KindParagraph:
		return genericNode{Type: "paragraph", Children: convertChildren(n, src)}
	case ast.KindTextBlock:
		return genericNode{Type: "text_block", Children: convertChildren(n, src)}
	case ast.KindText:
		t := n.(*ast.Text)
		value := string(t.Value(src))
		if t.SoftLineBreak() || t.HardLineBreak() {
			value += "\n"
		}
		return genericNode{Type: "text", Text: value}
	case ast.KindCodeSpan:
		// CodeSpan carries no opening and closing marker lengths, so
		// single backticks are the closest faithful rendering.
		return genericNode{Type: "code", Text: "`" + string(n.Text(src)) + "`"}
	case ast.KindList:
		return genericNode{Type: "list", Children: convertChildren(n, src)}
	case ast.KindListItem:
		return genericNode{Type: "list_item", Children: convertChildren(n, src)}
	case ast.KindEmphasis:
		if n.(*ast.Emphasis).Level == 2 {
			return genericNode{Type: "strong", Children: convertChildren(n, src)}
		}
		return genericNode{Type: "emph", Children: convertChildren(n, src)}
	case ast.KindLink:
		return genericNode{Type: "link", Children: convertChildren(n, src)}
	default:
		gn := genericNode{Type: strings.ToLower(n.Kind().String())}
		if n.HasChildren() {
			gn.Children = convertChildren(n, src)
		} else if n.Type() == ast.TypeInline {
			gn.Text = strings.TrimRight(string(n.Text(src)), "\n")
		}
		return gn
	}
}

// convertChildren maps the children of an AST node.
func convertChildren(n ast.Node, src []byte) []genericNode {
	if n.ChildCount() == 0 {
		return nil
	}
	children := make([]genericNode, 0, n.ChildCount())
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		children = append(children, convertNode(c, src))
	}
	return children
}

// fromGenericJSON maps a generic JSON tree to a Document.
func fromGenericJSON(root genericNode) (Document, error) {
	doc := Document{Requirements: []Requirement{}}
	current := -1
	inPurpose := false
	titleSet := false
	for _, child := range root.Children {
		if child.Type != "heading" {
			switch {
			case inPurpose:
				doc.Purpose = joinProse(doc.Purpose, proseBlock(child))
			case current >= 0 && len(doc.Requirements[current].Scenarios) > 0:
				if child.Type == "list" {
					scenario := &doc.Requirements[current].Scenarios[len(doc.Requirements[current].Scenarios)-1]
					scenario.Steps = append(scenario.Steps, stepsBlock(child)...)
				}
			case current >= 0:
				doc.Requirements[current].Body = joinProse(doc.Requirements[current].Body, proseBlock(child))
			}
			continue
		}
		level := child.Level
		text := renderNode(child)
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
func proseBlock(n genericNode) string {
	if n.Type != "paragraph" {
		return ""
	}
	return strings.TrimSpace(renderNode(n))
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

// stepsBlock renders a list's items as "- " prefixed step lines.
func stepsBlock(n genericNode) []string {
	steps := []string{}
	for _, item := range n.Children {
		if item.Type != "list_item" {
			continue
		}
		line := strings.TrimSpace(renderNode(item))
		if line != "" {
			steps = append(steps, "- "+line)
		}
	}
	return steps
}

// renderNode renders a generic node as text. Containers concatenate their
// children. Strong, emph, and link drop their markers. Code nodes carry
// their backticks already.
func renderNode(n genericNode) string {
	if len(n.Children) == 0 {
		return n.Text
	}
	var sb strings.Builder
	for _, child := range n.Children {
		sb.WriteString(renderNode(child))
	}
	return sb.String()
}
