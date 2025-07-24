//go:generate peg grammar.peg

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

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

type MCDocType int

const (
	MCDocTypeUnknown MCDocType = iota
	MCDocTypeString
	MCDocTypeInt
	MCDocTypeFloat
	MCDocTypeDouble
	MCDocTypeBoolean
	MCDocTypeArray
	MCDocTypeStruct
	MCDocTypeEnum
	MCDocTypeOptional
)

type MCDocField struct {
	Name         string
	Type         MCDocType
	TypeName     string // Store original type string for named type resolution
	Optional     bool
	VersionSince *Version
	VersionUntil *Version
	IsAvailable  bool
	ArrayType    *MCDocField
	StructFields map[string]*MCDocField
	EnumValues   []string
	Constraints  map[string]interface{}
}

type MCDocValidator struct {
	targetVersion Version
	schemaDir     string
}

func NewMCDocValidator(targetVersion Version, schemaDir string) *MCDocValidator {
	return &MCDocValidator{
		targetVersion: targetVersion,
		schemaDir:     schemaDir,
	}
}

func (p *MCDocValidator) determineSchemaPath(jsonPath string) (string, error) {
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
	schemaPathParts := append([]string{p.schemaDir, "java", "data"}, typePath...)
	schemaPath := strings.Join(schemaPathParts, string(os.PathSeparator)) + ".mcdoc"

	return schemaPath, nil
}

type MCDocSchema struct {
	Content string
	Fields  map[string]*MCDocField
	Structs map[string]*MCDocStruct
	Enums   map[string]*MCDocEnum
}

type MCDocStruct struct {
	Name   string
	Fields map[string]*MCDocField
}

type MCDocEnum struct {
	Name   string
	Values map[string]string
}

func (p *MCDocValidator) parseSchema(schemaPath string) (*MCDocSchema, error) {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	schema := &MCDocSchema{
		Content: string(content),
		Fields:  make(map[string]*MCDocField),
		Structs: make(map[string]*MCDocStruct),
		Enums:   make(map[string]*MCDocEnum),
	}

	// Parse the schema content and extract structures
	if err := p.parseSchemaContent(schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema content: %w", err)
	}

	return schema, nil
}

func (p *MCDocValidator) parseSchemaContent(schema *MCDocSchema) error {
	content := schema.Content
	lines := strings.Split(content, "\n")

	// Find the main dispatch structure and parse it precisely
	var mainStruct *MCDocStruct
	var inMainStruct bool
	var nestedBraceLevel int // Track additional nesting within the main struct

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "///") {
			i++
			continue
		}

		// Count braces to track nesting
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		// Parse dispatch (main structure)
		if strings.Contains(line, "dispatch") && strings.Contains(line, "to struct") {
			structName := p.extractStructName(line)
			if structName != "" {
				mainStruct = &MCDocStruct{
					Name:   structName,
					Fields: make(map[string]*MCDocField),
				}
				schema.Structs[structName] = mainStruct
				// If the line contains {, we're already in the struct
				if openBraces > 0 {
					inMainStruct = true
					nestedBraceLevel = 0
				}
				i++
				continue
			}
		}

		// If we haven't found the opening brace yet for dispatch, look for it
		if mainStruct != nil && !inMainStruct && openBraces > 0 {
			inMainStruct = true
			nestedBraceLevel = 0
		}

		// Parse struct fields - only when we're at the top level of the main struct (nestedBraceLevel = 0)
		// Check this BEFORE updating brace levels since the field definition line itself might have braces
		isTopLevelField := mainStruct != nil && inMainStruct && nestedBraceLevel == 0 && strings.Contains(line, ":") && !strings.Contains(line, "dispatch")

		// Update nested brace level when we're in the main struct (after checking for field parsing)
		if inMainStruct {
			nestedBraceLevel += openBraces - closeBraces
		}

		if isTopLevelField {
			// Check if this is a field definition by looking at the line pattern
			if p.isValidFieldLine(line) {
				field := p.parseField(line)
				if field != nil {
					// Check previous lines for version annotations
					field.VersionSince, field.VersionUntil = p.extractVersionFromContext(lines, i)
					field.IsAvailable = p.isFieldAvailableWithConstraints(field.VersionSince, field.VersionUntil)
					mainStruct.Fields[field.Name] = field
				}
			}
		}

		// Handle spread operator (...struct) at top level of main struct
		// Check for spread operator regardless of brace level since it can appear anywhere in the main struct
		if mainStruct != nil && inMainStruct && strings.Contains(line, "...struct") && !strings.Contains(line, ":") {
			p.parseSpreadStruct(line, mainStruct, schema, lines, i)
		}

		// End of main struct
		if inMainStruct && nestedBraceLevel < 0 {
			inMainStruct = false
		}

		i++
	}

	// Parse other struct and enum definitions
	p.parseOtherStructsAndEnums(schema, content)

	// Copy main struct fields to schema fields for backward compatibility
	if mainStruct != nil {
		for name, field := range mainStruct.Fields {
			schema.Fields[name] = field
		}
	}

	return nil
}

