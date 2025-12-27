// Package validation provides field-level validation support for domain entities.
// It collects validation errors and provides methods to query and format them.
// Inspired by this StackOverflow answer: https://stackoverflow.com/a/36652225
// and Rails ActiveModel validations.
package validation

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
//	  struct Entry {
//	     Key       string
//	     Namespace string
//	     Value     string
//	  }
//
//		 func (e *Entry) Validate() (validator Validator, error) {
//		   v := NewValidator()
//		   v.check("key", len(Key) >= 5 && len(Key) <= 20, "must be between 5 and 20 characters")
//	    v.check("namespace", len(Namespace) >= 5 && len(Namespace) <= 20, "must be between 5 and 20 characters")
//		   v.check("value", len(value) > 0, "may not be blank")
//		   return v
//		 }
//
//	  func main() {
//			e := Entry{}
//			v, err := e.Validate
//			if v.HasErrors() {
//				log.Printf("errors: %s", v.GetAllErrorMessages())
//		    }
//		  }
//
// Example for validating input to a service call
//
//	v := validation.NewValidator()
//	v.Check("Key", len(key) > 0, "must contain only alphanumeric characters, hyphens, underscores, periods, or tildes")
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

// Error implements the error interface for Validator
func (v *Validator) Error() string {
	return v.GetAllErrorMessages()
}

// Creates a new ValidationError
func newValidationError(field string, message string) ValidationError {
	return ValidationError{Field: field, Messages: []string{message}}
}

// Formats the error message as "field: message1, message2".
func (ve ValidationError) FormattedErrorMessage() string {
	return fmt.Sprintf("%s: %s", ve.Field, strings.Join(ve.Messages, ", "))
}
