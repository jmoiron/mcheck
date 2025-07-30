package main

// Expression represents a value in the mcdoc AST
type Expression interface {
	String() string
}

// PathSegment represents a single segment in a path (identifier or 'super')
type PathSegment struct {
	Value string
	IsSuper bool
}

func (ps PathSegment) String() string {
	return ps.Value
}

// Path represents a :: separated path like super::test::Type or ::java::util::List
type Path struct {
	Segments []PathSegment
	IsAbsolute bool // starts with ::
}

func (p Path) String() string {
	result := ""
	if p.IsAbsolute {
		result = "::"
	}
	for i, segment := range p.Segments {
		if i > 0 {
			result += "::"
		}
		result += segment.String()
	}
	return result
}

// Identifier represents a simple identifier
type Identifier struct {
	Name string
}

func (i Identifier) String() string {
	return i.Name
}

// String represents a string literal
type StringLiteral struct {
	Value string
}

func (s StringLiteral) String() string {
	return "\"" + s.Value + "\""
}

// Number represents a numeric literal
type NumberLiteral struct {
	Value string
}

func (n NumberLiteral) String() string {
	return n.Value
}

// Boolean represents a boolean literal
type BooleanLiteral struct {
	Value bool
}

func (b BooleanLiteral) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// StructExpression represents a struct definition
type StructExpression struct {
	Name   *Identifier // optional name for inline structs
	Fields []FieldExpression
}

func (s StructExpression) String() string {
	result := "struct"
	if s.Name != nil {
		result += " " + s.Name.Name
	}
	result += " { "
	for i, field := range s.Fields {
		if i > 0 {
			result += ", "
		}
		result += field.String()
	}
	result += " }"
	return result
}

// FieldExpression represents a field in a struct
type FieldExpression struct {
	Name     Identifier
	Type     Expression
	Optional bool
}

func (f FieldExpression) String() string {
	result := f.Name.Name
	if f.Optional {
		result += "?"
	}
	result += ": " + f.Type.String()
	return result
}