func (p *MCDocValidator) parseSpreadStruct(line string, mainStruct *MCDocStruct, schema *MCDocSchema, lines []string, currentIndex int) {
	// Extract the struct name from the spread line like "...struct ModernNoiseGeneratorSettings {"
	structName := p.extractStructName(line)
	if structName == "" {
		return
	}

	// Get version constraints for the spread operator
	spreadVersionSince, spreadVersionUntil := p.extractVersionFromContext(lines, currentIndex)

	// Find the struct definition and parse its fields inline
	braceLevel := 0
	foundStart := false

	// Look for the opening brace of the spread struct
	for i := currentIndex; i < len(lines); i++ {
		currentLine := strings.TrimSpace(lines[i])
		if currentLine == "" || strings.HasPrefix(currentLine, "//") {
			continue
		}

		braceLevel += strings.Count(currentLine, "{") - strings.Count(currentLine, "}")
		
		if !foundStart && strings.Contains(currentLine, "{") {
			foundStart = true
		}

		// If we've closed all braces, we've found the end of the spread struct
		if foundStart && braceLevel <= 0 {
			break
		}

		// Parse fields within the spread struct
		if foundStart && braceLevel > 0 && strings.Contains(currentLine, ":") && p.isValidFieldLine(currentLine) {
			field := p.parseField(currentLine)
			if field != nil {
				// Check for field-level version constraints
				field.VersionSince, field.VersionUntil = p.extractVersionFromContext(lines, i)
				
				// Apply spread-level version constraints as well
				if spreadVersionSince != nil {
					if field.VersionSince == nil || field.VersionSince.Compare(*spreadVersionSince) < 0 {
						field.VersionSince = spreadVersionSince
					}
				}
				if spreadVersionUntil != nil {
					if field.VersionUntil == nil || field.VersionUntil.Compare(*spreadVersionUntil) > 0 {
						field.VersionUntil = spreadVersionUntil
					}
				}
				
				field.IsAvailable = p.isFieldAvailableWithConstraints(field.VersionSince, field.VersionUntil)
				mainStruct.Fields[field.Name] = field
			}
		}
	}
}

func (p *MCDocValidator) isValidFieldLine(line string) bool {
	// Remove version annotations for analysis
	cleanLine := regexp.MustCompile(`#\[[^\]]+\]`).ReplaceAllString(line, "")
	cleanLine = strings.TrimSpace(cleanLine)

	// Skip lines that don't contain proper field definitions
	if !strings.Contains(cleanLine, ":") {
		return false
	}

	// Skip lines that start with [ (these are type definitions inside structs)
	if strings.HasPrefix(cleanLine, "[") {
		return false
	}

	// Skip lines that are complex type definitions
	if strings.Contains(cleanLine, "|") {
		return false
	}

	return true
}

