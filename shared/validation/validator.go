package validation

// Inspired by this StackOverflow answer: https://stackoverflow.com/a/36652225
// Other inspiration from Rails ActiveModel validations

import (
	"fmt"
	"strings"
)

type Validator struct {
	Errors []ValidationError
}

type ValidationError struct {
	Field    string
	Messages []string
}

// Create a new validator with the Errors field initialized to an empty slice
func NewValidator() *Validator {
	return &Validator{}
}

// Checks to see if the condition is true. If not, appends an error to the Errors field.
//
// Usage Example:
//
//	 func (e *Entry) Validate() error {
//	   validator := NewValidator()
//		  var errorArray []validator.ValidationError
//		  v.check("key", helpers.isUrlSafe(key), "must contain only alphanumeric characters, hyphens, underscores, periods, or tildes")
//		  return v
//	 }
//
// Example for validating input to a service call
//
//	v := validation.NewValidator()
//	v.Check("key", helpers.IsURLSafe(key), "must contain only alphanumeric characters, hyphens, underscores, periods, or tildes")
//	v.Check("namespace", helpers.IsURLSafe(namespace), "must contain only alphanumeric characters, hyphens, underscores, periods, or tildes")
//	if v.HasErrors() {
//		return nil, errors.New(v.GetAllErrorMessages())
//	}
func (v *Validator) Check(identifier string, checkResult bool, errMsg string, args ...any) {
	if !checkResult {
		v.Errors = append(v.Errors, newValidationError(identifier, fmt.Sprintf(errMsg, args...)))
	}
}

// Checks to see if there is an error for the given identifier
func (v *Validator) IsInError(identifier string) bool {
	return v != nil && v.GetError(identifier) != nil
}

// Checks to see if there are any errors
func (v *Validator) HasErrors() bool {
	return v != nil && len(v.Errors) > 0
}

// Returns the ValidationError for the given identifier
func (v *Validator) GetError(identifier string) *ValidationError {
	for i := range v.Errors {
		if v.Errors[i].Field == identifier {
			return &v.Errors[i]
		}
	}
	return nil
}

// Returns the error messages for the given identifier
func (v *Validator) GetErrorMessagesFor(identifier string) string {
	if err := v.GetError(identifier); err != nil {
		return err.FormattedErrorMessage()
	}
	return ""
}

// Returns all error messages
func (v *Validator) GetAllErrorMessages() string {
	var errMsgs []string
	for _, err := range v.Errors {
		errMsgs = append(errMsgs, err.FormattedErrorMessage())
	}
	return strings.Join(errMsgs, "; ")
}

// Creates a new ValidationError
func newValidationError(field string, message string) ValidationError {
	return ValidationError{Field: field, Messages: []string{message}}
}

// Formats the error message as "field: message1, message2".
func (ve ValidationError) FormattedErrorMessage() string {
	return fmt.Sprintf("%s: %s", ve.Field, strings.Join(ve.Messages, ", "))
}
