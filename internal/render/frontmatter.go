package render

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// fields holds the frontmatter of a definition. Every target reads the
// same fields.
type fields struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"tools"`
	Mode        string   `yaml:"mode"`
}

// content holds one definition's fields and body.
type content struct {
	fields fields
	body   string
}

// read parses the definition at path into fields and body. A file without
// frontmatter yields zero fields and its whole content as the body.
func read(path string) (content, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return content{}, fmt.Errorf("reading %s: %w", path, err)
	}
	out, err := parse(string(raw))
	if err != nil {
		return content{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return out, nil
}

// parse reads the YAML block between the leading and closing --- lines,
// and the body after it.
func parse(text string) (content, error) {
	lines := strings.Split(text, "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return content{body: strings.TrimSpace(text)}, nil
	}
	bodyAt := -1
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			bodyAt = i + 2
			break
		}
	}
	if bodyAt < 0 {
		return content{}, fmt.Errorf("unterminated frontmatter")
	}
	block := strings.Join(lines[1:bodyAt-1], "\n")
	var f fields
	if err := yaml.Unmarshal([]byte(block), &f); err != nil {
		return content{}, err
	}
	body := strings.TrimSpace(strings.Join(lines[bodyAt:], "\n"))
	return content{fields: f, body: body}, nil
}
