package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPEGParser(t *testing.T) {
	testDir := "tests/mcdocs"
	
	// Walk through all .mcdoc files in the test directory
	err := filepath.WalkDir(testDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Only test .mcdoc files
		if !strings.HasSuffix(d.Name(), ".mcdoc") {
			return nil
		}
		
		t.Run(d.Name(), func(t *testing.T) {
			testPEGParseFile(t, path)
		})
		
		return nil
	})
	
	if err != nil {
		t.Fatalf("Failed to walk test directory: %v", err)
	}
}

func testPEGParseFile(t *testing.T, filePath string) {
	// Read the mcdoc file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	
	// Create a new parser instance
	parser := &MCDocParser{
		Buffer: string(content),
		Pretty: true,
	}
	
	// Initialize the parser
	err = parser.Init()
	if err != nil {
		t.Fatalf("Failed to initialize parser: %v", err)
	}
	
	// Parse the content
	err = parser.Parse()
	if err != nil {
		t.Errorf("Failed to parse %s: %v", filePath, err)
		
		// Print the content for debugging
		t.Logf("File content:\n%s", string(content))
		return
	}
	
	// Print the syntax tree for successful parses (for debugging)
	t.Logf("Successfully parsed %s", filePath)
	if testing.Verbose() {
		t.Logf("Syntax tree for %s:", filePath)
		parser.PrintSyntaxTree()
	}
}

// Test individual parsing rules for debugging
func TestPEGParserIndividualRules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		rule     int // rule index from the generated parser
		wantFail bool
	}{
		{
			name:  "simple use statement",
			input: "use super::loot::LootCondition",
			rule:  2, // ruleUseStmt
		},
		{
			name:  "type alias",
			input: "type Predicate = (LootCondition | [LootCondition])",
			rule:  4, // ruleTypeAlias  
		},
		{
			name:  "dispatch statement",
			input: `dispatch minecraft:resource[predicate] to Predicate`,
			rule:  12, // ruleDispatchStmt
		},
		{
			name:  "simple struct",
			input: "struct Test { field: string }",
			rule:  5, // ruleStructDef
		},
		{
			name:  "enum definition",
			input: `enum(string) Category { Building = "building", Misc = "misc" }`,
			rule:  9, // ruleEnumDef
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &MCDocParser{
				Buffer: tt.input,
				Pretty: true,
			}
			
			err := parser.Init()
			if err != nil {
				t.Fatalf("Failed to initialize parser: %v", err)
			}
			
			err = parser.Parse(tt.rule)
			if tt.wantFail {
				if err == nil {
					t.Errorf("Expected parsing to fail, but it succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Failed to parse %q: %v", tt.input, err)
					t.Logf("Input: %s", tt.input)
				} else {
					t.Logf("Successfully parsed %q", tt.input)
					if testing.Verbose() {
						parser.PrintSyntaxTree()
					}
				}
			}
		})
	}
}

// Test whitespace and comment handling
func TestPEGParserWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "single line comment",
			input: `// This is a comment
use super::test`,
		},
		{
			name: "doc comment",
			input: `/// This is a doc comment
/// Another line
use super::test`,
		},
		{
			name: "mixed whitespace",
			input: `
			
use super::test

// Comment here
type Test = string`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &MCDocParser{
				Buffer: tt.input,
				Pretty: true,
			}
			
			err := parser.Init()
			if err != nil {
				t.Fatalf("Failed to initialize parser: %v", err)
			}
			
			err = parser.Parse()
			if err != nil {
				t.Errorf("Failed to parse %q: %v", tt.input, err)
			} else {
				t.Logf("Successfully parsed whitespace test")
			}
		})
	}
}