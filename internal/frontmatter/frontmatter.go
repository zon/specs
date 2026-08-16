package frontmatter

import (
	"fmt"
	"os"
	"strings"
)

// Fields holds the frontmatter of a definition. Every target reads the
// same fields.
type Fields struct {
	Name        string
	Description string
	Tools       []string
}

// Read reads the frontmatter fields from the definition at path. A file
// without frontmatter has zero fields.
func Read(path string) (Fields, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Fields{}, err
	}
	return parse(string(content))
}

// parse reads the fields between the leading and closing --- lines.
func parse(content string) (Fields, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Fields{}, nil
	}
	var (
		fields Fields
		block  *[]string
	)
	open := true
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			open = false
			break
		}
		if trimmed == "" {
			continue
		}
		if item, ok := strings.CutPrefix(trimmed, "- "); ok {
			if block == nil {
				block = &fields.Tools
			}
			*block = append(*block, unquote(item))
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "name":
			fields.Name = unquote(value)
		case "description":
			fields.Description = unquote(value)
		case "tools":
			if inline := parseInlineList(value); inline != nil {
				fields.Tools = inline
			} else if value == "" {
				block = &fields.Tools
			} else {
				fields.Tools = []string{unquote(value)}
			}
		}
	}
	if open {
		return Fields{}, fmt.Errorf("unterminated frontmatter")
	}
	return fields, nil
}

// parseInlineList reads a [a, b] value, or nil when the value is not
// bracketed.
func parseInlineList(value string) []string {
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil
	}
	inner := strings.TrimSpace(value[1 : len(value)-1])
	var items []string
	if inner == "" {
		return items
	}
	for _, part := range strings.Split(inner, ",") {
		items = append(items, unquote(strings.TrimSpace(part)))
	}
	return items
}

// unquote strips one matching pair of surrounding quotes.
func unquote(value string) string {
	if len(value) >= 2 {
		switch {
		case value[0] == '"' && value[len(value)-1] == '"':
			return value[1 : len(value)-1]
		case value[0] == '\'' && value[len(value)-1] == '\'':
			return value[1 : len(value)-1]
		}
	}
	return value
}
