package main

import "strings"

// StatementBuilder accumulates parsed mcdoc statements during parsing
type StatementBuilder struct {
	Statements []Statement
	Definitions map[string]Validator
	
	// Expression building stacks
	ExprStack []Expression
	PathSegmentStack []PathSegment
	
	// Tree builder for complex nested structures
	TreeBuilder TreeBuilder
}

// Statement represents a top-level mcdoc statement
type Statement interface {
	StatementType() StatementType
}

// UseStatement represents a use statement with its path
type UseStatement struct {
	Path Path
}

func (us UseStatement) StatementType() StatementType {
	return StatementTypeUse
}

// TypeAliasStatement represents a type alias
type TypeAliasStatement struct {
	Name      Identifier
	Type      Expression
	Validator Validator
}

func (tas TypeAliasStatement) StatementType() StatementType {
	return StatementTypeAlias
}

// StructStatement represents a struct definition
type StructStatement struct {
	Name      Identifier
	Validator Validator
}

func (ss StructStatement) StatementType() StatementType {
	return StatementTypeStruct
}

// EnumStatement represents an enum definition
type EnumStatement struct {
	Name      Identifier
	Validator Validator
}

func (es EnumStatement) StatementType() StatementType {
	return StatementTypeEnum
}

// DispatchStatement represents a dispatch statement
type DispatchStatement struct {
	Path      string // dispatch path like minecraft:loot_function[apply_bonus]
	Target    Expression
	Validator Validator
}

func (ds DispatchStatement) StatementType() StatementType {
	return StatementTypeDispatch
}

type StatementType int

const (
	StatementTypeUse StatementType = iota
	StatementTypeAlias
	StatementTypeStruct
	StatementTypeEnum
	StatementTypeDispatch
)

func (sb *StatementBuilder) Init() {
	sb.Statements = []Statement{}
	sb.Definitions = make(map[string]Validator)
	sb.ExprStack = []Expression{}
	sb.PathSegmentStack = []PathSegment{}
	sb.TreeBuilder.Init()
}

func (sb *StatementBuilder) AddUseStatement(path Path) {
	stmt := UseStatement{Path: path}
	sb.Statements = append(sb.Statements, stmt)
}

func (sb *StatementBuilder) AddUseStatementFromText(pathText string) {
	path := sb.parsePathFromText(pathText)
	sb.AddUseStatement(path)
}

func (sb *StatementBuilder) parsePathFromText(pathText string) Path {
	pathText = strings.TrimSpace(pathText)
	
	// Check if path is absolute (starts with ::)
	isAbsolute := strings.HasPrefix(pathText, "::")
	if isAbsolute {
		pathText = pathText[2:] // Remove leading ::
	}
	
	// Split on :: to get segments
	segmentTexts := strings.Split(pathText, "::")
	segments := make([]PathSegment, len(segmentTexts))
	
	for i, segmentText := range segmentTexts {
		segmentText = strings.TrimSpace(segmentText)
		segments[i] = PathSegment{
			Value:   segmentText,
			IsSuper: segmentText == "super",
		}
	}
	
	return Path{
		Segments:   segments,
		IsAbsolute: isAbsolute,
	}
}

func (sb *StatementBuilder) AddTypeAlias(name Identifier, expr Expression, validator Validator) {
	stmt := TypeAliasStatement{
		Name:      name,
		Type:      expr,
		Validator: validator,
	}
	sb.Statements = append(sb.Statements, stmt)
	sb.Definitions[name.Name] = validator
}

func (sb *StatementBuilder) AddStructDef(name Identifier, validator Validator) {
	stmt := StructStatement{
		Name:      name,
		Validator: validator,
	}
	sb.Statements = append(sb.Statements, stmt)
	sb.Definitions[name.Name] = validator
}

func (sb *StatementBuilder) AddEnumDef(name Identifier, validator Validator) {
	stmt := EnumStatement{
		Name:      name,
		Validator: validator,
	}
	sb.Statements = append(sb.Statements, stmt)
	sb.Definitions[name.Name] = validator
}

func (sb *StatementBuilder) AddDispatchStmt(path string, target Expression, validator Validator) {
	stmt := DispatchStatement{
		Path:      path,
		Target:    target,
		Validator: validator,
	}
	sb.Statements = append(sb.Statements, stmt)
}

// Expression builder methods

func (sb *StatementBuilder) NewPathSegment(value string) PathSegment {
	return PathSegment{
		Value:   value,
		IsSuper: value == "super",
	}
}

func (sb *StatementBuilder) NewPath(isAbsolute bool, segments []PathSegment) Path {
	return Path{
		Segments:   segments,
		IsAbsolute: isAbsolute,
	}
}

func (sb *StatementBuilder) NewIdentifier(name string) Identifier {
	return Identifier{Name: name}
}

func (sb *StatementBuilder) NewStringLiteral(value string) StringLiteral {
	return StringLiteral{Value: value}
}

func (sb *StatementBuilder) NewNumberLiteral(value string) NumberLiteral {
	return NumberLiteral{Value: value}
}

func (sb *StatementBuilder) NewBooleanLiteral(value bool) BooleanLiteral {
	return BooleanLiteral{Value: value}
}

// Stack-based expression building methods (following calculator pattern)

func (sb *StatementBuilder) PushIdentifier(name string) {
	identifier := Identifier{Name: strings.TrimSpace(name)}
	sb.ExprStack = append(sb.ExprStack, identifier)
	
	// Also push as PathSegment for path building
	segment := PathSegment{Value: strings.TrimSpace(name), IsSuper: false}
	sb.PathSegmentStack = append(sb.PathSegmentStack, segment)
}

