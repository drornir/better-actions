package expr

import (
	"fmt"
	"strings"
)

// Taken from actionlint/expr_ast.go

// Node is a node of expression syntax tree. To know the syntax, see
// https://docs.github.com/en/actions/learn-github-actions/expressions
type Node interface {
	// Token returns the first token of the node. This method is useful to get position of this node.
	Token() *Token
}

// Variable

// VariableNode is node for variable access.
type VariableNode struct {
	// Name is name of the variable
	Name string
	tok  *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *VariableNode) Token() *Token {
	return n.tok
}

// Literals

// NullNode is node for null literal.
type NullNode struct {
	tok *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *NullNode) Token() *Token {
	return n.tok
}

// BoolNode is node for boolean literal, true or false.
type BoolNode struct {
	// Value is value of the boolean literal.
	Value bool
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *BoolNode) Token() *Token {
	return n.tok
}

// IntNode is node for integer literal.
type IntNode struct {
	// Value is value of the integer literal.
	Value int64
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *IntNode) Token() *Token {
	return n.tok
}

// FloatNode is node for float literal.
type FloatNode struct {
	// Value is value of the float literal.
	Value float64
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *FloatNode) Token() *Token {
	return n.tok
}

// StringNode is node for string literal.
type StringNode struct {
	// Value is value of the string literal. Escapes are resolved and quotes at both edges are
	// removed.
	Value string
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *StringNode) Token() *Token {
	return n.tok
}

// Operators

// ObjectDerefNode represents property dereference of object like 'foo.bar'.
type ObjectDerefNode struct {
	// Receiver is an expression at receiver of property dereference.
	Receiver Node
	// Property is a name of property to access.
	Property string
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n ObjectDerefNode) Token() *Token {
	return n.Receiver.Token()
}

// ArrayDerefNode represents elements dereference of arrays like '*' in 'foo.bar.*.piyo'.
type ArrayDerefNode struct {
	// Receiver is an expression at receiver of array element dereference.
	Receiver Node
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n ArrayDerefNode) Token() *Token {
	return n.Receiver.Token()
}

// IndexAccessNode is node for index access, which represents dynamic object property access or
// array index access.
type IndexAccessNode struct {
	// Operand is an expression at operand of index access, which should be array or object.
	Operand Node
	// Index is an expression at index, which should be integer or string.
	Index Node
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *IndexAccessNode) Token() *Token {
	return n.Operand.Token()
}

// Note: Currently only ! is a logical unary operator

// NotOpNode is node for unary ! operator.
type NotOpNode struct {
	// Operand is an expression at operand of ! operator.
	Operand Node
	tok     *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *NotOpNode) Token() *Token {
	return n.tok
}

// CompareOpNodeKind is a kind of compare operators; ==, !=, <, <=, >, >=.
type CompareOpNodeKind int

const (
	// CompareOpNodeKindInvalid is invalid and initial value of CompareOpNodeKind values.
	CompareOpNodeKindInvalid CompareOpNodeKind = iota
	// CompareOpNodeKindLess is kind for < operator.
	CompareOpNodeKindLess
	// CompareOpNodeKindLessEq is kind for <= operator.
	CompareOpNodeKindLessEq
	// CompareOpNodeKindGreater is kind for > operator.
	CompareOpNodeKindGreater
	// CompareOpNodeKindGreaterEq is kind for >= operator.
	CompareOpNodeKindGreaterEq
	// CompareOpNodeKindEq is kind for == operator.
	CompareOpNodeKindEq
	// CompareOpNodeKindNotEq is kind for != operator.
	CompareOpNodeKindNotEq
)

// IsEqualityOp returns true when it represents == or != operator.
func (kind CompareOpNodeKind) IsEqualityOp() bool {
	return kind == CompareOpNodeKindEq || kind == CompareOpNodeKindNotEq
}

func (kind CompareOpNodeKind) String() string {
	switch kind {
	case CompareOpNodeKindLess:
		return "<"
	case CompareOpNodeKindLessEq:
		return "<="
	case CompareOpNodeKindGreater:
		return ">"
	case CompareOpNodeKindGreaterEq:
		return ">="
	case CompareOpNodeKindEq:
		return "=="
	case CompareOpNodeKindNotEq:
		return "!="
	default:
		return ""
	}
}

// CompareOpNode is node for binary expression to compare values; ==, !=, <, <=, > or >=.
type CompareOpNode struct {
	// Kind is a kind of this expression to show which operator is used.
	Kind CompareOpNodeKind
	// Left is an expression for left hand side of the binary operator.
	Left Node
	// Right is an expression for right hand side of the binary operator.
	Right Node
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *CompareOpNode) Token() *Token {
	return n.Left.Token()
}

// LogicalOpNodeKind is a kind of logical operators; && and ||.
type LogicalOpNodeKind int

