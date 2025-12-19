package game

import (
	"strings"
	"testing"
)

func TestGame_NameValidation(t *testing.T) {
	testCases := []struct {
		testName    string
		gameName    string
		shouldPass  bool
		description string
	}{
		{
			testName:    "valid name",
			gameName:    "Borderlands",
			shouldPass:  true,
			description: "valid name should pass",
		},
		{
			testName:    "empty name",
			gameName:    "",
			shouldPass:  false,
			description: "empty name should fail",
		},
		{
			testName:    "min length name",
			gameName:    "A",
			shouldPass:  true,
			description: "single character name should pass",
		},
		{
			testName:    "max length name",
			gameName:    strings.Repeat("a", MaxNameLength),
			shouldPass:  true,
			description: "name at max length should pass",
		},
		{
			testName:    "too long name",
			gameName:    strings.Repeat("a", MaxNameLength+1),
			shouldPass:  false,
			description: "name longer than max length should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			game := &Game{
				Name: tc.gameName,
			}

			v := game.Validate()

			if tc.shouldPass {
				if v.IsInError("name") {
					t.Errorf("%s: Expected no validation error, got: %s", tc.description, v.GetErrorMessagesFor("name"))
				}
			} else {
				if !v.IsInError("name") {
					t.Errorf("%s: Expected validation error, but validation passed", tc.description)
				}
			}
		})
	}
}
