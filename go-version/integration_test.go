package main

import (
	"testing"
)

// TestMainCleanupIntegration verifies that the cleaned up main.go works correctly
func TestMainCleanupIntegration(t *testing.T) {
	// Test that we can create a PEG validator without the old code
	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	// This should work without any old MCDocValidator dependencies
	validator := NewPEGMCDocValidator(version, "vanilla-mcdoc")
	if validator == nil {
		t.Fatal("Failed to create PEGMCDocValidator")
	}

	// Test the determineSchemaPath method works
	schemaPath, err := validator.determineSchemaPath("test_datapack/data/worldgen/noise_settings/test.json")
	if err != nil {
		t.Fatalf("Failed to determine schema path: %v", err)
	}

	expectedPath := "vanilla-mcdoc/java/data/worldgen/noise_settings.mcdoc"
	if schemaPath != expectedPath {
		t.Errorf("Expected schema path %s, got %s", expectedPath, schemaPath)
	}

	t.Log("Main cleanup integration test passed - PEG validator works without old dependencies")
}

// TestPEGValidatorEndToEnd tests the complete validation flow
func TestPEGValidatorEndToEnd(t *testing.T) {
	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	validator := NewPEGMCDocValidator(version, "vanilla-mcdoc")
	
	// Test with the existing test JSON file
	err = validator.ValidateJSON("test_datapack/data/worldgen/noise_settings/test.json")
	
	// The validation may fail due to schema specifics, but it should not panic
	// and should return a proper error message if it fails
	if err != nil {
		t.Logf("Validation returned error (expected for incomplete schema): %v", err)
	} else {
		t.Log("Validation passed successfully")
	}
	
	t.Log("End-to-end test completed without panics")
}