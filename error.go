package ego

import "errors"

var (
	// ErrDeclarationRequired is returned when there is no declaration block
	// in a template.
	ErrDeclarationRequired = errors.New("declaration required")
)
