package game

import (
	"strings"
	"testing"
)

func TestGame_Validation(t *testing.T) {
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
		{
			testName:    "description description",
			gameName:    "name",
			shouldPass:  true,
			description: "nasdf",
		},
		{
			testName:    "too short description",
			gameName:    "name",
			shouldPass:  false,
			description: "n",
		},
		{
			testName:    "too long description",
			gameName:    "name",
			shouldPass:  false,
			description: strings.Repeat("a", MaxDescriptionLength+1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			game := &Game{
				Name:        tc.gameName,
				Description: &tc.description,
			}

			v := game.Validate()

			if tc.shouldPass {
				if v.HasErrors() {
					t.Errorf("%s: Expected no validation error, got: %v", tc.description, v.Errors)
				}
			} else {
				if !v.HasErrors() {
					t.Error("Expected a validation error, but didn't find one.")
				}
			}
		})
	}
}
