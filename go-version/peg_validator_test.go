package main

import (
	"os"
	"testing"
)

func TestPEGValidatorBasic(t *testing.T) {
	// Create a test version
	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	// Create validator
	validator := NewPEGMCDocValidator(version, "vanilla-mcdoc")
	
	// Test parsing a simple schema
	statements, definitions, err := validator.parseSchemaWithPEG("vanilla-mcdoc/java/data/worldgen/noise_settings.mcdoc")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	t.Logf("Parsed %d statements and %d definitions", len(statements), len(definitions))
	
	// Check that we got some statements
	if len(statements) == 0 {
		t.Error("Expected some statements, got none")
	}

	// Print statement types for debugging
	for i, stmt := range statements {
		switch s := stmt.(type) {
		case UseStatement:
			t.Logf("Statement %d: Use - %s", i, s.Path.String())
		case StructStatement:
			t.Logf("Statement %d: Struct - %s", i, s.Name.Name)
		case TypeAliasStatement:
			t.Logf("Statement %d: TypeAlias - %s", i, s.Name.Name)
		case DispatchStatement:
			t.Logf("Statement %d: Dispatch - %s", i, s.Path)
		default:
			t.Logf("Statement %d: Unknown type - %T", i, s)
		}
	}
}

func TestPEGValidatorFindMainValidator(t *testing.T) {
	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	validator := NewPEGMCDocValidator(version, "vanilla-mcdoc")
	
	statements, definitions, err := validator.parseSchemaWithPEG("vanilla-mcdoc/java/data/worldgen/noise_settings.mcdoc")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	mainValidator := validator.findMainValidator(statements, definitions)
	if mainValidator == nil {
		t.Error("Expected to find a main validator, got nil")
	} else {
		t.Logf("Found main validator: %T", mainValidator)
	}
}

func TestPEGValidatorJSONValidation(t *testing.T) {
	// Check if test files exist
	if _, err := os.Stat("test_datapack/data/worldgen/noise_settings/test.json"); os.IsNotExist(err) {
		t.Skip("Test JSON file not found")
	}
	if _, err := os.Stat("vanilla-mcdoc/java/data/worldgen/noise_settings.mcdoc"); os.IsNotExist(err) {
		t.Skip("Schema file not found")
	}

	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	validator := NewPEGMCDocValidator(version, "vanilla-mcdoc")
	
	// This should not panic and should return a reasonable error or success
	err = validator.ValidateJSON("test_datapack/data/worldgen/noise_settings/test.json")
	
	// For now, just check it doesn't panic - we'll improve validation next
	t.Logf("Validation result: %v", err)
}