package ui

import (
	"errors"
	"soloterm/shared/validation"
	"strings"
)

// ValidatableForm represents a form that can display validation errors
type ValidatableForm interface {
	SetFieldErrors(errors map[string]string)
}

// HandleValidationError processes validation errors and applies them to a form
func HandleValidationError(err error, form ValidatableForm) bool {
	var validator *validation.Validator

	// If the error can be cast as a validator, then assign it and continue
	if errors.As(err, &validator) {

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
