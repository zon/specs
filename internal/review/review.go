// Package review runs a repository review.
package review

import (
	"github.com/zon/specs/internal/gitops"
	"github.com/zon/specs/internal/opencode"
)

// Options selects the review scope, model, and variant.
type Options struct {
	Scope   opencode.Scope
	Model   string
	Variant string
}

// Run reviews the repository against the guidelines for the scope.
func Run(opts Options) error {
	root, err := gitops.Root()
	if err != nil {
		return err
	}
	return opencode.Review(root, opts.Scope, opts.Model, opts.Variant)
}
