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

func TestParseVersionRequirement(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *VersionRequirement
	}{
		{
			name:  "until version 1.19",
			input: `#[until="1.19"]`,
			expected: &VersionRequirement{
				Since: nil,
				Until: &Version{Major: 1, Minor: 19, Patch: 0},
			},
		},
		{
			name:  "until version 1.18",
			input: `#[until="1.18"]`,
			expected: &VersionRequirement{
				Since: nil,
				Until: &Version{Major: 1, Minor: 18, Patch: 0},
			},
		},
		{
			name:  "until version 1.19.4",
			input: `#[until="1.19.4"]`,
			expected: &VersionRequirement{
				Since: nil,
				Until: &Version{Major: 1, Minor: 19, Patch: 4},
			},
		},
		{
			name:  "since version 1.19.4",
			input: `#[since="1.19.4"]`,
			expected: &VersionRequirement{
				Since: &Version{Major: 1, Minor: 19, Patch: 4},
				Until: nil,
			},
		},
		{
			name:  "since version 1.18.2",
			input: `#[since="1.18.2"]`,
			expected: &VersionRequirement{
				Since: &Version{Major: 1, Minor: 18, Patch: 2},
				Until: nil,
			},
		},
		{
			name:     "no version requirement",
			input:    `category: BiomeCategory,`,
			expected: nil,
		},
		{
			name:     "doc comment",
			input:    `/// Raises or lowers the terrain. Positive values are considered land and negative are oceans.`,
			expected: nil,
		},
		{
			name:     "empty line",
			input:    "",
			expected: nil,
		},
		{
			name:     "struct definition",
			input:    `spawners: struct {`,
			expected: nil,
		},
		{
			name:  "tabbed until version",
			input: `	#[until="1.18"]`,
			expected: &VersionRequirement{
				Since: nil,
				Until: &Version{Major: 1, Minor: 18, Patch: 0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseVersionRequirement(tc.input)

			if tc.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, but got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected %+v, but got nil", tc.expected)
				return
			}

			// Check Since version
			if tc.expected.Since == nil {
				if result.Since != nil {
					t.Errorf("Expected Since to be nil, but got %+v", result.Since)
				}
			} else {
				if result.Since == nil {
					t.Errorf("Expected Since to be %+v, but got nil", tc.expected.Since)
				} else if result.Since.Compare(*tc.expected.Since) != 0 {
					t.Errorf("Expected Since to be %+v, but got %+v", tc.expected.Since, result.Since)
				}
			}

			// Check Until version
			if tc.expected.Until == nil {
				if result.Until != nil {
					t.Errorf("Expected Until to be nil, but got %+v", result.Until)
				}
			} else {
				if result.Until == nil {
					t.Errorf("Expected Until to be %+v, but got nil", tc.expected.Until)
				} else if result.Until.Compare(*tc.expected.Until) != 0 {
					t.Errorf("Expected Until to be %+v, but got %+v", tc.expected.Until, result.Until)
				}
			}
		})
	}
}