const (
	// LogicalOpNodeKindInvalid is an invalid and initial value of LogicalOpNodeKind.
	LogicalOpNodeKindInvalid LogicalOpNodeKind = iota
	// LogicalOpNodeKindAnd is a kind for && operator.
	LogicalOpNodeKindAnd
	// LogicalOpNodeKindOr is a kind for || operator.
	LogicalOpNodeKindOr
)

func (k LogicalOpNodeKind) String() string {
	switch k {
	case LogicalOpNodeKindAnd:
		return "&&"
	case LogicalOpNodeKindOr:
		return "||"
	default:
		return "INVALID LOGICAL OPERATOR"
	}
}

// LogicalOpNode is node for logical binary operators; && or ||.
type LogicalOpNode struct {
	// Kind is a kind to show which operator is used.
	Kind LogicalOpNodeKind
	// Left is an expression for left hand side of the binary operator.
	Left Node
	// Right is an expression for right hand side of the binary operator.
	Right Node
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *LogicalOpNode) Token() *Token {
	return n.Left.Token()
}

// FuncCallNode represents function call in expression.
// Note that currently only calling builtin functions is supported.
type FuncCallNode struct {
	// Callee is a name of called function. This is string value because currently only built-in
	// functions can be called.
	Callee string
	// Args is arguments of the function call.
	Args []Node
	tok  *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *FuncCallNode) Token() *Token {
	return n.tok
}

// VisitExprNodeFunc is a visitor function for VisitExprNode(). The entering argument is set to
// true when it is called before visiting children. It is set to false when it is called after
// visiting children. It means that this function is called twice for the same node. The parent
// argument is the parent of the node. When the node is root, its parent is nil.
type VisitExprNodeFunc func(node, parent Node, entering bool)

func visitExprNode(n, p Node, f VisitExprNodeFunc) {
	f(n, p, true)
	switch n := n.(type) {
	case *ObjectDerefNode:
		visitExprNode(n.Receiver, n, f)
	case *ArrayDerefNode:
		visitExprNode(n.Receiver, n, f)
	case *IndexAccessNode:
		// Index must be visited before Operand to make UntrustedInputChecker work correctly.
		visitExprNode(n.Index, n, f)
		visitExprNode(n.Operand, n, f)
	case *NotOpNode:
		visitExprNode(n.Operand, n, f)
	case *CompareOpNode:
		visitExprNode(n.Left, n, f)
		visitExprNode(n.Right, n, f)
	case *LogicalOpNode:
		visitExprNode(n.Left, n, f)
		visitExprNode(n.Right, n, f)
	case *FuncCallNode:
		for _, a := range n.Args {
			visitExprNode(a, n, f)
		}
	}
	f(n, p, false)
}

// VisitExprNode visits the given expression syntax tree with given function f.
func VisitExprNode(n Node, f VisitExprNodeFunc) {
	visitExprNode(n, nil, f)
}

// VisualizeAST prints a tree of string representation with indentation as the marker of tree nesting.
func VisualizeAST(n Node) string {
	sb := strings.Builder{}
	indent := 0

	visitor := VisitExprNodeFunc(func(node, parent Node, entering bool) {
		if entering {
			sb.WriteString(strings.Repeat("  ", indent))
			sb.WriteString(fmt.Sprintf("%T ", node))
			switch n := node.(type) {

			case *ObjectDerefNode:
				sb.WriteString("\n")
				indent += 1
				sb.WriteString(strings.Repeat("  ", indent))
				sb.WriteString("receiver: ")
			case *IndexAccessNode:
				sb.WriteString("\n")
				indent += 1
				sb.WriteString(strings.Repeat("  ", indent))
				sb.WriteString("index: ")
			case *IntNode:
				switch parent.(type) {
				case *IndexAccessNode:
					indent -= 1
				}
				sb.WriteString(fmt.Sprintf("%d", n.Value))
			case *LogicalOpNode:
				sb.WriteString(n.Kind.String())
			case *CompareOpNode:
				sb.WriteString(n.Kind.String())
			case *FuncCallNode:
				sb.WriteString(n.Callee + "()")

			default:
				sb.WriteString(n.Token().Value)
			}
			sb.WriteString("\n")
			indent += 1
		} else {
			switch n := node.(type) {
			case *ObjectDerefNode:
				indent -= 1
				sb.WriteString(strings.Repeat("  ", indent))
				sb.WriteString("property: ")
				sb.WriteString(n.Property)
				sb.WriteString("\n")
			case *IndexAccessNode:
				indent -= 1
			case *IntNode:
				switch parent.(type) {
				case *IndexAccessNode:
					sb.WriteString(strings.Repeat("  ", indent-1))
					indent += 1
					sb.WriteString("operand:\n")
				}
			}
			indent -= 1
		}
	})

	VisitExprNode(n, visitor)
	return sb.String()
}
