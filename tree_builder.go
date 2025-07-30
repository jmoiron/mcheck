package main

import "fmt"

// TreeNode represents a node in the expression tree being built
type TreeNode struct {
	Type     string
	Value    interface{}
	Children []*TreeNode
	Parent   *TreeNode
}

func (tn *TreeNode) String() string {
	if tn.Value != nil {
		return fmt.Sprintf("%s(%v)", tn.Type, tn.Value)
	}
	return tn.Type
}

// TreeBuilder builds expression trees during parsing
type TreeBuilder struct {
	Root    *TreeNode
	Current *TreeNode
	Stack   []*TreeNode // Stack of nodes being built
}

func (tb *TreeBuilder) Init() {
	tb.Root = nil
	tb.Current = nil
	tb.Stack = []*TreeNode{}
}

// Push a new node of the given type and make it current
func (tb *TreeBuilder) PushNode(nodeType string) {
	node := &TreeNode{
		Type:     nodeType,
		Children: []*TreeNode{},
	}
	
	if tb.Current != nil {
		node.Parent = tb.Current
		tb.Current.Children = append(tb.Current.Children, node)
	}
	
	tb.Stack = append(tb.Stack, tb.Current)
	tb.Current = node
	
	if tb.Root == nil {
		tb.Root = node
	}
}

// Pop the current node and return to its parent
func (tb *TreeBuilder) PopNode() {
	if len(tb.Stack) > 0 {
		tb.Current = tb.Stack[len(tb.Stack)-1]
		tb.Stack = tb.Stack[:len(tb.Stack)-1]
	} else {
		tb.Current = nil
	}
}

// Add a leaf value to the current node
func (tb *TreeBuilder) AddValue(nodeType string, value interface{}) {
	node := &TreeNode{
		Type:  nodeType,
		Value: value,
	}
	
	if tb.Current != nil {
		node.Parent = tb.Current
		tb.Current.Children = append(tb.Current.Children, node)
	} else {
		tb.Root = node
		tb.Current = node
	}
}

// Get all leaf values of a specific type from current node's children
func (tb *TreeBuilder) GetChildValues(nodeType string) []interface{} {
	if tb.Current == nil {
		return nil
	}
	
	var values []interface{}
	for _, child := range tb.Current.Children {
		if child.Type == nodeType && child.Value != nil {
			values = append(values, child.Value)
		}
	}
	return values
}

// Get first child node of a specific type
func (tb *TreeBuilder) GetChildNode(nodeType string) *TreeNode {
	if tb.Current == nil {
		return nil
	}
	
	for _, child := range tb.Current.Children {
		if child.Type == nodeType {
			return child
		}
	}
	return nil
}

// Print the tree for debugging
func (tb *TreeBuilder) PrintTree() {
	if tb.Root != nil {
		tb.printNode(tb.Root, 0)
	}
}

func (tb *TreeBuilder) printNode(node *TreeNode, indent int) {
	// Debug printing removed for cleaner output
}