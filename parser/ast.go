package parser

import (
	"fmt"
	"strings"
)

const (
	NodeUndefined = -1
	NodeSource    = iota
	NodeFunctionDecl
	NodeSignature
	NodeBlock
	NodeCall
	NodeLetDecl
	NodeIdentifierList
	NodeExpressionList

	NodeBinaryPlus
	NodeBinaryMinus
	NodeMultiply
	NodeDivide

	NodeUnaryPlus
	NodeUnaryMinus

	NodeIntLiteral
	NodeFloatLiteral
	NodeStringLiteral
	NodeIdentifier
)

type Node struct {
	tag      Tag
	tokenIdx Index
	lhs, rhs Index
}

var NullNode = Node{
	tag:      NodeUndefined,
	tokenIdx: -1,
	lhs:      -1,
	rhs:      -1,
}

type SourceRoot struct{ Node }

func (n SourceRoot) Children(ast *AST) []int {
	// declarations
	decls := make([]int, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		decls = append(decls, c_i)
	}
	return decls
}

type FunctionDecl struct{ Node }

func (n FunctionDecl) Children(ast *AST) []int {
	// block
	if n.rhs != 0 {
		return []int{n.lhs, n.rhs}
	}
	// signature
	return []int{n.lhs}
}

func (n FunctionDecl) GetName(ast *AST) string {
	return ast.src.lexeme(ast.src.token(n.tokenIdx))
}

type Signature struct{ Node }

func (n Signature) Children(ast *AST) []int {
	// parameters
	if n.lhs != 0 {
		return []int{n.lhs}
	}
	return []int{}
}

type IdentifierList struct{ Node }

func (n IdentifierList) Identifiers(ast *AST) []string {
	// identifiers
	ids := make([]string, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		ids = append(ids, ast.src.lexeme(ast.src.token(c_i)))
	}
	return ids
}

// Source { idx=...; lhs=start; rhs=end }
// FunctionDecl { idx=name; lhs=signature; rhs=body }
// Signature { idx=...; lhs=parameters; rhs=0 }
// Parameters { idx=...; lhs=start; rhs=end }

// Block { tag: NodeBlock; lhs=start; rhs=end }
// Call { tag: NodeCall; lhs=expression; rhs=body }
// LetDecl { tag: NodeLetDecl; lhs=identifier; rhs=init }
// IdentifierList { idx=...; lhs=start; rhs=end }
// ExpressionList { tag: NodeExpressionList; lhs=start; rhs=end }
// BinaryOp { tag: NodePlus, NodeMinus...; lhs; rhs }
// UnaryOp { tag: NodeUnaryMinus...; lhs; rhs=-1 }
// Literal { tag: NodeIntLiteral...; lhs; rhs=-1 }
// Identifier { tag: NodeIdentifier; lhs; rhs=-1 }

type AST struct {
	src   *Source
	nodes []Node
	extra []Index
}

// TODO: Do ast in sexpr
func (ast *AST) GetNodeString(n Node) string {
	switch n.tag {
	case NodeSource:
		return "Source"
	case NodeFunctionDecl:
		return fmt.Sprintf("FunctionDecl [name=%s]", FunctionDecl{n}.GetName(ast))
	case NodeSignature:
		return "Signature"
	case NodeIdentifierList:
		ids := IdentifierList{n}.Identifiers(ast)
		return strings.Join(ids, ", ")
	case NodeIntLiteral:
		fallthrough
	case NodeFloatLiteral:
		fallthrough
	case NodeStringLiteral:
		return ast.src.lexeme(ast.src.token(n.lhs))
	}
	return "Undefined"

}

func (ast *AST) Traverse(f func(*AST, Node)) {
	ast.traverseNode(f, 0)
}

func (ast *AST) traverseNode(f func(*AST, Node), current int) {
	n := ast.nodes[current]
	if n.tag == NodeUndefined ||
		n.tokenIdx == -1 ||
		n.lhs == -1 ||
		n.rhs == -1 {
		panic(fmt.Sprintf("While traversing nodes, encountered null node (at %d)", current))
	}

	f(ast, n)

	switch n.tag {
	case NodeSource:
		n_ := SourceRoot{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(f, c)
		}
	case NodeFunctionDecl:
		n_ := FunctionDecl{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(f, c)
		}
	case NodeSignature:
		n_ := Signature{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(f, c)
		}
	case NodeIdentifierList:
		return

	case NodeBlock:
		// n_ := Signature{n}.Children(ast)
		// for _, c := range n_ {
		// 	ast.traverseNode(f, c)
		// }
	}
}
