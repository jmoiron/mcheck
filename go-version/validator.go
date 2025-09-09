package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Version represents a Minecraft version for comparison
type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

func parseVersion(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch := 0
	if len(parts) == 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// ValidationContext holds context information for validation
type ValidationContext struct {
	Version     Version
	Path        []string // current path in the JSON for error reporting
	Definitions map[string]Validator // type definitions from use statements and type aliases
}

// ValidationError represents a validation error
type ValidationError struct {
	Path    []string
	Message string
}

func (e ValidationError) Error() string {
	if len(e.Path) == 0 {
		return e.Message
	}
	return fmt.Sprintf("at %s: %s", strings.Join(e.Path, "."), e.Message)
}

// Validator interface for all validation types
type Validator interface {
	Validate(value interface{}, ctx *ValidationContext) error
	AppliesForVersion(ctx *ValidationContext) bool
}

// BaseValidator contains common fields for version checking
type BaseValidator struct {
	Since string // version when this was introduced
	Until string // version when this was removed
}

func (bv BaseValidator) AppliesForVersion(ctx *ValidationContext) bool {
	if bv.Since != "" {
		sinceVersion, err := parseVersion(bv.Since)
		if err == nil && ctx.Version.Compare(sinceVersion) < 0 {
			return false
		}
	}
	if bv.Until != "" {
		untilVersion, err := parseVersion(bv.Until)
		if err == nil && ctx.Version.Compare(untilVersion) > 0 {
			return false
		}
	}
	return true
}

// PrimitiveValidator validates primitive types (string, int, float, boolean)
type PrimitiveValidator struct {
	BaseValidator
	Type string // "string", "int", "float", "boolean", "double", "any"
}

func (pv PrimitiveValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !pv.AppliesForVersion(ctx) {
		return nil
	}
	
	switch pv.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected string, got %T", value)}
		}
	case "int":
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return ValidationError{Path: ctx.Path, Message: "expected integer, got float"}
			}
		case int, int64:
			// OK
		default:
			return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected int, got %T", value)}
		}
	case "float", "double":
		if _, ok := value.(float64); !ok {
			return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected float, got %T", value)}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected boolean, got %T", value)}
		}
	case "any":
		// any type is always valid
	default:
		return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("unknown primitive type: %s", pv.Type)}
	}
	return nil
}

// RangeValidator validates numeric ranges with inclusive/exclusive bounds
type RangeValidator struct {
	BaseValidator
	Min         *float64
	Max         *float64
	MinExclusive bool
	MaxExclusive bool
}

func (rv RangeValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !rv.AppliesForVersion(ctx) {
		return nil
	}
	
	var numValue float64
	switch v := value.(type) {
	case float64:
		numValue = v
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	default:
		return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected number for range validation, got %T", value)}
	}
	
	if rv.Min != nil {
		if rv.MinExclusive {
			if numValue <= *rv.Min {
				return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("value %g must be greater than %g", numValue, *rv.Min)}
			}
		} else {
			if numValue < *rv.Min {
				return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("value %g must be greater than or equal to %g", numValue, *rv.Min)}
			}
		}
	}
	
	if rv.Max != nil {
		if rv.MaxExclusive {
			if numValue >= *rv.Max {
				return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("value %g must be less than %g", numValue, *rv.Max)}
			}
		} else {
			if numValue > *rv.Max {
				return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("value %g must be less than or equal to %g", numValue, *rv.Max)}
			}
		}
	}
	
	return nil
}

// ArrayValidator validates arrays with optional constraints
type ArrayValidator struct {
	BaseValidator
	ElementValidator Validator
	LengthConstraint *RangeValidator
}

func (av ArrayValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !av.AppliesForVersion(ctx) {
		return nil
	}
	
	arr, ok := value.([]interface{})
	if !ok {
		return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected array, got %T", value)}
	}
	
	// Validate array length if constrained
	if av.LengthConstraint != nil {
		lengthValue := float64(len(arr))
		if err := av.LengthConstraint.Validate(lengthValue, ctx); err != nil {
			return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("array length validation failed: %s", err.Error())}
		}
	}
	
	// Validate each element
	for i, elem := range arr {
		ctx.Path = append(ctx.Path, fmt.Sprintf("[%d]", i))
		if err := av.ElementValidator.Validate(elem, ctx); err != nil {
			return err
		}
		ctx.Path = ctx.Path[:len(ctx.Path)-1]
	}
	
	return nil
}

