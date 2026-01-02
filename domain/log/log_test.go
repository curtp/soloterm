package log

import (
	"testing"
)

func TestLog_LogTypeValidation(t *testing.T) {
	testCases := []struct {
		testName    string
		logType     LogType
		shouldPass  bool
		description string
	}{
		{
			testName:    "character action",
			logType:     CHARACTER_ACTION,
			shouldPass:  true,
			description: "character_action should pass",
		},
		{
			testName:    "oracle question",
			logType:     ORACLE_QUESTION,
			shouldPass:  true,
			description: "oracle_question should pass",
		},
		{
			testName:    "mechanics",
			logType:     MECHANICS,
			shouldPass:  true,
			description: "mechanics should pass",
		},
		{
			testName:    "invalid",
			logType:     "FOOBAR",
			shouldPass:  false,
			description: "invalid should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			log := &Log{
				LogType: tc.logType,
			}

			v := log.Validate()

			if tc.shouldPass {
				if v.HasError("log_type") {
					err := v.GetError("log_type")
					t.Errorf("%s: Expected no validation error, got: %v", tc.description, err.Messages)
				}
			} else {
				if !v.HasError("log_type") {
					t.Errorf("%s: Expected validation error, but validation passed", tc.description)
				}
			}
		})
	}
}

func TestLog_RequiredFieldValidation(t *testing.T) {
	testCases := []struct {
		testName        string
		logType         LogType
		description     string
		result          string
		narrative       string
		shouldPass      bool
		testDescription string
	}{
		{
			testName:        "all fields",
			logType:         CHARACTER_ACTION,
			description:     "foobar",
			result:          "barfoo",
			narrative:       "barfoo",
			shouldPass:      true,
			testDescription: "should pass",
		},
		{
			testName:        "description",
			logType:         CHARACTER_ACTION,
			description:     "",
			result:          "barfoo",
			narrative:       "barfoo",
			shouldPass:      false,
			testDescription: "description is required",
		},
		{
			testName:        "result",
			logType:         CHARACTER_ACTION,
			description:     "foobar",
			result:          "",
			narrative:       "barfoo",
			shouldPass:      false,
			testDescription: "result is required",
		},
		{
			testName:        "narrative",
			logType:         CHARACTER_ACTION,
			description:     "foobar",
			result:          "barfoo",
			narrative:       "",
			shouldPass:      false,
			testDescription: "narrative is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			log := &Log{
				LogType:     tc.logType,
				Description: tc.description,
				Result:      tc.result,
				Narrative:   tc.narrative,
			}

			v := log.Validate()

			if tc.shouldPass {
				if v.HasErrors() {
					t.Errorf("Expected no validation error, got: %v", v.Errors)
				}
			} else {
				if !v.HasErrors() {
					t.Error("Expected validation error, but validation passed")
				}
			}
		})
	}
}
