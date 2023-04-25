package parser

import "fmt"

const (
	NodeUndefined = -1
	NodeSource    = iota
	NodeFunctionDecl
	NodeSignature
	NodeParameters
	NodeBlock
	NodeCall
	NodeLetDecl

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

// Source { tag: NodeSource; lhs=start; rhs=end }
// FunctionDecl { tag: NodeFunctionDecl; lhs=signature; rhs=body }
// Signature { tag: NodeSignature; lhs=name; rhs=parameters }
// Parameters { tag: NodeParameters; lhs=start; rhs=end }
// Block { tag: NodeBlock; lhs=start; rhs=end }
// Call { tag: NodeCall; lhs=expression; rhs=body }
// LetDecl { tag: NodeLetDecl; lhs=identifier; rhs=init }
// BinaryOp { tag: NodePlus, NodeMinus...; lhs; rhs }
// UnaryOp { tag: NodeUnaryMinus...; lhs; rhs=-1 }
// Literal { tag: NodeIntLiteral...; lhs; rhs=-1 }
// Identifier { tag: NodeIdentifier; lhs; rhs=-1 }

type AST struct {
	src   *Source
	nodes []Node
	extra []Index
}

func (ast *AST) GetNodeString(n Node) string {
	switch n.tag {
	case NodeSource:
		return "Source"
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
	ast.traverseNodes(f, 0)
}

// TODO: Do ast in sexpr

func (ast *AST) traverseNodes(f func(*AST, Node), current int) {
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
		for i := n.lhs; i < n.rhs; i++ {
			c_i := ast.extra[i]
			ast.traverseNodes(f, c_i)
		}
	case NodeFunctionDecl:
		ast.traverseNodes(f, n.lhs)
	}
}
