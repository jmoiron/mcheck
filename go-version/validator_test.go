package main

import (
	"testing"
)

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Version
		hasError bool
	}{
		{"1.20.1", Version{1, 20, 1}, false},
		{"1.19", Version{1, 19, 0}, false},
		{"2.0.0", Version{2, 0, 0}, false},
		{"invalid", Version{}, true},
		{"1", Version{}, true},
		{"1.2.3.4", Version{}, true},
		{"1.x.2", Version{}, true},
	}

	for _, test := range tests {
		result, err := parseVersion(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("For input %s, expected %v, got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestVersionComparison(t *testing.T) {
	v1, _ := parseVersion("1.20.1")
	v2, _ := parseVersion("1.20.2")
	v3, _ := parseVersion("1.19.4")
	v4, _ := parseVersion("2.0.0")

	if v1.Compare(v2) >= 0 {
		t.Error("1.20.1 should be less than 1.20.2")
	}
	if v2.Compare(v1) <= 0 {
		t.Error("1.20.2 should be greater than 1.20.1")
	}
	if v1.Compare(v3) <= 0 {
		t.Error("1.20.1 should be greater than 1.19.4")
	}
	if v4.Compare(v1) <= 0 {
		t.Error("2.0.0 should be greater than 1.20.1")
	}
	if v1.Compare(v1) != 0 {
		t.Error("1.20.1 should be equal to itself")
	}
}

func TestVersionString(t *testing.T) {
	v, _ := parseVersion("1.20.1")
	expected := "1.20.1"
	if v.String() != expected {
		t.Errorf("Expected version string %s, got %s", expected, v.String())
	}
}

func TestPrimitiveValidator(t *testing.T) {
	ctx := &ValidationContext{
		Version: Version{1, 20, 1},
		Path:    []string{},
	}

	// Test string validation
	stringValidator := &PrimitiveValidator{Type: "string"}
	
	// Valid string
	if err := stringValidator.Validate("hello", ctx); err != nil {
		t.Errorf("Expected string validation to pass, got: %v", err)
	}
	
	// Invalid string (number)
	if err := stringValidator.Validate(42, ctx); err == nil {
		t.Error("Expected string validation to fail for number, but it passed")
	}

	// Test int validation
	intValidator := &PrimitiveValidator{Type: "int"}
	
	// Valid int (JSON unmarshals numbers as float64)
	if err := intValidator.Validate(float64(42), ctx); err != nil {
		t.Errorf("Expected int validation to pass for float64, got: %v", err)
	}
	
	// Invalid int (string)
	if err := intValidator.Validate("42", ctx); err == nil {
		t.Error("Expected int validation to fail for string, but it passed")
	}

	// Test boolean validation
	boolValidator := &PrimitiveValidator{Type: "boolean"}
	
	// Valid boolean
	if err := boolValidator.Validate(true, ctx); err != nil {
		t.Errorf("Expected boolean validation to pass, got: %v", err)
	}
	
	// Invalid boolean (string)
	if err := boolValidator.Validate("true", ctx); err == nil {
		t.Error("Expected boolean validation to fail for string, but it passed")
	}
}

func TestStructValidator(t *testing.T) {
	ctx := &ValidationContext{
		Version: Version{1, 20, 1},
		Path:    []string{},
	}

	// Create a struct validator with required and optional fields
	structValidator := &StructValidator{
		Fields: []StructField{
			{
				Name:      "required_field",
				Validator: &PrimitiveValidator{Type: "string"},
				Optional:  false,
			},
			{
				Name:      "optional_field",
				Validator: &PrimitiveValidator{Type: "int"},
				Optional:  true,
			},
		},
	}

	// Test valid struct with both fields
	validData := map[string]interface{}{
		"required_field": "hello",
		"optional_field": float64(42),
	}
	if err := structValidator.Validate(validData, ctx); err != nil {
		t.Errorf("Expected validation to pass for valid struct, got: %v", err)
	}

	// Test valid struct with only required field
	validDataMinimal := map[string]interface{}{
		"required_field": "hello",
	}
	if err := structValidator.Validate(validDataMinimal, ctx); err != nil {
		t.Errorf("Expected validation to pass for struct with only required field, got: %v", err)
	}

	// Test invalid struct missing required field
	invalidDataMissing := map[string]interface{}{
		"optional_field": float64(42),
	}
	if err := structValidator.Validate(invalidDataMissing, ctx); err == nil {
		t.Error("Expected validation to fail for struct missing required field, but it passed")
	}

	// Test invalid struct with unexpected field
	invalidDataExtra := map[string]interface{}{
		"required_field":   "hello",
		"unexpected_field": "bad",
	}
	if err := structValidator.Validate(invalidDataExtra, ctx); err == nil {
		t.Error("Expected validation to fail for struct with unexpected field, but it passed")
	}
}