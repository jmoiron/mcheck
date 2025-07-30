package main

import (
	"strings"
)

// SchemaConverter converts parsed statements into proper validators
type SchemaConverter struct {
	version     Version
	statements  []Statement
	definitions map[string]Validator
}

func NewSchemaConverter(version Version, statements []Statement) *SchemaConverter {
	return &SchemaConverter{
		version:     version,
		statements:  statements,
		definitions: make(map[string]Validator),
	}
}

// ConvertToValidators creates proper validators from parsed statements
func (sc *SchemaConverter) ConvertToValidators() (map[string]Validator, error) {
	// First pass: create basic validators for all defined types
	for _, stmt := range sc.statements {
		switch s := stmt.(type) {
		case StructStatement:
			// Create a struct validator with basic fields
			structValidator := &StructValidator{
				BaseValidator: BaseValidator{},
				Fields:        []StructField{}, // Empty for now, will be populated later
			}
			sc.definitions[s.Name.Name] = structValidator
		case TypeAliasStatement:
			// For now, create a primitive validator
			aliasValidator := &PrimitiveValidator{
				BaseValidator: BaseValidator{},
				Type:          "any", // Accept any type for aliases for now
			}
			sc.definitions[s.Name.Name] = aliasValidator
		case DispatchStatement:
			// Create a dispatch validator that delegates to the target
			dispatchValidator := &PrimitiveValidator{
				BaseValidator: BaseValidator{},
				Type:          "any", // Accept any structure for dispatch
			}
			sc.definitions["_dispatch"] = dispatchValidator
		}
	}

	// Second pass: resolve references and build field validators
	// For now, keep it simple and focus on basic structure validation
	
	return sc.definitions, nil
}

// GetMainValidator finds the primary validator for validation
func (sc *SchemaConverter) GetMainValidator() Validator {
	// Look for dispatch statements first
	for _, stmt := range sc.statements {
		if _, ok := stmt.(DispatchStatement); ok {
			if validator, exists := sc.definitions["_dispatch"]; exists {
				return validator
			}
		}
	}

	// If no dispatch, look for structs that might be the main type
	for _, stmt := range sc.statements {
		if structStmt, ok := stmt.(StructStatement); ok {
			// Look for common main struct names
			name := strings.ToLower(structStmt.Name.Name)
			if strings.Contains(name, "settings") || strings.Contains(name, "generator") {
				if validator, exists := sc.definitions[structStmt.Name.Name]; exists {
					return validator
				}
			}
		}
	}

	// Return the first struct validator we find
	for _, stmt := range sc.statements {
		if structStmt, ok := stmt.(StructStatement); ok {
			if validator, exists := sc.definitions[structStmt.Name.Name]; exists {
				return validator
			}
		}
	}

	return nil
}

// CreateBasicStructValidator creates a basic struct validator that accepts any fields
// This allows for graceful validation where we accept the structure but don't validate details
func (sc *SchemaConverter) CreateBasicStructValidator() Validator {
	return &BasicStructValidator{
		BaseValidator: BaseValidator{},
	}
}

// BasicStructValidator accepts any object structure
type BasicStructValidator struct {
	BaseValidator
}

func (bsv BasicStructValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !bsv.AppliesForVersion(ctx) {
		return nil
	}
	
	// Accept any map[string]interface{} (JSON object)
	if _, ok := value.(map[string]interface{}); !ok {
		return ValidationError{Path: ctx.Path, Message: "expected object structure"}
	}
	
	return nil // Accept any fields within the object
}