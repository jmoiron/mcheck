package main

import (
	"testing"
)

func TestStructParsing(t *testing.T) {
	input := `struct TestStruct {
	field1: string,
	field2?: int
}`

	parser := &MCDocParser{
		Buffer: input,
		Pretty: true,
	}

	err := parser.Init()
	if err != nil {
		t.Fatalf("Failed to initialize parser: %v", err)
	}

	err = parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Execute the actions to build the statements
	parser.Execute()

	t.Logf("Parser has %d statements", len(parser.Statements))
	
	// Check that we captured the struct statement
	if len(parser.Statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(parser.Statements))
	}

	if len(parser.Statements) > 0 {
		stmt := parser.Statements[0]
		if structStmt, ok := stmt.(StructStatement); ok {
			t.Logf("Struct statement: %s", structStmt.Name.Name)
			
			if structStmt.Name.Name != "TestStruct" {
				t.Errorf("Expected struct name 'TestStruct', got %s", structStmt.Name.Name)
			}
		} else {
			t.Errorf("Expected StructStatement, got %T", stmt)
		}
	}
}