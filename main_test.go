package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidationGoodFiles(t *testing.T) {
	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	parser := NewMCDocParser(version, "vanilla-mcdoc")
	
	// Walk through all JSON files in tests/good
	err = filepath.Walk("tests/good", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		
		// Test validation - should pass
		if err := parser.validateJSON(path); err != nil {
			t.Errorf("Expected validation to pass for %s, but got error: %v", path, err)
		}
		
		return nil
	})
	
	if err != nil {
		t.Fatalf("Error walking tests/good directory: %v", err)
	}
}

func TestValidationBadFiles(t *testing.T) {
	// Test with version 1.20.1 to catch schema validation issues
	version, err := parseVersion("1.20.1")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	parser := NewMCDocParser(version, "vanilla-mcdoc")
	
	// Walk through all JSON files in tests/bad
	err = filepath.Walk("tests/bad", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		
		// Test validation - should fail for schema validation reasons
		if err := parser.validateJSON(path); err == nil {
			t.Errorf("Expected validation to fail for %s, but it passed", path)
		} else {
			t.Logf("Expected failure for %s: %v", path, err)
		}
		
		return nil
	})
	
	if err != nil {
		t.Fatalf("Error walking tests/bad directory: %v", err)
	}
}