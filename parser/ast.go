package parser

import (
	"fmt"
	"math"
	"strings"
)

type NodeTag int
type NodeIndex int

const (
	NodeIndexInvalid   NodeIndex = math.MinInt
	NodeIndexUndefined           = -1
)

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
	tokenIdx TokenIndex
	lhs, rhs NodeIndex
}

var InvalidNode = Node{
	tag:      NodeInvalid,
	tokenIdx: TokenIndexInvalid,
	lhs:      NodeIndexInvalid,
	rhs:      NodeIndexInvalid,
}

// General pattern of typed nodes:
// Struct with indexes and booleans
// Function to convert from node to typed node
// Function to convert node to string
// Function to get node children

type SourceRoot struct {
	declarations []NodeIndex
}

func (ast AST) SourceRoot(n Node) SourceRoot {
	decls := make([]NodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := NodeIndex(ast.extra[i])
		decls = append(decls, c_i)
	}
	return SourceRoot{
		declarations: decls,
	}
}

func (ast AST) SourceRoot_Children(i NodeIndex) []NodeIndex {
	n := ast.SourceRoot(ast.nodes[i])
	return n.declarations
}

func (ast AST) SourceRoot_String(i NodeIndex) string {
	return "Source"
}

type FunctionDecl struct {
	name      NodeIndex
	signature NodeIndex
	body      NodeIndex
}

func (ast AST) FunctionDecl(n Node) FunctionDecl {
	node := FunctionDecl{}
	// find identifier node by it's token index
	node.name = NodeIndexInvalid
	for i := range ast.nodes {
		if ast.nodes[i].tokenIdx == n.tokenIdx {
			node.name = NodeIndex(i)
		}
	}
	if node.name == NodeIndexInvalid {
		panic("This shouldn't have happened!")
	}
	node.signature = n.lhs
	node.body = n.rhs
	return node
}

func (ast AST) FunctionDecl_Children(i NodeIndex) []NodeIndex {
	n := ast.FunctionDecl(ast.nodes[i])
	return []NodeIndex{n.name, n.signature, n.body}
}

func (ast AST) FunctionDecl_String(i NodeIndex) string {
	n := ast.FunctionDecl(ast.nodes[i])
	return fmt.Sprintf("FunctionDecl %s",
		ast.Identifier_String(n.name))
}

type Signature struct {
	parameters []NodeIndex
}

func (ast AST) Signature(n Node) Signature {
	params := make([]NodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := NodeIndex(ast.extra[i])
		params = append(params, c_i)
	}
	return Signature{
		parameters: params,
	}
}

func (ast AST) Signature_Children(i NodeIndex) []NodeIndex {
	n := ast.Signature(ast.nodes[i])
	return n.parameters
}

func (ast AST) Signature_String(i NodeIndex) string {
	return "Signature"
}

type Block struct {
	statements []NodeIndex
}

func (ast AST) Block(n Node) Block {
	statements := make([]NodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := NodeIndex(ast.extra[i])
		statements = append(statements, c_i)
	}
	return Block{
		statements: statements,
	}
}

func (ast AST) Block_Children(i NodeIndex) []NodeIndex {
	n := ast.Block(ast.nodes[i])
	return n.statements
}

func (ast AST) Block_String(i NodeIndex) string {
	return "Block"
}

type ConstDecl struct {
	identifierList NodeIndex
	expressionList NodeIndex
}

func (ast AST) ConstDecl(n Node) ConstDecl {
	return ConstDecl{
		identifierList: n.lhs,
		expressionList: n.rhs,
	}
}

func (ast AST) ConstDecl_Children(i NodeIndex) []NodeIndex {
	n := ast.ConstDecl(ast.nodes[i])
	return []NodeIndex{n.identifierList, n.expressionList}
}

func (ast AST) ConstDecl_String(i NodeIndex) string {
	return "ConstDecl"
}

// tokenIdx=... lhs=expr rhs=identifier
type Selector struct{ Node }

// tokenIdx=... lhs=expr rhs=exprList
type Call struct{ Node }

type IdentifierList struct {
	identifiers []NodeIndex
}

func (ast AST) IdentifierList(n Node) IdentifierList {
	ids := make([]NodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		ids = append(ids, NodeIndex(c_i))
	}
	return IdentifierList{
		identifiers: ids,
	}
}

func (ast AST) IdentifierList_Children(i NodeIndex) []NodeIndex {
	n := ast.IdentifierList(ast.nodes[i])
	return n.identifiers
}

func (ast AST) IdentifierList_String(i NodeIndex) string {
	ids := ast.IdentifierList_Children(i)
	s := make([]string, 0, len(ids))
	for _, i := range ids {
		s = append(s, ast.Identifier_String(i))
	}
	return strings.Join(s, " ")
}

type Identifier struct {
	token TokenIndex
}

func (ast AST) Identifier(n Node) Identifier {
	return Identifier{
		token: n.tokenIdx,
	}
}

func (ast AST) Identifier_Children(i NodeIndex) []NodeIndex {
	return []NodeIndex{}
}

func (ast AST) Identifier_String(i NodeIndex) string {
	n := ast.Identifier(ast.nodes[i])
	return ast.src.lexeme(ast.src.token(n.token))
}

