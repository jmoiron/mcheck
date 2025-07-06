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
	if len(parts) != 3 {
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
	
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
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
	Optional     bool
	VersionSince *Version
	VersionUntil *Version
	IsAvailable  bool
	ArrayType    *MCDocField
	StructFields map[string]*MCDocField
	EnumValues   []string
	Constraints  map[string]interface{}
}

type MCDocParser struct {
	targetVersion Version
	schemaDir     string
}

func NewMCDocParser(targetVersion Version, schemaDir string) *MCDocParser {
	return &MCDocParser{
		targetVersion: targetVersion,
		schemaDir:     schemaDir,
	}
}

func (p *MCDocParser) determineSchemaPath(jsonPath string) (string, error) {
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

func (p *MCDocParser) parseSchema(schemaPath string) (*MCDocSchema, error) {
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

func (p *MCDocParser) parseSchemaContent(schema *MCDocSchema) error {
	content := schema.Content
	lines := strings.Split(content, "\n")

	// Find the main dispatch structure
	var mainStruct *MCDocStruct
	var currentStruct *MCDocStruct
	var currentEnum *MCDocEnum
	var braceLevel int
	
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "///") {
			i++
			continue
		}

		// Count braces to track nesting
		braceLevel += strings.Count(line, "{") - strings.Count(line, "}")

		// Parse dispatch (main structure)
		if strings.Contains(line, "dispatch") && strings.Contains(line, "to struct") {
			structName := p.extractStructName(line)
			if structName != "" {
				mainStruct = &MCDocStruct{
					Name:   structName,
					Fields: make(map[string]*MCDocField),
				}
				schema.Structs[structName] = mainStruct
				currentStruct = mainStruct
				i++
				continue
			}
		}

		// Parse struct definition
		if strings.HasPrefix(line, "struct ") {
			structName := p.extractStructName(line)
			if structName != "" {
				newStruct := &MCDocStruct{
					Name:   structName,
					Fields: make(map[string]*MCDocField),
				}
				schema.Structs[structName] = newStruct
				// Only switch context if this is a top-level struct
				if braceLevel <= 1 {
					currentStruct = newStruct
				}
				i++
				continue
			}
		}

		// Parse enum definition
		if strings.HasPrefix(line, "enum") {
			enumName := p.extractEnumName(line)
			if enumName != "" {
				currentEnum = &MCDocEnum{
					Name:   enumName,
					Values: make(map[string]string),
				}
				schema.Enums[enumName] = currentEnum
				i++
				continue
			}
		}

		// Parse struct fields
		if currentStruct != nil && strings.Contains(line, ":") && !strings.Contains(line, "dispatch") {
			field := p.parseField(line)
			if field != nil {
				field.IsAvailable = p.isFieldAvailable(line)
				currentStruct.Fields[field.Name] = field
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
		if line == "}" {
			if currentEnum != nil {
				currentEnum = nil
			}
		}

		i++
	}

	// Copy main struct fields to schema fields for backward compatibility
	if mainStruct != nil {
		for name, field := range mainStruct.Fields {
			schema.Fields[name] = field
		}
	}

	return nil
}

func (p *MCDocParser) extractStructName(line string) string {
	// Extract struct name from lines like "struct SpawnerData {" or "dispatch ... to struct Biome {"
	re := regexp.MustCompile(`struct\s+(\w+)\s*\{`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *MCDocParser) extractEnumName(line string) string {
	// Extract enum name from lines like "enum(string) BiomeCategory {"
	re := regexp.MustCompile(`enum\s*\([^)]*\)\s*(\w+)\s*\{`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *MCDocParser) parseField(line string) *MCDocField {
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
	
	field := &MCDocField{
		Name:     fieldName,
		Optional: optional,
		Type:     p.parseType(fieldType),
	}
	
	// Parse version constraints
	field.VersionSince = p.extractVersionSince(line)
	field.VersionUntil = p.extractVersionUntil(line)
	
	return field
}

func (p *MCDocParser) parseEnumValue(line string) (string, string) {
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

func (p *MCDocParser) parseType(typeStr string) MCDocType {
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
	default:
		return MCDocTypeUnknown
	}
}

func (p *MCDocParser) extractVersionSince(line string) *Version {
	re := regexp.MustCompile(`#\[since="([^"]+)"\]`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		if version, err := parseVersion(matches[1]); err == nil {
			return &version
		}
	}
	return nil
}

func (p *MCDocParser) extractVersionUntil(line string) *Version {
	re := regexp.MustCompile(`#\[until="([^"]+)"\]`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		if version, err := parseVersion(matches[1]); err == nil {
			return &version
		}
	}
	return nil
}

func (p *MCDocParser) isFieldAvailable(line string) bool {
	// Extract version constraints from the line
	sinceVersion := p.extractVersionSince(line)
	untilVersion := p.extractVersionUntil(line)
	
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

func (p *MCDocParser) validateJSON(jsonPath string) error {
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
	
	// Find the main struct for validation
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

func (p *MCDocParser) validateStruct(jsonData map[string]interface{}, structDef *MCDocStruct, schema *MCDocSchema, path string) []string {
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

func (p *MCDocParser) validateField(jsonValue interface{}, field *MCDocField, schema *MCDocSchema, path string) []string {
	var issues []string
	
	switch field.Type {
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
			}
		} else {
			issues = append(issues, fmt.Sprintf("Field '%s' should be an object, got %T", path, jsonValue))
		}
	}
	
	return issues
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
			parser := NewMCDocParser(targetVersion, schemaDir)
			return parser.validateJSON(jsonPath)
		},
	}
	
	rootCmd.Flags().StringVarP(&version, "version", "v", "1.20.1", "Target Minecraft version")
	rootCmd.Flags().StringVarP(&schemaDir, "schema-dir", "s", "", "Path to vanilla-mcdoc directory")
	
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}