package opencode

import (
	"fmt"
	"os/exec"
	"strings"
)

// DefaultModel is the model opencode uses unless the user gives one.
const DefaultModel = "deepseek/deepseek-v4-flash"

// DefaultVariant is the variant the CLI passes when the user gives no
// model and no variant.
const DefaultVariant = "high"

// CodeReviewPrompt is the prompt the code scope gives opencode. It
// directs each issue to a refactor-<slug>.yaml project rather than
// editing the code.
const CodeReviewPrompt = "Review the repository against the guidelines in docs/zpecs/code.md. Do not edit the code. Write each issue you find to the refactor-<slug>.yaml project for its refactoring in projects/. Update a matching project when one exists. Write no project when you find no issues."

// ArchitectureReviewPrompt is the prompt the architecture scope gives
// opencode. It directs each issue to a refactor-<slug>.yaml project
// rather than editing the code.
const ArchitectureReviewPrompt = "Review the repository against the guidelines in docs/zpecs/architecture.md. Do not edit the code. Write each issue you find to the refactor-<slug>.yaml project for its refactoring in projects/."

// ProseReviewPrompt is the prompt the prose scope gives opencode. It
// fixes each issue by editing the offending text in place.
const ProseReviewPrompt = "Review the repository against the guidelines in docs/zpecs/prose.md. Fix each prose issue you find immediately by editing the offending text in place."

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

// prompt builds the review prompt naming the scope's guidelines.
func prompt(scope Scope) (string, error) {
	switch scope {
	case ScopeCode:
		return CodeReviewPrompt, nil
	case ScopeArchitecture:
		return ArchitectureReviewPrompt, nil
	case ScopeProse:
		return ProseReviewPrompt, nil
	default:
		return "", fmt.Errorf("unknown scope %d", scope)
	}
}

// Review runs opencode on the repository at root with the model, variant,
// and the scope's guidelines. An empty model selects DefaultModel. An
// empty variant omits the flag. opencode chooses its default.
func Review(root string, scope Scope, model, variant string) error {
	if model == "" {
		model = DefaultModel
	}
	message, err := prompt(scope)
	if err != nil {
		return err
	}
	args := []string{"run", "--model", model}
	if variant != "" {
		args = append(args, "--variant", variant)
	}
	args = append(args, message)
	cmd := exec.Command("opencode", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("opencode review failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
