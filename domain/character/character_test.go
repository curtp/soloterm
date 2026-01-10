package character

import (
	"strings"
	"testing"
)

func TestCharacter_Validation(t *testing.T) {
	testCases := []struct {
		testName   string
		name       string
		system     string
		role       string
		species    string
		shouldPass bool
	}{
		{
			testName:   "valid name",
			name:       "Frank",
			role:       "Fighter",
			species:    "Human",
			system:     "Some System",
			shouldPass: true,
		},
		{
			testName:   "empty name",
			name:       "",
			system:     "Some System",
			role:       "Fighter",
			species:    "Human",
			shouldPass: false,
		},
		{
			testName:   "min length name",
			name:       "A",
			role:       "Fighter",
			species:    "Human",
			system:     "Some System",
			shouldPass: true,
		},
		{
			testName:   "max length name",
			name:       strings.Repeat("a", MaxNameLength),
			system:     "Some System",
			role:       "Fighter",
			species:    "Human",
			shouldPass: true,
		},
		{
			testName:   "too long name",
			name:       strings.Repeat("a", MaxNameLength+1),
			system:     "Some System",
			role:       "Fighter",
			species:    "Human",
			shouldPass: false,
		},
		{
			testName:   "empty system",
			name:       "Frank",
			system:     "",
			role:       "Fighter",
			species:    "Human",
			shouldPass: false,
		},
		{
			testName:   "min length system",
			name:       "Frank",
			system:     strings.Repeat("a", MinSystemLength),
			role:       "Fighter",
			species:    "Human",
			shouldPass: true,
		},
		{
			testName:   "max length system",
			name:       "frank",
			system:     strings.Repeat("a", MaxSystemLength),
			role:       "Fighter",
			species:    "Human",
			shouldPass: true,
		},
		{
			testName:   "too long system",
			name:       "frank",
			system:     strings.Repeat("a", MaxSystemLength+1),
			role:       "Fighter",
			species:    "Human",
			shouldPass: false,
		},
		{
			testName:   "empty role",
			name:       "Frank",
			system:     "FlexD6",
			role:       "",
			species:    "Human",
			shouldPass: false,
		},
		{
			testName:   "min length role",
			name:       "Frank",
			system:     "FlexD6",
			role:       strings.Repeat("a", MinRoleLength),
			species:    "Human",
			shouldPass: true,
		},
		{
			testName:   "max length role",
			name:       "frank",
			system:     "FlexD6",
			role:       strings.Repeat("a", MaxRoleLength),
			species:    "Human",
			shouldPass: true,
		},
		{
			testName:   "too long role",
			name:       "frank",
			system:     "FlexD6",
			role:       strings.Repeat("a", MaxRoleLength+1),
			species:    "Human",
			shouldPass: false,
		},
		{
			testName:   "empty species",
			name:       "Frank",
			system:     "FlexD6",
			role:       "Fighter",
			species:    "",
			shouldPass: false,
		},
		{
			testName:   "min length species",
			name:       "Frank",
			system:     "FlexD6",
			role:       "Fighter",
			species:    strings.Repeat("a", MinSpeciesLength),
			shouldPass: true,
		},
		{
			testName:   "max length species",
			name:       "frank",
			system:     "FlexD6",
			role:       "Fighter",
			species:    strings.Repeat("a", MaxSpeciesLength),
			shouldPass: true,
		},
		{
			testName:   "too long species",
			name:       "frank",
			system:     "FlexD6",
			role:       "Fighter",
			species:    strings.Repeat("a", MaxSpeciesLength+1),
			shouldPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			character := &Character{
				Name:    tc.name,
				System:  tc.system,
				Role:    tc.role,
				Species: tc.species,
			}

			v := character.Validate()

			if tc.shouldPass {
				if v.HasErrors() {
					t.Errorf("Expected no validation error, got: %v", v.Errors)
				}
			} else {
				if !v.HasErrors() {
					t.Error("Expected a validation error, but didn't find one.")
				}
			}
		})
	}
}

func TestCharacter_AttributeValidation(t *testing.T) {
	testCases := []struct {
		testName     string
		character_id int64
		name         string
		value        string
		shouldPass   bool
	}{
		{
			testName:     "valid attribute",
			character_id: 1,
			name:         "Health",
			value:        "10",
			shouldPass:   true,
		},
		{
			testName:     "missing character id",
			character_id: 0,
			name:         "Health",
			value:        "10",
			shouldPass:   false,
		},
		{
			testName:     "missing name",
			character_id: 1,
			name:         "",
			value:        "10",
			shouldPass:   false,
		},
		{
			testName:     "min length name",
			character_id: 1,
			name:         strings.Repeat("a", MinAttributeNameLength),
			value:        "10",
			shouldPass:   true,
		},
		{
			testName:     "max length name",
			character_id: 1,
			name:         strings.Repeat("a", MaxNameLength),
			value:        "10",
			shouldPass:   true,
		},
		{
			testName:     "too long name",
			character_id: 1,
			name:         strings.Repeat("a", MaxNameLength+1),
			value:        "10",
			shouldPass:   false,
		},
		{
			testName:     "missing value",
			character_id: 1,
			name:         "frank",
			value:        "",
			shouldPass:   true,
		},
		{
			testName:     "min length value",
			character_id: 1,
			name:         "frank",
			value:        strings.Repeat("a", MinAttributeValueLength),
			shouldPass:   true,
		},
		{
			testName:     "min length value",
			character_id: 1,
			name:         "frank",
			value:        strings.Repeat("a", MaxAttributeValueLength),
			shouldPass:   true,
		},
		{
			testName:     "too long name",
			character_id: 1,
			name:         "frank",
			value:        strings.Repeat("a", MaxAttributeValueLength+1),
			shouldPass:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			attribute := &Attribute{
				CharacterID: tc.character_id,
				Name:        tc.name,
				Value:       tc.value,
			}

			v := attribute.Validate()

			if tc.shouldPass {
				if v.HasErrors() {
					t.Errorf("Expected no validation error, got: %v", v.Errors)
				}
			} else {
				if !v.HasErrors() {
					t.Error("Expected a validation error, but didn't find one.")
				}
			}
		})
	}
}
