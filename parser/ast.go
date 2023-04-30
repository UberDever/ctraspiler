package parser

import (
	"fmt"
	"math"
	"strings"
)

type NodeTag = int

const (
	NodeInvalid   NodeTag = math.MinInt
	NodeUndefined         = -1
)
const (
	NodeSource = iota
	NodeBlock

	NodeFunctionDecl
	NodeSignature
	NodeConstDecl
	NodeAssignment

	NodeSelector
	NodeCall

	NodeOr
	NodeAnd
	NodeEquals
	NodeNotEquals
	NodeGreaterThan
	NodeLessThan
	NodeGreaterThanEquals
	NodeLessThanEquals
	NodeBinaryPlus
	NodeBinaryMinus
	NodeMultiply
	NodeDivide

	NodeUnaryPlus
	NodeUnaryMinus
	NodeNot

	NodeIntLiteral
	NodeFloatLiteral
	NodeStringLiteral
	NodeIdentifier

	NodeIdentifierList
	NodeExpressionList
)

// IdentifierList { idx=...; lhs=start; rhs=end }
// ExpressionList { tag: NodeExpressionList; lhs=start; rhs=end }
// BinaryOp { tag: NodePlus, NodeMinus...; lhs; rhs }
// UnaryOp { tag: NodeUnaryMinus...; lhs; rhs=-1 }
// Literal { tag: NodeIntLiteral...; lhs; rhs=-1 }
// Identifier { tag: NodeIdentifier; lhs; rhs=-1 }

// Every node is semantically represented in the traversal process, before that
// we store nodes as non-typed data
// Legend for node fields representation:
// 1. ... means field is unused/not relevant
// 2. start/end pair means node is variadic (0 and more children)
// 3. extra means that node has known! amount of children, stored in extra buffer
// that amount depends on the node type
type Node struct {
	tag      NodeTag
	tokenIdx Index
	lhs, rhs Index
}

var NullNode = Node{
	tag:      NodeInvalid,
	tokenIdx: TokenEOF,
	lhs:      NodeInvalid,
	rhs:      NodeInvalid,
}

// tokenIdx=... lhs=start rhs=end
type SourceRoot struct{ Node }

func (n SourceRoot) Children(ast *AST) []int {
	decls := make([]int, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		decls = append(decls, c_i)
	}
	return decls
}

// tokenIdx=name lhs=signature rhs=body
type FunctionDecl struct{ Node }

func (n FunctionDecl) Signature(ast *AST) int {
	return n.lhs
}

func (n FunctionDecl) Body(ast *AST) int {
	return n.rhs
}

func (n FunctionDecl) GetName(ast *AST) string {
	return ast.src.lexeme(ast.src.token(n.tokenIdx))
}

// tokenIdx=... lhs=identifierList rhs=...
type Signature struct{ Node }

func (n Signature) Input(ast *AST) int {
	return n.lhs
}

// tokenIdx=... lhs=start rhs=end
type Block struct{ Node }

func (n Block) Children(ast *AST) []int {
	statements := make([]int, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		statements = append(statements, c_i)
	}
	return statements
}

// tokenIdx=const lhs=identifierList rhs=exprList
type ConstDecl struct{ Node }

func (n ConstDecl) Identifiers(ast *AST) []int {
	lhs := ast.nodes[n.lhs]
	return IdentifierList{lhs}.Children(ast)
}

func (n ConstDecl) Expressions(ast *AST) []int {
	rhs := ast.nodes[n.rhs]
	return ExpressionList{rhs}.Children(ast)
}

// tokenIdx=... lhs=expr rhs=identifier
type Selector struct{ Node }

// tokenIdx=... lhs=expr rhs=exprList
type Call struct{ Node }

type IdentifierList struct{ Node }

func (n IdentifierList) Children(ast *AST) []int {
	// identifiers
	ids := make([]int, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		ids = append(ids, c_i)
	}
	return ids
}

type ExpressionList struct{ Node }

func (n ExpressionList) Children(ast *AST) []int {
	// expressions
	exprs := make([]int, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		exprs = append(exprs, c_i)
	}
	return exprs
}

type AST struct {
	src   *Source
	nodes []Node
	extra []Index
}

func (ast *AST) GetNodeString(n Node) string {
	switch n.tag {
	case NodeSource:
		return "Source"
	case NodeFunctionDecl:
		return fmt.Sprintf("FunctionDecl %s", FunctionDecl{n}.GetName(ast))
	case NodeSignature:
		return "Signature"
	case NodeBlock:
		return "Block"
	case NodeConstDecl:
		return "ConstDecl"
	case NodeSelector:
		return "Selector"
	case NodeCall:
		return "Call"

	case NodeOr:
		return "||"
	case NodeAnd:
		return "&&"
	case NodeEquals:
		return "=="
	case NodeNotEquals:
		return "!="
	case NodeGreaterThan:
		return ">"
	case NodeLessThan:
		return "<"
	case NodeGreaterThanEquals:
		return ">="
	case NodeLessThanEquals:
		return "<="
	case NodeBinaryPlus:
		return "+"
	case NodeBinaryMinus:
		return "-"
	case NodeMultiply:
		return "*"
	case NodeDivide:
		return "/"

	case NodeUnaryPlus:
		return "+"
	case NodeUnaryMinus:
		return "-"
	case NodeNot:
		return "Not"

	case NodeIdentifierList:
		ids := IdentifierList{n}.Children(ast)
		s := make([]string, 0, len(ids))
		for _, i := range ids {
			c := ast.nodes[i]
			s = append(s, ast.src.lexeme(ast.src.token(c.tokenIdx)))
		}
		return strings.Join(s, " ")
	case NodeExpressionList:
		return "ExpressionList"
	case NodeIdentifier:
		return ast.src.lexeme(ast.src.token(n.tokenIdx))
	case NodeIntLiteral:
		fallthrough
	case NodeFloatLiteral:
		fallthrough
	case NodeStringLiteral:
		return ast.src.lexeme(ast.src.token(n.tokenIdx))
	}
	return "Undefined"

}