type ExpressionList struct {
	expressions []NodeIndex
}

func (ast AST) ExpressionList(n Node) ExpressionList {
	exprs := make([]NodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		exprs = append(exprs, NodeIndex(c_i))
	}
	return ExpressionList{
		expressions: exprs,
	}
}

func (ast AST) ExpressionList_Children(i NodeIndex) []NodeIndex {
	n := ast.IdentifierList(ast.nodes[i])
	return n.identifiers
}

func (ast AST) ExpressionList_String(i NodeIndex) string {
	ids := ast.ExpressionList_Children(i)
	s := make([]string, 0, len(ids))
	// TODO: Do a table of expression_string things here
	// for _, i := range ids {
	// 	// s = append(s, ast.Identifier_String(i))
	// }
	return strings.Join(s, " ")
}

type AST struct {
	src   *Source
	nodes []Node
	extra []AnyIndex
}

// TODO: Add here a table of function pointers like NodeType -> (Node -> []Index)
func (ast AST) GetNodeString(n Node) string {
	// case NodeSelector:
	// 	return "Selector"
	// case NodeCall:
	// 	return "Call"
	// case NodeAssignment:
	// 	return "Assign"

	// case NodeOr:
	// 	return "||"
	// case NodeAnd:
	// 	return "&&"
	// case NodeEquals:
	// 	return "=="
	// case NodeNotEquals:
	// 	return "!="
	// case NodeGreaterThan:
	// 	return ">"
	// case NodeLessThan:
	// 	return "<"
	// case NodeGreaterThanEquals:
	// 	return ">="
	// case NodeLessThanEquals:
	// 	return "<="
	// case NodeBinaryPlus:
	// 	return "+"
	// case NodeBinaryMinus:
	// 	return "-"
	// case NodeMultiply:
	// 	return "*"
	// case NodeDivide:
	// 	return "/"

	// case NodeUnaryPlus:
	// 	return "+"
	// case NodeUnaryMinus:
	// 	return "-"
	// case NodeNot:
	// 	return "Not"

	return "Undefined"

}

type NodeAction = func(*AST, Node)

func (ast *AST) Traverse(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNode(onEnter, onExit, 0)
}

// TODO: Add here a table of function pointers like NodeType -> (Node -> []Index)
func (ast *AST) traverseNode(onEnter NodeAction, onExit NodeAction, i NodeIndex) {
	if i == NodeIndexUndefined {
		return
	}

	if i == NodeIndexInvalid {
		panic(fmt.Sprintf("While traversing nodes, encountered invalid index %d", i))
	}
	n := ast.nodes[i]
	if n.tag == NodeInvalid ||
		n.tokenIdx == TokenIndexInvalid ||
		n.lhs == NodeIndexInvalid ||
		n.rhs == NodeIndexInvalid {
		panic(fmt.Sprintf("While traversing nodes, encountered null node (at %d)", i))
	}

	onEnter(ast, n)
	// TODO: this switch should focus on concrete node treating, not node traversal
	// cause this will be handled by the map of function pointers
	// and likely this switch won't be present here at all, but will be somewhere
	// else (like in code gen)
	// switch n.tag {
	// case NodeSource:
	// 	nodes := SourceRoot{n}.Children(ast)
	// 	for _, c := range nodes {
	// 		ast.traverseNode(onEnter, onExit, c)
	// 	}
	// case NodeFunctionDecl:
	// 	node := FunctionDecl{n}
	// 	ast.traverseNode(onEnter, onExit, node.Signature(ast))
	// 	ast.traverseNode(onEnter, onExit, node.Body(ast))
	// case NodeSignature:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)
	// case NodeBlock:
	// 	nodes := Block{n}.Children(ast)
	// 	for _, c := range nodes {
	// 		ast.traverseNode(onEnter, onExit, c)
	// 	}
	// case NodeConstDecl:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)
	// 	ast.traverseNode(onEnter, onExit, n.rhs)

	// case NodeSelector:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)
	// 	ast.traverseNode(onEnter, onExit, n.rhs)

	// case NodeCall:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)
	// 	ast.traverseNode(onEnter, onExit, n.rhs)

	// case NodeAssignment:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)
	// 	ast.traverseNode(onEnter, onExit, n.rhs)

	// case NodeOr,
	// 	NodeAnd,
	// 	NodeEquals,
	// 	NodeNotEquals,
	// 	NodeGreaterThan,
	// 	NodeLessThan,
	// 	NodeGreaterThanEquals,
	// 	NodeLessThanEquals,
	// 	NodeBinaryPlus,
	// 	NodeBinaryMinus,
	// 	NodeMultiply,
	// 	NodeDivide:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)
	// 	ast.traverseNode(onEnter, onExit, n.rhs)

	// case NodeUnaryPlus,
	// 	NodeUnaryMinus,
	// 	NodeNot:
	// 	ast.traverseNode(onEnter, onExit, n.lhs)

	// case NodeIdentifierList:
	// 	break // do nothing
	// case NodeExpressionList:
	// 	n_ := ExpressionList{n}.Children(ast)
	// 	for _, c := range n_ {
	// 		ast.traverseNode(onEnter, onExit, c)
	// 	}
	// }

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
