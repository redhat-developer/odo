package validation

import (
	"gopkg.in/AlecAivazis/survey.v1"
	"reflect"
	"testing"
)

func Test_validateName(t *testing.T) {
	tests := []struct {
		testCase string
		name     string
		wantErr  bool
	}{
		{
			testCase: "Test case - 1",
			name:     "app",
			wantErr:  false,
		},
		{
			testCase: "Test case - 2",
			name:     "app123",
			wantErr:  false,
		},
		{
			testCase: "Test case - 3",
			name:     "app-123",
			wantErr:  false,
		},
		{
			testCase: "Test case - 4",
			name:     "app.123",
			wantErr:  true,
		},
		{
			testCase: "Test case - 5",
			name:     "app_123",
			wantErr:  true,
		},
		{
			testCase: "Test case - 6",
			name:     "App",
			wantErr:  true,
		},
		{
			testCase: "Test case - 7",
			name:     "b7pdkva7ynxf8qoyuh02tbgu2ufcy4jq7udyom7it0g8gouc39x3gy0p1wrsbt6",
			wantErr:  false,
		},
		{
			testCase: "Test case - 8",
			name:     "b7pdkva7ynxf8qoyuh02tbgu2ufcy4jq7udyom7it0g8gouc39x3gy0p1wrsbt6p",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Log("Running test", tt.testCase)
		t.Run(tt.testCase, func(t *testing.T) {
			if err := ValidateName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("Expected error = %v, But got = %v", tt.wantErr, err)
			}
		})
	}
}

func TestGetValidator(t *testing.T) {
	// add test validator
	testValidator := func(ans interface{}) error { return nil }

	tests := []struct {
		name        string
		validatable Validatable
		expected    []survey.Validator
	}{
		{
			name:        "default",
			validatable: Validatable{},
			expected:    []survey.Validator{NilValidator},
		},
		{
			name:        "required",
			validatable: Validatable{Required: true},
			expected:    []survey.Validator{survey.Required},
		},
		{
			name:        "unknown type",
			validatable: Validatable{Type: "foo"},
			expected:    []survey.Validator{NilValidator},
		},
		{
			name:        "integer",
			validatable: Validatable{Type: "integer"},
			expected:    []survey.Validator{IntegerValidator},
		},
		{
			name:        "integer and required",
			validatable: Validatable{Type: "integer", Required: true},
			expected:    []survey.Validator{survey.Required, IntegerValidator},
		},
		{
			name:        "additional validator (name)",
			validatable: Validatable{AdditionalValidators: []Validator{NameValidator}},
			expected:    []survey.Validator{NameValidator},
		},
		{
			name:        "integer, required and additional validator (name)",
			validatable: Validatable{Type: "integer", Required: true, AdditionalValidators: []Validator{NameValidator}},
			expected:    []survey.Validator{survey.Required, IntegerValidator, NameValidator},
		},
		{
			name:        "test validator",
			validatable: Validatable{AdditionalValidators: []Validator{testValidator}},
			expected:    []survey.Validator{testValidator},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, chain := internalGetValidatorFor(tt.validatable)

			// if validator chain is empty, only possible result is NilValidator
			if len(chain) == 0 {
				// check that function pointers are equal
				f1 := reflect.ValueOf(NilValidator).Pointer()
				f2 := reflect.ValueOf(validator).Pointer()
				if f1 != f2 {
					t.Error("test failed, expected NilValidator")
				}
			} else {
				if len(tt.expected) != len(chain) {
					t.Errorf("test failed, validator chains don't have the same length, expected %d, got %d", len(tt.expected), len(chain))
				}

				for e := range chain {
					// check that function pointers are equal
					f1 := reflect.ValueOf(tt.expected[e]).Pointer()
					f2 := reflect.ValueOf(chain[e]).Pointer()
					if f1 != f2 {
						t.Errorf("test failed, different validators at position %d, expected %v, got %v", e, tt.expected[e], chain[e])
					}
				}
			}
		})
	}
}