func (p *MCDocValidator) parseOtherStructsAndEnums(schema *MCDocSchema, content string) {
	lines := strings.Split(content, "\n")
	var currentStruct *MCDocStruct
	var currentEnum *MCDocEnum
	var braceLevel int // Track nesting level within the current struct

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "///") {
			continue
		}

		// Track brace levels
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		braceLevel += openBraces - closeBraces

		// Parse standalone struct definitions
		if strings.HasPrefix(line, "struct ") {
			structName := p.extractStructName(line)
			if structName != "" {
				if structName == "BiomeEffects" {
					fmt.Printf("DEBUG: Found BiomeEffects struct definition\n")
				}
				currentStruct = &MCDocStruct{
					Name:   structName,
					Fields: make(map[string]*MCDocField),
				}
				schema.Structs[structName] = currentStruct
				braceLevel = 0 // Reset brace level for new struct
				continue
			}
		}

		// Parse enum definitions
		if strings.HasPrefix(line, "enum") {
			enumName := p.extractEnumName(line)
			if enumName != "" {
				currentEnum = &MCDocEnum{
					Name:   enumName,
					Values: make(map[string]string),
				}
				schema.Enums[enumName] = currentEnum
				continue
			}
		}

		// Parse struct fields - try to catch basic fields
		if currentStruct != nil && strings.Contains(line, ":") && p.isValidFieldLine(line) {
			// Only parse simple fields, skip nested structs
			if strings.Contains(line, "struct ") {
				continue // Skip nested struct definitions
			}
			if currentStruct.Name == "BiomeEffects" {
				fmt.Printf("DEBUG: Parsing BiomeEffects field: %s (braceLevel: %d)\n", line, braceLevel)
			}
			field := p.parseField(line)
			if field != nil {
				field.VersionSince, field.VersionUntil = p.extractVersionFromContext(lines, i)
				field.IsAvailable = p.isFieldAvailableWithConstraints(field.VersionSince, field.VersionUntil)
				currentStruct.Fields[field.Name] = field
				if currentStruct.Name == "BiomeEffects" {
					fmt.Printf("DEBUG: Added BiomeEffects field: %s (type: %d, available: %t)\n", field.Name, field.Type, field.IsAvailable)
				}
			}
		}

		// Parse enum values
		if currentEnum != nil && strings.Contains(line, "=") {
			key, value := p.parseEnumValue(line)
			if key != "" {
				currentEnum.Values[key] = value
			}
		}

		// End of struct/enum
		if line == "}" && braceLevel == 0 {
			if currentEnum != nil {
				currentEnum = nil
			}
			if currentStruct != nil {
				if currentStruct.Name == "BiomeEffects" {
					fmt.Printf("DEBUG: Finished parsing BiomeEffects with %d fields\n", len(currentStruct.Fields))
				}
				currentStruct = nil
			}
		}
	}
}

