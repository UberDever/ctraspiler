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

var nodeIndexToString = [...]func(AST, NodeIndex) string{
	NodeSource: SourceRoot_String,
	NodeBlock:  Block_String,

	NodeFunctionDecl: FunctionDecl_String,
	NodeSignature:    Signature_String,
	NodeConstDecl:    ConstDecl_String,
	NodeAssignment:   Assignment_String,

	NodeSelector: Selector_String,
	NodeCall:     Call_String,

	NodeOr:                Or_String,
	NodeAnd:               And_String,
	NodeEquals:            Equals_String,
	NodeNotEquals:         NotEquals_String,
	NodeGreaterThan:       GreaterThan_String,
	NodeLessThan:          LessThan_String,
	NodeGreaterThanEquals: GreaterThanEquals_String,
	NodeLessThanEquals:    LessThanEquals_String,
	NodeBinaryPlus:        BinaryPlus_String,
	NodeBinaryMinus:       BinaryMinus_String,
	NodeMultiply:          Multiply_String,
	NodeDivide:            Divide_String,

	NodeUnaryPlus:  UnaryPlus_String,
	NodeUnaryMinus: UnaryMinus_String,
	NodeNot:        Not_String,

	NodeIntLiteral:    IntLiteral_String,
	NodeFloatLiteral:  FloatLiteral_String,
	NodeStringLiteral: StringLiteral_String,
	NodeIdentifier:    Identifier_String,

	NodeIdentifierList: IdentifierList_String,
	NodeExpressionList: ExpressionList_String,
}

var nodeIndexToChildren = [...]func(AST, NodeIndex) []NodeIndex{
	NodeSource: SourceRoot_Children,
	NodeBlock:  Block_Children,

	NodeFunctionDecl: FunctionDecl_Children,
	NodeSignature:    Signature_Children,
	NodeConstDecl:    ConstDecl_Children,
	NodeAssignment:   Assignment_Children,

	NodeSelector: Selector_Children,
	NodeCall:     Call_Children,

	NodeOr:                Or_Children,
	NodeAnd:               And_Children,
	NodeEquals:            Equals_Children,
	NodeNotEquals:         NotEquals_Children,
	NodeGreaterThan:       GreaterThan_Children,
	NodeLessThan:          LessThan_Children,
	NodeGreaterThanEquals: GreaterThanEquals_Children,
	NodeLessThanEquals:    LessThanEquals_Children,
	NodeBinaryPlus:        BinaryPlus_Children,
	NodeBinaryMinus:       BinaryMinus_Children,
	NodeMultiply:          Multiply_Children,
	NodeDivide:            Divide_Children,

	NodeUnaryPlus:  UnaryPlus_Children,
	NodeUnaryMinus: UnaryMinus_Children,
	NodeNot:        Not_Children,

	NodeIntLiteral:    IntLiteral_Children,
	NodeFloatLiteral:  FloatLiteral_Children,
	NodeStringLiteral: StringLiteral_Children,
	NodeIdentifier:    Identifier_Children,

	NodeIdentifierList: IdentifierList_Children,
	NodeExpressionList: ExpressionList_Children,
}

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

func SourceRoot_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.SourceRoot(ast.nodes[i])
	return n.declarations
}

func SourceRoot_String(ast AST, i NodeIndex) string {
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
		id := ast.nodes[i]
		if n.tokenIdx == id.tokenIdx &&
			id.tag == NodeIdentifier {
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

func FunctionDecl_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.FunctionDecl(ast.nodes[i])
	return []NodeIndex{n.name, n.signature, n.body}
}

func FunctionDecl_String(ast AST, i NodeIndex) string {
	return "FunctionDecl"
}

type Signature struct {
	parameters NodeIndex
}

func (ast AST) Signature(n Node) Signature {
	return Signature{
		parameters: n.lhs,
	}
}

func Signature_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Signature(ast.nodes[i])
	return []NodeIndex{n.parameters}
}

func Signature_String(ast AST, i NodeIndex) string {
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

func Block_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Block(ast.nodes[i])
	return n.statements
}

func Block_String(ast AST, i NodeIndex) string {
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

func ConstDecl_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.ConstDecl(ast.nodes[i])
	return []NodeIndex{n.identifierList, n.expressionList}
}

func ConstDecl_String(ast AST, i NodeIndex) string {
	return "ConstDecl"
}

type Assignment struct {
	lhsList NodeIndex
	rhsList NodeIndex
}

func (ast AST) Assignment(n Node) Assignment {
	return Assignment{
		lhsList: n.lhs,
		rhsList: n.rhs,
	}
}