// StructField represents a field in a struct validator
type StructField struct {
	Name      string
	Validator Validator
	Optional  bool
	BaseValidator
}

// StructValidator validates object structures
type StructValidator struct {
	BaseValidator
	Fields       []StructField
	SpreadFields []Validator // for ...OtherStruct syntax
}

func (sv StructValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !sv.AppliesForVersion(ctx) {
		return nil
	}
	
	obj, ok := value.(map[string]interface{})
	if !ok {
		return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected object, got %T", value)}
	}
	
	// Track which fields we've seen
	seenFields := make(map[string]bool)
	
	// Validate each defined field
	for _, field := range sv.Fields {
		if !field.AppliesForVersion(ctx) {
			continue
		}
		
		fieldValue, exists := obj[field.Name]
		if !exists {
			if !field.Optional {
				return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("required field '%s' is missing", field.Name)}
			}
			continue
		}
		
		seenFields[field.Name] = true
		ctx.Path = append(ctx.Path, field.Name)
		if err := field.Validator.Validate(fieldValue, ctx); err != nil {
			return err
		}
		ctx.Path = ctx.Path[:len(ctx.Path)-1]
	}
	
	// Validate spread fields (additional properties allowed by ...OtherStruct)
	for fieldName, fieldValue := range obj {
		if seenFields[fieldName] {
			continue
		}
		
		// Try to validate against spread fields
		validated := false
		for _, spreadValidator := range sv.SpreadFields {
			ctx.Path = append(ctx.Path, fieldName)
			if err := spreadValidator.Validate(fieldValue, ctx); err == nil {
				validated = true
				ctx.Path = ctx.Path[:len(ctx.Path)-1]
				break
			}
			ctx.Path = ctx.Path[:len(ctx.Path)-1]
		}
		
		if !validated && len(sv.SpreadFields) == 0 {
			return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("unexpected field '%s'", fieldName)}
		}
	}
	
	return nil
}

// UnionValidator validates union types (value must match one of the alternatives)
type UnionValidator struct {
	BaseValidator
	Alternatives []Validator
}

func (uv UnionValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !uv.AppliesForVersion(ctx) {
		return nil
	}
	
	var errors []string
	for _, alt := range uv.Alternatives {
		if err := alt.Validate(value, ctx); err == nil {
			return nil // Successfully validated against one alternative
		} else {
			errors = append(errors, err.Error())
		}
	}
	
	return ValidationError{
		Path:    ctx.Path,
		Message: fmt.Sprintf("value does not match any union alternative: %s", strings.Join(errors, "; ")),
	}
}

// LiteralValidator validates literal values (strings, numbers, booleans)
type LiteralValidator struct {
	BaseValidator
	Value interface{}
}

func (lv LiteralValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !lv.AppliesForVersion(ctx) {
		return nil
	}
	
	if !reflect.DeepEqual(value, lv.Value) {
		return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("expected literal value %v, got %v", lv.Value, value)}
	}
	return nil
}

// ReferenceValidator validates references to other types
type ReferenceValidator struct {
	BaseValidator
	TypeName string
}

func (rv ReferenceValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !rv.AppliesForVersion(ctx) {
		return nil
	}
	
	validator, exists := ctx.Definitions[rv.TypeName]
	if !exists {
		return ValidationError{Path: ctx.Path, Message: fmt.Sprintf("undefined type reference: %s", rv.TypeName)}
	}
	
	return validator.Validate(value, ctx)
}

// AttributedValidator wraps another validator with attributes (version constraints)
type AttributedValidator struct {
	BaseValidator
	InnerValidator Validator
	Attributes     map[string]string // attribute name -> value
}

func (av AttributedValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !av.AppliesForVersion(ctx) {
		return nil
	}
	
	// TODO: Handle specific attributes like #[id], #[nbt_path], etc.
	// For now, just validate the inner type
	return av.InnerValidator.Validate(value, ctx)
}

// ConstrainedValidator applies constraints (like ranges) to a base type
type ConstrainedValidator struct {
	BaseValidator
	InnerValidator Validator
	Constraint     Validator // typically a RangeValidator
}

func (cv ConstrainedValidator) Validate(value interface{}, ctx *ValidationContext) error {
	if !cv.AppliesForVersion(ctx) {
		return nil
	}
	
	// First validate the base type
	if err := cv.InnerValidator.Validate(value, ctx); err != nil {
		return err
	}
	
	// Then apply the constraint
	return cv.Constraint.Validate(value, ctx)
}