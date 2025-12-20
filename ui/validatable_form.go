package ui

import "soloterm/shared/validation"

// ValidatableForm represents a form that can display validation errors
type ValidatableForm interface {
	SetFieldErrors(errors map[string]string)
}

// handleValidationError processes validation errors and applies them to a form
func handleValidationError(err error, form ValidatableForm) bool {
	if validator, ok := err.(*validation.Validator); ok {
		// Build field errors map from validator
		fieldErrors := make(map[string]string)
		for _, fieldErr := range validator.Errors {
			fieldErrors[fieldErr.Field] = fieldErr.FormattedErrorMessage()
		}

		// Set all errors at once and update labels
		form.SetFieldErrors(fieldErrors)
		return true
	}
	return false
}
