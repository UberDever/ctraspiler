package parser

import (
	"fmt"
	"strings"
)

const (
	NodeUndefined = -1
	NodeSource    = iota
	NodeBlock
	NodeIdentifierList
	NodeExpressionList

	NodeFunctionDecl
	NodeSignature
	NodeConstDecl

	NodeCall

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

// Source { idx=...; lhs=start; rhs=end }
// FunctionDecl { idx=name; lhs=signature; rhs=body }
// Signature { idx=...; lhs=parameters; rhs=0 }
// Parameters { idx=...; lhs=start; rhs=end }
// ConstDecl { idx=const; lhs=identifierList; rhs=expList }

// Block { tag: NodeBlock; lhs=start; rhs=end }
// Call { tag: NodeCall; lhs=expression; rhs=body }
// IdentifierList { idx=...; lhs=start; rhs=end }
// ExpressionList { tag: NodeExpressionList; lhs=start; rhs=end }
// BinaryOp { tag: NodePlus, NodeMinus...; lhs; rhs }
// UnaryOp { tag: NodeUnaryMinus...; lhs; rhs=-1 }
// Literal { tag: NodeIntLiteral...; lhs; rhs=-1 }
// Identifier { tag: NodeIdentifier; lhs; rhs=-1 }

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
	return []int{n.lhs}
}

type Block struct{ Node }

func (n Block) Children(ast *AST) []int {
	// statements
	statements := make([]int, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		statements = append(statements, c_i)
	}
	return statements
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
	case NodeIdentifierList:
		ids := IdentifierList{n}.Identifiers(ast)
		return strings.Join(ids, " ")
	case NodeIntLiteral:
		fallthrough
	case NodeFloatLiteral:
		fallthrough
	case NodeStringLiteral:
		return ast.src.lexeme(ast.src.token(n.lhs))
	}
	return "Undefined"

}

type NodeAction = func(*AST, Node)

func (ast *AST) Traverse(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNode(onEnter, onExit, 0)
}

func (ast *AST) traverseNode(onEnter NodeAction, onExit NodeAction, node_i int) {
	n := ast.nodes[node_i]
	if n.tag == NodeUndefined ||
		n.tokenIdx == -1 ||
		n.lhs == -1 ||
		n.rhs == -1 {
		panic(fmt.Sprintf("While traversing nodes, encountered null node (at %d)", node_i))
	}
	onEnter(ast, n)

	switch n.tag {
	case NodeSource:
		n_ := SourceRoot{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(onEnter, onExit, c)
		}
	case NodeFunctionDecl:
		n_ := FunctionDecl{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(onEnter, onExit, c)
		}
	case NodeSignature:
		n_ := Signature{n}.Children(ast)
		for _, c := range n_ {
			ast.traverseNode(onEnter, onExit, c)
		}
	case NodeBlock:
		n_ := Block{n}.Children(ast)
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