func (p *MCDocValidator) extractStructName(line string) string {
	// Extract struct name from lines like "struct SpawnerData {" or "dispatch ... to struct Biome {"
	re := regexp.MustCompile(`struct\s+(\w+)\s*\{`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *MCDocValidator) extractEnumName(line string) string {
	// Extract enum name from lines like "enum(string) BiomeCategory {"
	re := regexp.MustCompile(`enum\s*\([^)]*\)\s*(\w+)\s*\{`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *MCDocValidator) parseField(line string) *MCDocField {
	// Parse field definitions like "temperature: float" or "spawners: struct {"

	// Remove version annotations for parsing
	cleanLine := regexp.MustCompile(`#\[[^\]]+\]`).ReplaceAllString(line, "")
	cleanLine = strings.TrimSpace(cleanLine)

	if !strings.Contains(cleanLine, ":") {
		return nil
	}

	parts := strings.SplitN(cleanLine, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	fieldName := strings.TrimSpace(parts[0])
	fieldType := strings.TrimSpace(parts[1])

	// Remove trailing comma
	fieldType = strings.TrimSuffix(fieldType, ",")

	// Check if optional
	optional := strings.HasSuffix(fieldName, "?")
	if optional {
		fieldName = strings.TrimSuffix(fieldName, "?")
	}

	// Skip lines that start with /// (doc comments)
	if strings.HasPrefix(strings.TrimSpace(line), "///") {
		return nil
	}

	// Skip empty field names
	if fieldName == "" {
		return nil
	}

	field := &MCDocField{
		Name:     fieldName,
		Optional: optional,
		Type:     p.parseType(fieldType),
		TypeName: fieldType, // Store original type string
	}

	// Debug for complex fields
	if fieldName == "carvers" || fieldName == "features" || fieldName == "effects" {
		fmt.Printf("DEBUG %s: fieldType='%s', parsedType=%d, typeName='%s'\n", fieldName, fieldType, field.Type, field.TypeName)
	}

	// Parse version constraints
	field.VersionSince = p.extractVersionSince(line)
	field.VersionUntil = p.extractVersionUntil(line)

	return field
}

func (p *MCDocValidator) parseEnumValue(line string) (string, string) {
	// Parse enum values like "Beach = "beach","
	parts := strings.Split(line, "=")
	if len(parts) != 2 {
		return "", ""
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	value = strings.Trim(value, `",`)

	return key, value
}

func (p *MCDocValidator) parseType(typeStr string) MCDocType {
	typeStr = strings.TrimSpace(typeStr)

	switch {
	case typeStr == "string":
		return MCDocTypeString
	case typeStr == "int":
		return MCDocTypeInt
	case typeStr == "float":
		return MCDocTypeFloat
	case typeStr == "double":
		return MCDocTypeDouble
	case typeStr == "boolean":
		return MCDocTypeBoolean
	case strings.HasPrefix(typeStr, "[") && strings.HasSuffix(typeStr, "]"):
		return MCDocTypeArray
	case strings.HasPrefix(typeStr, "struct"):
		return MCDocTypeStruct
	case strings.Contains(typeStr, "{"):
		return MCDocTypeStruct
	case strings.HasPrefix(typeStr, "("):
		// Handle complex union types like "([ConfiguredFeatureRef] | ...)" or "(struct ... | ...)"
		// If the union contains "struct", treat as struct
		if strings.Contains(typeStr, "struct") {
			return MCDocTypeStruct
		}
		// For incomplete types like just "(", we can't determine the type reliably
		// Return Unknown so validation can be more flexible
		if typeStr == "(" {
			return MCDocTypeUnknown
		}
		// Otherwise treat as array for complex array types
		return MCDocTypeArray
	default:
		// Handle complex types like BiomeCategory, BiomeEffects, etc.
		// For now, treat them as struct types
		return MCDocTypeStruct
	}
}

type VersionRequirement struct {
	Since *Version
	Until *Version
}

// parseVersionRequirement determines if a line contains version requirements and extracts them
func parseVersionRequirement(line string) *VersionRequirement {
	req := &VersionRequirement{}

	// Check for #[since="version"] pattern
	sinceRe := regexp.MustCompile(`#\[since="([^"]+)"\]`)
	if matches := sinceRe.FindStringSubmatch(line); len(matches) > 1 {
		if version, err := parseVersion(matches[1]); err == nil {
			req.Since = &version
		}
	}

	// Check for #[until="version"] pattern
	untilRe := regexp.MustCompile(`#\[until="([^"]+)"\]`)
	if matches := untilRe.FindStringSubmatch(line); len(matches) > 1 {
		if version, err := parseVersion(matches[1]); err == nil {
			req.Until = &version
		}
	}

	// Return nil if no version requirements found
	if req.Since == nil && req.Until == nil {
		return nil
	}

	return req
}

func (p *MCDocValidator) extractVersionSince(line string) *Version {
	req := parseVersionRequirement(line)
	if req != nil {
		return req.Since
	}
	return nil
}

func (p *MCDocValidator) extractVersionUntil(line string) *Version {
	req := parseVersionRequirement(line)
	if req != nil {
		return req.Until
	}
	return nil
}

func (p *MCDocValidator) extractVersionFromContext(lines []string, currentIndex int) (*Version, *Version) {
	var sinceVersion *Version
	var untilVersion *Version

	// Check the current line and a few lines before for version annotations
	for j := currentIndex; j >= 0 && j > currentIndex-3; j-- {
		line := strings.TrimSpace(lines[j])

		if since := p.extractVersionSince(line); since != nil {
			sinceVersion = since
		}
		if until := p.extractVersionUntil(line); until != nil {
			untilVersion = until
		}

		// Stop looking if we hit another field definition or structural element (but not doc comments)
		if j < currentIndex && strings.Contains(line, ":") && !strings.HasPrefix(line, "///") {
			break
		}
	}

	return sinceVersion, untilVersion
}

func (p *MCDocValidator) isFieldAvailableWithConstraints(sinceVersion *Version, untilVersion *Version) bool {
	if sinceVersion != nil {
		if p.targetVersion.Compare(*sinceVersion) < 0 {
			return false
		}
	}

	if untilVersion != nil {
		if p.targetVersion.Compare(*untilVersion) > 0 {
			return false
		}
	}

	return true
}

func (p *MCDocValidator) isFieldAvailable(line string) bool {
	// Extract version constraints from the line
	sinceVersion := p.extractVersionSince(line)
	untilVersion := p.extractVersionUntil(line)

	return p.isFieldAvailableWithConstraints(sinceVersion, untilVersion)
}

func (p *MCDocValidator) validateJSON(jsonPath string) error {
	// Determine the schema file to use
	schemaPath, err := p.determineSchemaPath(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to determine schema path: %w", err)
	}

	// Check if schema file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return fmt.Errorf("schema file not found: %s", schemaPath)
	}

	// Parse the schema
	schema, err := p.parseSchema(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Read and parse the JSON file
	jsonContent, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(jsonContent, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Perform comprehensive validation
	fmt.Printf("Validating %s against schema %s for version %s\n", jsonPath, schemaPath, p.targetVersion.String())

	// Find the main struct for validation (should be the dispatch target)
	var mainStruct *MCDocStruct
	for _, s := range schema.Structs {
		if len(s.Fields) > 0 {
			mainStruct = s
			break
		}
	}

	if mainStruct == nil {
		return fmt.Errorf("no main struct found in schema")
	}

	// Validate against the main struct
	issues := p.validateStruct(jsonData, mainStruct, schema, "")

	if len(issues) > 0 {
		fmt.Println("Validation issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		return fmt.Errorf("validation failed with %d issues", len(issues))
	}

	fmt.Println("Validation successful!")
	return nil
}

func (p *MCDocValidator) validateStruct(jsonData map[string]interface{}, structDef *MCDocStruct, schema *MCDocSchema, path string) []string {
	var issues []string

	// Check required fields
	for fieldName, field := range structDef.Fields {
		if !field.IsAvailable {
			continue // Skip fields not available in this version
		}

		fieldPath := path
		if fieldPath != "" {
			fieldPath += "."
		}
		fieldPath += fieldName

		jsonValue, exists := jsonData[fieldName]

		if !exists {
			if !field.Optional {
				issues = append(issues, fmt.Sprintf("Required field '%s' is missing", fieldPath))
			}
			continue
		}

		// Validate field type and value
		fieldIssues := p.validateField(jsonValue, field, schema, fieldPath)
		issues = append(issues, fieldIssues...)
	}

	// Check for unknown fields
	for fieldName := range jsonData {
		if _, exists := structDef.Fields[fieldName]; !exists {
			fieldPath := path
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += fieldName
			issues = append(issues, fmt.Sprintf("Unknown field '%s'", fieldPath))
		}
	}

	return issues
}

func (p *MCDocValidator) validateField(jsonValue interface{}, field *MCDocField, schema *MCDocSchema, path string) []string {
	var issues []string

	switch field.Type {
	case MCDocTypeUnknown:
		// For unknown types (complex union types), accept any valid JSON value
		// This is for cases where we can't determine the exact type from incomplete mcdoc syntax
		return issues // No validation errors for unknown types
	case MCDocTypeString:
		if _, ok := jsonValue.(string); !ok {
			issues = append(issues, fmt.Sprintf("Field '%s' should be a string, got %T", path, jsonValue))
		}
	case MCDocTypeInt:
		switch v := jsonValue.(type) {
		case float64:
			if v != float64(int(v)) {
				issues = append(issues, fmt.Sprintf("Field '%s' should be an integer, got float %v", path, v))
			}
		case int:
			// OK
		default:
			issues = append(issues, fmt.Sprintf("Field '%s' should be an integer, got %T", path, jsonValue))
		}
	case MCDocTypeFloat, MCDocTypeDouble:
		if _, ok := jsonValue.(float64); !ok {
			if _, ok := jsonValue.(int); !ok {
				issues = append(issues, fmt.Sprintf("Field '%s' should be a number, got %T", path, jsonValue))
			}
		}
	case MCDocTypeBoolean:
		if _, ok := jsonValue.(bool); !ok {
			issues = append(issues, fmt.Sprintf("Field '%s' should be a boolean, got %T", path, jsonValue))
		}
	case MCDocTypeArray:
		if arr, ok := jsonValue.([]interface{}); ok {
			// Validate array elements if we have array type info
			for i, item := range arr {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				if field.ArrayType != nil {
					itemIssues := p.validateField(item, field.ArrayType, schema, itemPath)
					issues = append(issues, itemIssues...)
				}
			}
		} else {
			issues = append(issues, fmt.Sprintf("Field '%s' should be an array, got %T", path, jsonValue))
		}
	case MCDocTypeStruct:
		if structData, ok := jsonValue.(map[string]interface{}); ok {
			// Find the struct definition and validate
			if field.StructFields != nil {
				// Use embedded struct fields
				tempStruct := &MCDocStruct{Fields: field.StructFields}
				structIssues := p.validateStruct(structData, tempStruct, schema, path)
				issues = append(issues, structIssues...)
			} else {
				// Try to resolve named struct type from schema
				structName := p.getStructNameFromField(field)
				if structName != "" {
					// Debug output for effects field
					if field.Name == "effects" {
						fmt.Printf("DEBUG: Looking for struct '%s'\n", structName)
						fmt.Printf("DEBUG: Available structs: ")
						for name := range schema.Structs {
							fmt.Printf("%s ", name)
						}
						fmt.Printf("\n")
					}
					
					if structDef, exists := schema.Structs[structName]; exists {
						if field.Name == "effects" {
							fmt.Printf("DEBUG: Found struct '%s' with %d fields\n", structName, len(structDef.Fields))
							fmt.Printf("DEBUG: BiomeEffects fields: ")
							for fname := range structDef.Fields {
								fmt.Printf("%s ", fname)
							}
							fmt.Printf("\n")
						}
						structIssues := p.validateStruct(structData, structDef, schema, path)
						issues = append(issues, structIssues...)
					} else {
						// Unknown struct type, skip validation
						// This allows for forward compatibility with unknown types
						if field.Name == "effects" {
							fmt.Printf("DEBUG: Struct '%s' not found\n", structName)
						}
					}
				}
			}
		} else {
			issues = append(issues, fmt.Sprintf("Field '%s' should be an object, got %T", path, jsonValue))
		}
	}

	return issues
}

func (p *MCDocValidator) getStructNameFromField(field *MCDocField) string {
	// Use the stored TypeName for named struct lookups
	typeName := strings.TrimSpace(field.TypeName)
	
	// Skip inline struct definitions
	if strings.HasPrefix(typeName, "struct {") {
		return ""
	}
	
	// TEMPORARY: Re-enable BiomeEffects to test the mechanism
	// if typeName == "BiomeEffects" {
	// 	return ""
	// }
	
	// For simple named types like "SpawnerData", return as-is
	if typeName != "" && !strings.Contains(typeName, " ") && !strings.Contains(typeName, "(") {
		return typeName
	}
	
	return ""
}

func main() {
	var (
		version   string
		schemaDir string
	)

	rootCmd := &cobra.Command{
		Use:   "mcheck <json-file>",
		Short: "Validate Minecraft datapack JSON files against mcdoc schemas",
		Long: `mcheck is a tool for validating Minecraft datapack JSON files against
mcdoc schemas with version-specific constraints.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonPath := args[0]

			// Parse the target version
			targetVersion, err := parseVersion(version)
			if err != nil {
				return fmt.Errorf("invalid version format: %w", err)
			}

			// Find schema directory if not provided
			if schemaDir == "" {
				// Look for vanilla-mcdoc directory
				if _, err := os.Stat("vanilla-mcdoc"); err == nil {
					schemaDir = "vanilla-mcdoc"
				} else {
					return fmt.Errorf("schema directory not found, please specify with --schema-dir")
				}
			}

			// Create parser and validate
			parser := NewMCDocValidator(targetVersion, schemaDir)
			return parser.validateJSON(jsonPath)
		},
	}

	rootCmd.Flags().StringVarP(&version, "version", "v", "1.20.1", "Target Minecraft version")
	rootCmd.Flags().StringVarP(&schemaDir, "schema-dir", "s", "", "Path to vanilla-mcdoc directory")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
