package ui

import (
	"soloterm/shared/validation"
	"strings"
)

// ValidatableForm represents a form that can display validation errors
type ValidatableForm interface {
	SetFieldErrors(errors map[string]string)
}

// handleValidationError processes validation errors and applies them to a form
func handleValidationError(err error, form ValidatableForm) bool {
	if validator, ok := err.(*validation.Validator); ok {
		// Build field errors map from validator
		fieldErrors := make(map[string]string)
		for identifier, messages := range validator.Errors {
			fieldErrors[identifier] = strings.Join(messages, ", ")
		}

		// Set all errors at once and update labels
		form.SetFieldErrors(fieldErrors)
		return true
	}
	return false
}
