package validation

import (
	"testing"
)

func TestValidator(t *testing.T) {

	t.Run("no errors", func(t *testing.T) {

		v := NewValidator()

		v.Check("id", 1 == 1, "1 should equal 1")

		if v.HasErrors() {
			t.Error("Expected no errors")
		}
		if v.HasError("id") {
			t.Error("Expected no error")
		}
		if v.GetError("id") != nil {
			t.Error("Expected no error")
		}
		if v.Error() != "" {
			t.Errorf("Expected blank message for error %s", v.Error())
		}

	})

	t.Run("one identifier", func(t *testing.T) {

		v := NewValidator()
		if v.Error() != "" {
			t.Errorf("Expected blank message for error %s", v.Error())
		}

		v.Check("id", 1 == 2, "1 shouldn't equal 2")

		if !v.HasErrors() {
			t.Error("Expected an error")
		}

		if !v.HasError("id") {
			t.Error("Expected an error on id")
		}

		if v.GetError("id")[0] != "1 shouldn't equal 2" {
			t.Errorf("Expected message '1 shouldn't equal 2', but got %s", v.GetError("id")[0])
		}
	})

	t.Run("one identifier, multiple errors", func(t *testing.T) {
		v := NewValidator()
		v.Check("id", 1 == 2, "1 shouldn't equal 2")
		v.Check("id", 2 == 3, "2 shouldn't equal 3")
		v.Check("name", "a" == "a", "a should equal a")

		if !v.HasErrors() {
			t.Errorf("Expected an error")
		}

		if !v.HasError("id") {
			t.Errorf("Expected an error on id")
		}

		if v.HasError("name") {
			t.Errorf("Found an error on name, there shouldn't be one")
		}

		// There should be 2 messages for 1 error
		if len(v.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(v.Errors))
		}

		if len(v.GetError("id")) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(v.GetError("id")))
		}

		if v.GetError("id")[0] != "1 shouldn't equal 2" {
			t.Errorf("Expected message '1 shouldn't equal 2', but got %s", v.GetError("id")[0])
		}

		if v.GetError("id")[1] != "2 shouldn't equal 3" {
			t.Errorf("Expected message '2 shouldn't equal 3', but got %s", v.GetError("id")[0])
		}

	})

	t.Run("multiple identifiers", func(t *testing.T) {
		v := NewValidator()
		v.Check("id", 1 == 2, "1 shouldn't equal 2")
		v.Check("name", "a" == "b", "a shouldn't equal b")
		v.Check("name", "a" == "a", "a should equal a")

		if !v.HasErrors() {
			t.Errorf("Expected an error")
		}

		// There should be 2 errors
		if len(v.Errors) != 2 {
			t.Errorf("Expected 2 error, got %d", len(v.Errors))
		}

		if len(v.GetError("id")) != 1 {
			t.Errorf("Expected 1 message for id, got %d", len(v.GetError("id")))
		}

		if len(v.GetError("name")) != 1 {
			t.Errorf("Expected 1 message for name, got %d", len(v.GetError("name")))
		}

		if v.GetError("id")[0] != "1 shouldn't equal 2" {
			t.Errorf("Expected message 'a shouldn't equal b', but got %s", v.GetError("id")[0])
		}

		if v.GetError("name")[0] != "a shouldn't equal b" {
			t.Errorf("Expected message 'a shouldn't equal b', but got %s", v.GetError("id")[0])
		}

		if v.Error() != "id: 1 shouldn't equal 2; name: a shouldn't equal b" {
			t.Errorf("Expected message 'id: 1 shouldn't equal 2; name: a shouldn't equal b', got %s", v.Error())
		}

	})

}