func Assignment_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Assignment(ast.nodes[i])
	return []NodeIndex{n.lhsList, n.rhsList}
}

func Assignment_String(ast AST, i NodeIndex) string {
	return "Assign"
}

type Selector struct {
	lhsExpr    NodeIndex
	identifier NodeIndex
}

func (ast AST) Selector(n Node) Selector {
	return Selector{
		lhsExpr:    n.lhs,
		identifier: n.rhs,
	}
}

func Selector_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Selector(ast.nodes[i])
	return []NodeIndex{n.lhsExpr, n.identifier}
}

func Selector_String(ast AST, i NodeIndex) string {
	return "Get"
}

type Call struct {
	lhsExpr   NodeIndex
	arguments NodeIndex
}

func (ast AST) Call(n Node) Call {
	return Call{
		lhsExpr:   n.lhs,
		arguments: n.rhs,
	}
}

func Call_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Call(ast.nodes[i])
	return []NodeIndex{n.lhsExpr, n.arguments}
}

func Call_String(ast AST, i NodeIndex) string {
	return "Call"
}

type Or struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) Or(n Node) Or {
	return Or{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func Or_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Or(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func Or_String(ast AST, i NodeIndex) string {
	return "||"
}

type And struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) And(n Node) And {
	return And{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func And_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.And(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func And_String(ast AST, i NodeIndex) string {
	return "&&"
}

type Equals struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) Equals(n Node) Equals {
	return Equals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func Equals_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Equals(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func Equals_String(ast AST, i NodeIndex) string {
	return "=="
}

type NotEquals struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) NotEquals(n Node) NotEquals {
	return NotEquals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NotEquals_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.NotEquals(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func NotEquals_String(ast AST, i NodeIndex) string {
	return "!="
}

type GreaterThan struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) GreaterThan(n Node) GreaterThan {
	return GreaterThan{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func GreaterThan_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.GreaterThan(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func GreaterThan_String(ast AST, i NodeIndex) string {
	return ">"
}

type LessThan struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) LessThan(n Node) LessThan {
	return LessThan{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func LessThan_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.LessThan(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func LessThan_String(ast AST, i NodeIndex) string {
	return "<"
}

type GreaterThanEquals struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) GreaterThanEquals(n Node) GreaterThanEquals {
	return GreaterThanEquals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func GreaterThanEquals_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.GreaterThanEquals(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func GreaterThanEquals_String(ast AST, i NodeIndex) string {
	return ">="
}

type LessThanEquals struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) LessThanEquals(n Node) LessThanEquals {
	return LessThanEquals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func LessThanEquals_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.LessThanEquals(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func LessThanEquals_String(ast AST, i NodeIndex) string {
	return "<="
}

type BinaryPlus struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) BinaryPlus(n Node) BinaryPlus {
	return BinaryPlus{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func BinaryPlus_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.BinaryPlus(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func BinaryPlus_String(ast AST, i NodeIndex) string {
	return "+"
}

type BinaryMinus struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) BinaryMinus(n Node) BinaryMinus {
	return BinaryMinus{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func BinaryMinus_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.BinaryMinus(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func BinaryMinus_String(ast AST, i NodeIndex) string {
	return "-"
}

type Multiply struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) Multiply(n Node) Multiply {
	return Multiply{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func Multiply_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Multiply(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func Multiply_String(ast AST, i NodeIndex) string {
	return "*"
}

type Divide struct {
	lhs NodeIndex
	rhs NodeIndex
}

func (ast AST) Divide(n Node) Divide {
	return Divide{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func Divide_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Divide(ast.nodes[i])
	return []NodeIndex{n.lhs, n.rhs}
}

func Divide_String(ast AST, i NodeIndex) string {
	return "/"
}

type UnaryPlus struct {
	unary NodeIndex
}

func (ast AST) UnaryPlus(n Node) UnaryPlus {
	return UnaryPlus{
		unary: n.lhs,
	}
}

func UnaryPlus_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.UnaryPlus(ast.nodes[i])
	return []NodeIndex{n.unary}
}

func UnaryPlus_String(ast AST, i NodeIndex) string {
	return "+"
}

type UnaryMinus struct {
	unary NodeIndex
}

func (ast AST) UnaryMinus(n Node) UnaryMinus {
	return UnaryMinus{
		unary: n.lhs,
	}
}

func UnaryMinus_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.UnaryMinus(ast.nodes[i])
	return []NodeIndex{n.unary}
}

func UnaryMinus_String(ast AST, i NodeIndex) string {
	return "-"
}

type Not struct {
	unary NodeIndex
}

func (ast AST) Not(n Node) Not {
	return Not{
		unary: n.lhs,
	}
}

func Not_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.Not(ast.nodes[i])
	return []NodeIndex{n.unary}
}

func Not_String(ast AST, i NodeIndex) string {
	return "!"
}

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

func IdentifierList_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.IdentifierList(ast.nodes[i])
	return n.identifiers
}

func IdentifierList_String(ast AST, i NodeIndex) string {
	ids := IdentifierList_Children(ast, i)
	s := make([]string, 0, len(ids))
	for _, i := range ids {
		s = append(s, Identifier_String(ast, i))
	}
	return strings.Join(s, " ")
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

func ExpressionList_Children(ast AST, i NodeIndex) []NodeIndex {
	n := ast.IdentifierList(ast.nodes[i])
	return n.identifiers
}

func ExpressionList_String(ast AST, i NodeIndex) string {
	return "ExpressionList"
}

type IntLiteral struct {
	token TokenIndex
}

func (ast AST) IntLiteral(n Node) IntLiteral {
	return IntLiteral{
		token: n.tokenIdx,
	}
}

func IntLiteral_Children(ast AST, i NodeIndex) []NodeIndex {
	return []NodeIndex{}
}

func IntLiteral_String(ast AST, i NodeIndex) string {
	n := ast.IntLiteral(ast.nodes[i])
	return ast.src.lexeme(ast.src.token(n.token))
}

type FloatLiteral struct {
	token TokenIndex
}

func (ast AST) FloatLiteral(n Node) FloatLiteral {
	return FloatLiteral{
		token: n.tokenIdx,
	}
}

func FloatLiteral_Children(ast AST, i NodeIndex) []NodeIndex {
	return []NodeIndex{}
}

func FloatLiteral_String(ast AST, i NodeIndex) string {
	n := ast.FloatLiteral(ast.nodes[i])
	return ast.src.lexeme(ast.src.token(n.token))
}

type StringLiteral struct {
	token TokenIndex
}

func (ast AST) StringLiteral(n Node) StringLiteral {
	return StringLiteral{
		token: n.tokenIdx,
	}
}

func StringLiteral_Children(ast AST, i NodeIndex) []NodeIndex {
	return []NodeIndex{}
}

func StringLiteral_String(ast AST, i NodeIndex) string {
	n := ast.StringLiteral(ast.nodes[i])
	return ast.src.lexeme(ast.src.token(n.token))
}

type Identifier struct {
	token TokenIndex
}

func (ast AST) Identifier(n Node) Identifier {
	return Identifier{
		token: n.tokenIdx,
	}
}

func Identifier_Children(ast AST, i NodeIndex) []NodeIndex {
	return []NodeIndex{}
}

func Identifier_String(ast AST, i NodeIndex) string {
	n := ast.Identifier(ast.nodes[i])
	return ast.src.lexeme(ast.src.token(n.token))
}

type AST struct {
	src   *Source
	nodes []Node
	extra []AnyIndex
}

func (ast AST) NodeString(i NodeIndex) string {
	n := ast.nodes[i]
	getString := nodeIndexToString[n.tag]
	if getString == nil {
		panic(fmt.Sprintf("Handler for %d is not set", n.tag))
	}
	return getString(ast, i)
}

type NodeAction = func(*AST, NodeIndex) bool

func (ast *AST) Traverse(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNode(onEnter, onExit, 0)
}

func (ast *AST) traverseNode(onEnter NodeAction, onExit NodeAction, i NodeIndex) {
	if i == NodeIndexUndefined {
		return
	}
	n := ast.nodes[i]
	stopTraversal := onEnter(ast, i)
	defer onExit(ast, i)
	if stopTraversal {
		return
	}

	getChildren := nodeIndexToChildren[n.tag]
	if getChildren == nil {
		panic(fmt.Sprintf("Handler for %d is not set", n.tag))
	}
	children := getChildren(*ast, i)
	for _, c := range children {
		ast.traverseNode(onEnter, onExit, c)
	}
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
	onEnter := func(ast *AST, i NodeIndex) (stopTraversal bool) {
		str.WriteByte('(')
		str.WriteString(ast.NodeString(i))

		// filter nodes that are composite by themselves
		n := ast.nodes[i]
		stopTraversal = n.tag == NodeIdentifierList
		return
	}
	onExit := func(ast *AST, i NodeIndex) (stopTraversal bool) {
		str.WriteByte(')')
		return false
	}
	ast.Traverse(onEnter, onExit)

	if doFormat {
		return formatSExpr(str.String())
	}
	return str.String()
}