func (sb *StatementBuilder) PushString(value string) {
	// Remove surrounding quotes
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	stringLit := StringLiteral{Value: value}
	sb.ExprStack = append(sb.ExprStack, stringLit)
}

func (sb *StatementBuilder) PushNumber(value string) {
	numberLit := NumberLiteral{Value: strings.TrimSpace(value)}
	sb.ExprStack = append(sb.ExprStack, numberLit)
}

func (sb *StatementBuilder) PushBoolean(value string) {
	boolValue := strings.TrimSpace(value) == "true"
	boolLit := BooleanLiteral{Value: boolValue}
	sb.ExprStack = append(sb.ExprStack, boolLit)
}

func (sb *StatementBuilder) PushSuperKeyword() {
	segment := PathSegment{Value: "super", IsSuper: true}
	sb.PathSegmentStack = append(sb.PathSegmentStack, segment)
}

func (sb *StatementBuilder) BuildPathFromSegments(hasLeadingDoubleColon bool) {
	// Take all segments from the stack
	segments := make([]PathSegment, len(sb.PathSegmentStack))
	copy(segments, sb.PathSegmentStack)
	sb.PathSegmentStack = sb.PathSegmentStack[:0] // Clear the stack
	
	path := Path{
		Segments:   segments,
		IsAbsolute: hasLeadingDoubleColon,
	}
	
	sb.ExprStack = append(sb.ExprStack, path)
}

func (sb *StatementBuilder) PopPathAndAddUseStatement() {
	if len(sb.ExprStack) == 0 {
		return
	}
	
	// Pop the path from the expression stack
	pathExpr := sb.ExprStack[len(sb.ExprStack)-1]
	sb.ExprStack = sb.ExprStack[:len(sb.ExprStack)-1]
	
	if path, ok := pathExpr.(Path); ok {
		stmt := UseStatement{Path: path}
		sb.Statements = append(sb.Statements, stmt)
	}
}

// Struct building methods using TreeBuilder

func (sb *StatementBuilder) BeginStruct() {
	sb.TreeBuilder.PushNode("struct")
}

func (sb *StatementBuilder) EndStruct() {
	// Convert the tree structure to a StructExpression
	structExpr := sb.buildStructFromTree()
	sb.TreeBuilder.PopNode()
	
	// Push the built struct to the expression stack
	sb.ExprStack = append(sb.ExprStack, structExpr)
}

func (sb *StatementBuilder) BeginField() {
	sb.TreeBuilder.PushNode("field")
}

func (sb *StatementBuilder) EndField() {
	sb.TreeBuilder.PopNode()
}

func (sb *StatementBuilder) MarkFieldOptional() {
	sb.TreeBuilder.AddValue("optional", true)
}

func (sb *StatementBuilder) AddFieldColon() {
	sb.TreeBuilder.AddValue("colon", true)
}

func (sb *StatementBuilder) PopStructAndAddStatement() {
	if len(sb.ExprStack) < 1 {
		return
	}
	
	// Pop the struct expression
	_ = sb.ExprStack[len(sb.ExprStack)-1] // structExpr, will use later
	sb.ExprStack = sb.ExprStack[:len(sb.ExprStack)-1]
	
	// The struct name should be the first identifier pushed (TestStruct)
	// Find it by looking for the first Identifier in the stack
	var nameExpr Expression
	var nameIndex int = -1
	for i, expr := range sb.ExprStack {
		if _, ok := expr.(Identifier); ok {
			nameExpr = expr
			nameIndex = i
			break
		}
	}
	
	if nameIndex == -1 {
		return
	}
	
	// Remove the name from the stack
	sb.ExprStack = append(sb.ExprStack[:nameIndex], sb.ExprStack[nameIndex+1:]...)
	
	if nameIdent, ok := nameExpr.(Identifier); ok {
		// Create a validator placeholder for now
		validator := &PrimitiveValidator{Type: "struct"}
		
		stmt := StructStatement{
			Name:      nameIdent,
			Validator: validator,
		}
		sb.Statements = append(sb.Statements, stmt)
		
		// Make sure Definitions map is initialized
		if sb.Definitions == nil {
			sb.Definitions = make(map[string]Validator)
		}
		sb.Definitions[nameIdent.Name] = validator
	}
}

func (sb *StatementBuilder) buildStructFromTree() Expression {
	// This would build a proper StructExpression from the tree
	// For now, return a simple placeholder
	return Identifier{Name: "StructPlaceholder"}
}

func (sb *StatementBuilder) PrintDebug() {
	// Debug functionality removed for cleaner output
}

// Dispatch statement building methods

func (sb *StatementBuilder) BeginDispatch() {
	// Dispatch parsing placeholder
}

func (sb *StatementBuilder) AddDispatchPath(path string) {
	// Store dispatch path for later use
}

func (sb *StatementBuilder) AddDispatchTarget() {
	// Create a dispatch statement with a placeholder validator
	validator := &PrimitiveValidator{Type: "dispatch"}
	
	// For now, create a basic dispatch statement
	stmt := DispatchStatement{
		Path:      "minecraft:resource", // placeholder
		Target:    Identifier{Name: "dispatch_target"},
		Validator: validator,
	}
	sb.Statements = append(sb.Statements, stmt)
}

// GetDefinitions returns all type definitions from the parsed statements
func (sb *StatementBuilder) GetDefinitions() map[string]Validator {
	return sb.Definitions
}