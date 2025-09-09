package main

import (
	"testing"
)

func TestStatementBuilderBasic(t *testing.T) {
	input := `use super::test::Type
use ::java::util::List`

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
	t.Logf("Parser definitions: %v", len(parser.Definitions))
	
	// Check that we captured the use statements
	if len(parser.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(parser.Statements))
	}

	if len(parser.Statements) > 0 {
		stmt := parser.Statements[0]
		if useStmt, ok := stmt.(UseStatement); ok {
			t.Logf("First use statement: %s", useStmt.Path.String())
			
			// Check the path structure
			if len(useStmt.Path.Segments) != 3 {
				t.Errorf("Expected 3 path segments, got %d", len(useStmt.Path.Segments))
			}
			
			if useStmt.Path.Segments[0].Value != "super" || !useStmt.Path.Segments[0].IsSuper {
				t.Errorf("Expected first segment to be 'super', got %s (IsSuper: %v)", 
					useStmt.Path.Segments[0].Value, useStmt.Path.Segments[0].IsSuper)
			}
		} else {
			t.Errorf("Expected UseStatement, got %T", stmt)
		}
	}

	if len(parser.Statements) > 1 {
		stmt := parser.Statements[1]
		if useStmt, ok := stmt.(UseStatement); ok {
			t.Logf("Second use statement: %s", useStmt.Path.String())
			
			// Check that it's absolute
			if !useStmt.Path.IsAbsolute {
				t.Errorf("Expected second path to be absolute")
			}
		} else {
			t.Errorf("Expected UseStatement, got %T", stmt)
		}
	}
}