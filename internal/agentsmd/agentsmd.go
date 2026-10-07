// Package agentsmd edits the "## Zpecs" section of a repository's
// AGENTS.md, leaving the rest of the document alone.
package agentsmd

import (
	"os"
	"path/filepath"
	"strings"
)

// FileName is the document at the repository root that holds the section.
const FileName = "AGENTS.md"

// Heading starts the section the system owns.
const Heading = "## Zpecs"

// Update replaces the "## Zpecs" section in root/AGENTS.md with section. It
// appends the section when the document has none and creates the document
// when it is absent. It leaves the rest of the text alone.
func Update(root, section string) error {
	path := filepath.Join(root, FileName)
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(Replace(string(content), section)), 0o644)
}

// Replace returns content with its "## Zpecs" section replaced by section,
// or section appended when content has no such section. It keeps the text
// before and after the section.
func Replace(content, section string) string {
	lines := strings.Split(content, "\n")
	start := sectionStart(lines)
	var blocks []string
	if start == -1 {
		blocks = addBlock(blocks, content)
	} else {
		blocks = addBlock(blocks, strings.Join(lines[:start], "\n"))
	}
	blocks = addBlock(blocks, section)
	if start != -1 {
		if end := sectionEnd(lines, start); end < len(lines) {
			blocks = addBlock(blocks, strings.Join(lines[end:], "\n"))
		}
	}
	return strings.Join(blocks, "\n\n") + "\n"
}

// sectionStart returns the index of the line that starts the section, or -1.
func sectionStart(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == Heading {
			return i
		}
	}
	return -1
}

// sectionEnd returns the index of the line that ends the section, which is
// the first heading after it or the end of the document.
func sectionEnd(lines []string, start int) int {
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			return i
		}
	}
	return len(lines)
}

// addBlock adds text as a block when it holds more than whitespace.
func addBlock(blocks []string, text string) []string {
	if strings.TrimSpace(text) == "" {
		return blocks
	}
	return append(blocks, strings.TrimSpace(text))
}
