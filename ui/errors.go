package ui

import "soloterm/shared/validation"

// ValidationError wraps a validator for error handling
type ValidationError struct {
	Validator *validation.Validator
}

func (e *ValidationError) Error() string {
	return e.Validator.GetAllErrorMessages()
}
