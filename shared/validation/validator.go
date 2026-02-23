// Package validation provides simple validation functionality for collecting and querying validation errors.
//
// Inspired by a Stack Overflow post I can no longer locate. Specifically the Check function. It is not mine.
// Credit to the helpful person who posted it. I wish I could add a link here.
package validation

import (
	"fmt"
	"slices"
	"strings"
)

// Validator collects validation errors for one or more identifiers.
type Validator struct {
	Errors map[string][]string
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string][]string),
	}
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
//		v.Check("key", len(e.Key) >= 5, "must be at least 5 characters")
//		v.Check("key", len(e.Key) <= 20, "must be at most 20 characters")
//		v.Check("namespace", len(e.Namespace) > 0, "may not be blank")
//		v.Check("value", len(e.Value) > 0, "may not be blank")
//		return v
//	}
//
//	func main() {
//		e := Entry{Key: "ab"}
//		v := e.Validate()
//		if v.HasErrors() {
//			// Iterate over v.Errors to format as needed
//			for identifier, messages := range v.Errors {
//				fmt.Printf("%s: %s\n", identifier, strings.Join(messages, ", "))
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
		v.Errors[identifier] = append(v.Errors[identifier], fmt.Sprintf(errMsg, args...))
	}
}

// HasError returns true if the given identifier has a validation error.
func (v *Validator) HasError(identifier string) bool {
	_, exists := v.Errors[identifier]
	return exists
}

// HasErrors returns true if any validation errors have been recorded.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

// GetError returns the error messages for the given identifier, or nil if none exists.
func (v *Validator) GetError(identifier string) []string {
	return v.Errors[identifier]
}

// Error implements the error interface.
// Aggregates all errors into a semicolon delimited string, sorted by identifier.
func (v *Validator) Error() string {
	if len(v.Errors) == 0 {
		return ""
	}

	keys := make([]string, 0, len(v.Errors))
	for identifier := range v.Errors {
		keys = append(keys, identifier)
	}
	slices.Sort(keys)

	var errMsgs []string
	for _, identifier := range keys {
		errMsgs = append(errMsgs, fmt.Sprintf("%s: %s", identifier, strings.Join(v.Errors[identifier], ", ")))
	}
	return strings.Join(errMsgs, "; ")
}
