package opencode

import (
	"fmt"
	"os/exec"
	"strings"
)

// DefaultModel is the model opencode uses unless the user gives one.
const DefaultModel = "deepseek/deepseek-v4-flash"

// Scope selects the review guidelines.
type Scope int

const (
	// ScopeCode reviews against the code guidelines.
	ScopeCode Scope = iota
	// ScopeArchitecture reviews against the architecture guidelines.
	ScopeArchitecture
	// ScopeProse reviews against the prose guidelines.
	ScopeProse
)

// UnmarshalText maps a scope token to its constant.
func (s *Scope) UnmarshalText(text []byte) error {
	switch string(text) {
	case "code":
		*s = ScopeCode
	case "architecture":
		*s = ScopeArchitecture
	case "prose":
		*s = ScopeProse
	default:
		return fmt.Errorf("unknown scope %q", string(text))
	}
	return nil
}

// guideline returns the guidelines doc path for a scope.
func guideline(scope Scope) (string, error) {
	switch scope {
	case ScopeCode:
		return "docs/zpecs/code.md", nil
	case ScopeArchitecture:
		return "docs/zpecs/architecture.md", nil
	case ScopeProse:
		return "docs/zpecs/prose.md", nil
	default:
		return "", fmt.Errorf("unknown scope %d", scope)
	}
}

// prompt builds the review prompt naming the scope's guidelines.
func prompt(scope Scope) (string, error) {
	doc, err := guideline(scope)
	if err != nil {
		return "", err
	}
	return "Review the repository against the guidelines in " + doc + ".", nil
}

// Review runs opencode on the repository at root with the model and the
// scope's guidelines. An empty model selects DefaultModel.
func Review(root string, scope Scope, model string) error {
	if model == "" {
		model = DefaultModel
	}
	message, err := prompt(scope)
	if err != nil {
		return err
	}
	cmd := exec.Command("opencode", "run", "--model", model, message)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("opencode review failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
