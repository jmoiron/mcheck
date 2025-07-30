package main

import (
	"testing"
)

func TestLiteralParsing(t *testing.T) {
	// Test parsing various literals to make sure they get pushed to the expression stack
	tests := []struct {
		name  string
		input string
	}{
		{"identifier", "test_identifier"},
		{"string", `"hello world"`},
		{"number", "42"},
		{"float", "3.14"},
		{"boolean true", "true"},
		{"boolean false", "false"},
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

			// Parse just the relevant rule based on the input
			var rule int
			switch tt.name {
			case "identifier":
				rule = int(ruleIdentifier)
			case "string":
				rule = int(ruleString)
			case "number", "float":
				rule = int(ruleNumber)
			case "boolean true", "boolean false":
				rule = int(ruleBoolean)
			}

			err = parser.Parse(rule)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.input, err)
			}

			parser.Execute()

			// Check that something was pushed to the expression stack
			if len(parser.ExprStack) == 0 {
				t.Errorf("Expected expression to be pushed to stack for %s", tt.input)
			} else {
				expr := parser.ExprStack[0]
				t.Logf("Parsed %s: %T = %s", tt.input, expr, expr.String())
			}
		})
	}
}