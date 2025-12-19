package log

import (
	"testing"
)

func TestLog_LogTypeValidation(t *testing.T) {
	testCases := []struct {
		testName    string
		log_type    LogType
		shouldPass  bool
		description string
	}{
		{
			testName:    "character action",
			log_type:    CHARACTER_ACTION,
			shouldPass:  true,
			description: "character_action should pass",
		},
		{
			testName:    "oracle question",
			log_type:    ORACLE_QUESTION,
			shouldPass:  true,
			description: "oracle_question should pass",
		},
		{
			testName:    "mechanics",
			log_type:    MECHANICS,
			shouldPass:  true,
			description: "mechanics should pass",
		},
		{
			testName:    "invalid",
			log_type:    "FOOBAR",
			shouldPass:  false,
			description: "invalid should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			log := &Log{
				LogType: tc.log_type,
			}

			v := log.Validate()

			if tc.shouldPass {
				if v.IsInError("log_type") {
					t.Errorf("%s: Expected no validation error, got: %s", tc.description, v.GetErrorMessagesFor("name"))
				}
			} else {
				if !v.IsInError("log_type") {
					t.Errorf("%s: Expected validation error, but validation passed", tc.description)
				}
			}
		})
	}
}

func TestLog_RequiredFieldValidation(t *testing.T) {
	testCases := []struct {
		testName         string
		log_type         LogType
		description      string
		result           string
		narrative        string
		shouldPass       bool
		test_description string
	}{
		{
			testName:         "all fields",
			log_type:         CHARACTER_ACTION,
			description:      "foobar",
			result:           "barfoo",
			narrative:        "barfoo",
			shouldPass:       true,
			test_description: "should pass",
		},
		{
			testName:         "description",
			log_type:         CHARACTER_ACTION,
			description:      "",
			result:           "barfoo",
			narrative:        "barfoo",
			shouldPass:       false,
			test_description: "description is required",
		},
		{
			testName:         "result",
			log_type:         CHARACTER_ACTION,
			description:      "foobar",
			result:           "",
			narrative:        "barfoo",
			shouldPass:       false,
			test_description: "result is required",
		},
		{
			testName:         "narrative",
			log_type:         CHARACTER_ACTION,
			description:      "foobar",
			result:           "barfoo",
			narrative:        "",
			shouldPass:       false,
			test_description: "narrative is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			log := &Log{
				LogType:     tc.log_type,
				Description: tc.description,
				Result:      tc.result,
				Narrative:   tc.narrative,
			}

			v := log.Validate()

			if tc.shouldPass {
				if v.HasErrors() {
					t.Errorf("Expected no validation error, got: %s", v.GetAllErrorMessages())
				}
			} else {
				if !v.HasErrors() {
					t.Error("Expected validation error, but validation passed")
				}
			}
		})
	}
}
