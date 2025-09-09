package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PEGMCDocValidator uses the PEG parser for validation
type PEGMCDocValidator struct {
	targetVersion Version
	schemaDir     string
}

func NewPEGMCDocValidator(targetVersion Version, schemaDir string) *PEGMCDocValidator {
	return &PEGMCDocValidator{
		targetVersion: targetVersion,
		schemaDir:     schemaDir,
	}
}

func (v *PEGMCDocValidator) ValidateJSON(jsonPath string) error {
	// Determine the schema file to use
	schemaPath, err := v.determineSchemaPath(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to determine schema path: %w", err)
	}

	// Check if schema file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return fmt.Errorf("schema file not found: %s", schemaPath)
	}

	// Validating JSON against schema

	// Parse the mcdoc schema using our PEG parser
	statements, _, err := v.parseSchemaWithPEG(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to parse schema with PEG: %w", err)
	}

	// Schema parsed successfully

	// Read and parse the JSON file
	jsonContent, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(jsonContent, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert parsed statements to proper validators
	converter := NewSchemaConverter(v.targetVersion, statements)
	validatorMap, err := converter.ConvertToValidators()
	if err != nil {
		return fmt.Errorf("failed to convert statements to validators: %w", err)
	}

	// Create validation context
	ctx := &ValidationContext{
		Version:     v.targetVersion,
		Path:        []string{},
		Definitions: validatorMap,
	}

	// Find the main validator
	mainValidator := converter.GetMainValidator()
	if mainValidator == nil {
		// If no specific main validator found, create a basic struct validator
		mainValidator = converter.CreateBasicStructValidator()
	}

	// Perform actual JSON validation against the parsed schema
	if err := mainValidator.Validate(jsonData, ctx); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func (v *PEGMCDocValidator) parseSchemaWithPEG(schemaPath string) ([]Statement, map[string]Validator, error) {
	// Read the schema file
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Create PEG parser
	parser := &MCDocParser{
		Buffer: string(content),
		Pretty: true,
	}

	// Initialize parser
	err = parser.Init()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize parser: %w", err)
	}

	// Parse the content
	err = parser.Parse()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse mcdoc: %w", err)
	}

	// Execute actions to build statements
	parser.Execute()

	// Return the parsed statements and definitions
	return parser.Statements, parser.GetDefinitions(), nil
}

func (v *PEGMCDocValidator) findMainValidator(statements []Statement, definitions map[string]Validator) Validator {
	// Look for dispatch statements first (they define the main entry point)
	for _, stmt := range statements {
		if dispatchStmt, ok := stmt.(DispatchStatement); ok {
			return dispatchStmt.Validator
		}
	}

	// If no dispatch statement, look for struct definitions
	for _, stmt := range statements {
		if structStmt, ok := stmt.(StructStatement); ok {
			return structStmt.Validator
		}
	}

	// If no struct statement, look for type aliases
	for _, stmt := range statements {
		if aliasStmt, ok := stmt.(TypeAliasStatement); ok {
			return aliasStmt.Validator
		}
	}

	// Return nil if no suitable validator found
	return nil
}

func (v *PEGMCDocValidator) determineSchemaPath(jsonPath string) (string, error) {
	// Extract the relative path from the datapack structure
	// Expected structure: data/(optional namespace)/type/subtype/file.json
	parts := strings.Split(filepath.Clean(jsonPath), string(os.PathSeparator))

	// Find the "data" directory and extract the type path
	dataIndex := -1
	for i, part := range parts {
		if part == "data" {
			dataIndex = i
			break
		}
	}

	if dataIndex == -1 || dataIndex+2 >= len(parts) {
		return "", fmt.Errorf("invalid datapack structure: %s", jsonPath)
	}

	// Get the path from after "data" to the file
	// For data/worldgen/noise_settings/foo.json -> worldgen/noise_settings
	// For data/namespace/worldgen/noise_settings/foo.json -> worldgen/noise_settings (skip namespace)
	typePath := parts[dataIndex+1:]

	// Remove the filename to get the directory structure
	if len(typePath) > 0 {
		typePath = typePath[:len(typePath)-1]
	}

	if len(typePath) == 0 {
		return "", fmt.Errorf("invalid datapack structure: %s", jsonPath)
	}

	// If the first part looks like a namespace (not a known type), skip it
	knownTypes := []string{"worldgen", "advancement", "recipe", "loot_table", "structure", "dimension", "dimension_type", "biome", "configured_carver", "configured_feature", "placed_feature", "processor_list", "template_pool", "structure_set", "noise_settings", "density_function", "multi_noise_biome_source_parameter_list", "chat_type", "damage_type", "trim_pattern", "trim_material", "wolf_variant", "painting_variant", "jukebox_song", "banner_pattern", "enchantment", "item_modifier", "predicate", "tag", "function", "gametest"}

	if len(typePath) > 1 {
		firstPart := typePath[0]
		isKnownType := false
		for _, knownType := range knownTypes {
			if firstPart == knownType {
				isKnownType = true
				break
			}
		}
		// If the first part is not a known type, assume it's a namespace and skip it
		if !isKnownType {
			typePath = typePath[1:]
		}
	}

	if len(typePath) == 0 {
		return "", fmt.Errorf("invalid datapack structure: %s", jsonPath)
	}

	// Build the schema path: vanilla-mcdoc/java/data/worldgen/noise_settings.mcdoc
	schemaPathParts := append([]string{v.schemaDir, "java", "data"}, typePath...)
	schemaPath := strings.Join(schemaPathParts, string(os.PathSeparator)) + ".mcdoc"

	return schemaPath, nil
}