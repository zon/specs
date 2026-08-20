package frontmatter

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Fields holds the frontmatter of a definition. Every target reads the
// same fields.
type Fields struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"tools"`
	Mode        string   `yaml:"mode"`
}

// Content holds one definition's fields and body.
type Content struct {
	Fields Fields
	Body   string
}

// Read parses the definition at path into fields and body. A file without
// frontmatter has zero fields and its whole content as the body.
func Read(path string) (Content, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Content{}, err
	}
	return parse(string(content))
}

// parse reads the YAML block between the leading and closing --- lines,
// and the body after it.
func parse(content string) (Content, error) {
	lines := strings.Split(content, "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return Content{Body: strings.TrimSpace(content)}, nil
	}
	bodyAt := -1
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			bodyAt = i + 2
			break
		}
	}
	if bodyAt < 0 {
		return Content{}, fmt.Errorf("unterminated frontmatter")
	}
	block := strings.Join(lines[1:bodyAt-1], "\n")
	var fields Fields
	if err := yaml.Unmarshal([]byte(block), &fields); err != nil {
		return Content{}, err
	}
	body := strings.TrimSpace(strings.Join(lines[bodyAt:], "\n"))
	return Content{Fields: fields, Body: body}, nil
}
