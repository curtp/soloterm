// Package validation provides simple validation functionality.
// It provides a simple Check method for evaluating conditions and setting error messages.
// It collects validation errors and provides methods to query them.
// Inspired by this StackOverflow answer: https://stackoverflow.com/a/36652225
// and Rails ActiveModel validations.
package validation

import (
	"fmt"
	"strings"
)

// Validator collects validation errors for one or more identifiers.
type Validator struct {
	Errors []ValidationError
}

// ValidationError represents one or more validation errors for a specific identifier.
type ValidationError struct {
	Identifier string   // Used for grouping error messages together
	Messages   []string // Error messages for the identifier
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{}
}

// Check validates a condition and records an error if the condition is false.
//
// Usage Example:
//
//	type Entry struct {
//		Key       string
//		Namespace string
//		Value     string
//	}
//
//	func (e *Entry) Validate() *Validator {
//		v := NewValidator()
//		v.Check("key", len(e.Key) >= 5 && len(e.Key) <= 20, "must be between 5 and 20 characters")
//		v.Check("namespace", len(e.Namespace) >= 5 && len(e.Namespace) <= 20, "must be between 5 and 20 characters")
//		v.Check("value", len(e.Value) > 0, "may not be blank")
//		return v
//	}
//
//	func main() {
//		e := Entry{}
//		v := e.Validate()
//		if v.HasErrors() {
//			// Iterate over v.Errors to format as needed
//			for _, err := range v.Errors {
//				log.Printf("%s: %s", err.Identifier, err.Messages)
//			}
//		}
//	}
//
// Example for validating input to a service call:
//
//	v := validation.NewValidator()
//	v.Check("key", len(key) > 0, "must not be empty")
//	v.Check("namespace", helpers.IsURLSafe(namespace), "must contain only alphanumeric characters, hyphens, underscores, periods, or tildes")
//	if v.HasErrors() {
//		return nil, v // Return validator as error (implements error interface)
//	}
func (v *Validator) Check(identifier string, checkResult bool, errMsg string, args ...any) {
	if !checkResult {
		v.Errors = append(v.Errors, newValidationError(identifier, fmt.Sprintf(errMsg, args...)))
	}
}

// IsInError returns true if the given identifier has a validation error.
func (v *Validator) IsInError(identifier string) bool {
	return v.GetError(identifier) != nil
}

// HasErrors returns true if any validation errors have been recorded.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

// GetError returns the ValidationError for the given identifier, or nil if none exists.
func (v *Validator) GetError(identifier string) *ValidationError {
	for i := range v.Errors {
		if v.Errors[i].Identifier == identifier {
			return &v.Errors[i]
		}
	}
	return nil
}

// Error implements the error interface.
// Returns all errors formatted as "identifier: message; identifier2: message2".
func (v *Validator) Error() string {
	var errMsgs []string
	for _, err := range v.Errors {
		errMsgs = append(errMsgs, err.FormattedErrorMessage())
	}
	return strings.Join(errMsgs, "; ")
}

// newValidationError creates a new ValidationError with a single message.
func newValidationError(identifier string, message string) ValidationError {
	return ValidationError{Identifier: identifier, Messages: []string{message}}
}

// FormattedErrorMessage returns the error formatted as "identifier: message1, message2".
func (ve ValidationError) FormattedErrorMessage() string {
	return fmt.Sprintf("%s: %s", ve.Identifier, strings.Join(ve.Messages, ", "))
}