type NodeAction = func(*AST, Node)

func (ast *AST) Traverse(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNode(onEnter, onExit, 0)
}

func (ast *AST) traverseNode(onEnter NodeAction, onExit NodeAction, i Index) {
	if i == IndexUndefined {
		return
	}

	if i == IndexInvalid {
		panic(fmt.Sprintf("While traversing nodes, encountered invalid index %d", i))
	}
	n := ast.nodes[i]
	if n.tag == NodeInvalid ||
		n.tokenIdx == TokenEOF ||
		n.lhs == NodeInvalid ||
		n.rhs == NodeInvalid {
		panic(fmt.Sprintf("While traversing nodes, encountered null node (at %d)", i))
	}

	onEnter(ast, n)
	switch n.tag {
	case NodeSource:
		nodes := SourceRoot{n}.Children(ast)
		for _, c := range nodes {
			ast.traverseNode(onEnter, onExit, c)
		}
	case NodeFunctionDecl:
		node := FunctionDecl{n}
		ast.traverseNode(onEnter, onExit, node.Signature(ast))
		ast.traverseNode(onEnter, onExit, node.Body(ast))
	case NodeSignature:
		ast.traverseNode(onEnter, onExit, n.lhs)
	case NodeBlock:
		nodes := Block{n}.Children(ast)
		for _, c := range nodes {
			ast.traverseNode(onEnter, onExit, c)
		}
	case NodeConstDecl:
		ast.traverseNode(onEnter, onExit, n.lhs)
		ast.traverseNode(onEnter, onExit, n.rhs)

	case NodeSelector:
		ast.traverseNode(onEnter, onExit, n.lhs)
		ast.traverseNode(onEnter, onExit, n.rhs)

	case NodeCall:
		ast.traverseNode(onEnter, onExit, n.lhs)
		ast.traverseNode(onEnter, onExit, n.rhs)

	case NodeOr,
		NodeAnd,
		NodeEquals,
		NodeNotEquals,
		NodeGreaterThan,
		NodeLessThan,
		NodeGreaterThanEquals,
		NodeLessThanEquals,
		NodeBinaryPlus,
		NodeBinaryMinus,
		NodeMultiply,
		NodeDivide:
		ast.traverseNode(onEnter, onExit, n.lhs)
		ast.traverseNode(onEnter, onExit, n.rhs)

	case NodeUnaryPlus,
		NodeUnaryMinus,
		NodeNot:
		ast.traverseNode(onEnter, onExit, n.lhs)

	case NodeIdentifierList:
		break // do nothing
	case NodeExpressionList:
		n_ := ExpressionList{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(onEnter, onExit, c)
		}
	}

	onExit(ast, n)
}

func formatSExpr(sexpr string) string {
	formatted := strings.Builder{}
	depth := -1
	for i := range sexpr {
		if sexpr[i] == '(' {
			depth++
			formatted.WriteByte('\n')
			for j := 0; j < depth; j++ {
				formatted.WriteString("    ")
			}
			formatted.WriteByte('(')
		} else if sexpr[i] == ')' {
			depth--
			formatted.WriteByte(')')
		} else {
			formatted.WriteByte(sexpr[i])
		}
	}
	return formatted.String()
}

func unformatSExpr(s string) string {
	formatted := strings.Builder{}
	skipWS := func(i int) (int, bool) {
		wasSpace := false
		for s[i] == ' ' || s[i] == '\n' || s[i] == '\t' {
			wasSpace = true
			i++
			if i >= len(s) {
				break
			}
		}
		return i, wasSpace
	}

	for i := 0; i < len(s); i++ {
		j, wasSpace := skipWS(i)
		if j >= len(s) {
			break
		}
		i = j
		if wasSpace {
			if s[i] != '(' && s[i] != ')' {
				formatted.WriteByte(' ')
			}
		}
		formatted.WriteByte(s[i])
	}
	return formatted.String()
}

func (ast *AST) dump(doFormat bool) string {
	str := strings.Builder{}
	onEnter := func(ast *AST, node Node) {
		str.WriteByte('(')
		str.WriteString(ast.GetNodeString(node))
	}
	onExit := func(ast *AST, node Node) {
		str.WriteByte(')')
	}
	ast.Traverse(onEnter, onExit)

	if doFormat {
		return formatSExpr(str.String())
	}
	return str.String()
}
