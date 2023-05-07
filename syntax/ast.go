package syntax

import (
	"fmt"
	"math"
	"strings"
)

type nodeTag int
type nodeIndex int
type nodeAction = func(*AST, nodeIndex) bool

const (
	nodeIndexInvalid   nodeIndex = math.MinInt
	nodeIndexUndefined           = -1
)

const (
	nodeUndefined = -1
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
	NodeMax
)

// Note: These three tables (constructor, string and children) handlers
// can be very convenient and robust at compile time if language has compile
// time introspection / macro system. But the latter assumes AST transformation,
// what is not the point here - we just want to fill the maps with handlers, that
// named specifically, that's all
var NodeConstructor = [...]func(tokenIndex, nodeIndex, nodeIndex) Node{
	NodeSource: NewSourceRoot,
	NodeBlock:  NewBlock,

	NodeFunctionDecl: NewFunctionDecl,
	NodeSignature:    NewSignature,
	NodeConstDecl:    NewConstDecl,
	NodeAssignment:   NewAssignment,

	NodeSelector: NewSelector,
	NodeCall:     NewCall,

	NodeOr:                NewOr,
	NodeAnd:               NewAnd,
	NodeEquals:            NewEquals,
	NodeNotEquals:         NewNotEquals,
	NodeGreaterThan:       NewGreaterThan,
	NodeLessThan:          NewLessThan,
	NodeGreaterThanEquals: NewGreaterThanEquals,
	NodeLessThanEquals:    NewLessThanEquals,
	NodeBinaryPlus:        NewBinaryPlus,
	NodeBinaryMinus:       NewBinaryMinus,
	NodeMultiply:          NewMultiply,
	NodeDivide:            NewDivide,

	NodeUnaryPlus:  NewUnaryPlus,
	NodeUnaryMinus: NewUnaryMinus,
	NodeNot:        NewNot,

	NodeIntLiteral:    NewIntLiteral,
	NodeFloatLiteral:  NewFloatLiteral,
	NodeStringLiteral: NewStringLiteral,
	NodeIdentifier:    NewIdentifier,

	NodeIdentifierList: NewIdentifierList,
	NodeExpressionList: NewExpressionList,
}

var NodeString = [...]func(AST, nodeIndex) string{
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

var NodeChildren = [...]func(AST, nodeIndex) []nodeIndex{
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

func init() {
	for i := 0; i < NodeMax; i++ {
		{
			h := NodeConstructor[i]
			if h == nil {
				panic(fmt.Sprintf("Node tag to constructor table is not full, failed at %d", i))
			}
		}
		{
			h := NodeChildren[i]
			if h == nil {
				panic(fmt.Sprintf("Node tag to children table is not full, failed at %d", i))
			}
		}
		{
			h := NodeString[i]
			if h == nil {
				panic(fmt.Sprintf("Node tag to string table is not full, failed at %d", i))
			}
		}
	}
}

type Node struct {
	tag      nodeTag
	tokenIdx tokenIndex
	lhs, rhs nodeIndex
}

// General pattern of typed nodes:
// Full and interpreted struct data (typed node)
// Untyped node constructor
// Function to convert from node to typed node
// Function to convert node to string
// Function to get node children

type SourceRoot struct {
	declarations []nodeIndex
}

func (ast AST) SourceRoot(n Node) SourceRoot {
	decls := make([]nodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := nodeIndex(ast.extra[i])
		decls = append(decls, c_i)
	}
	return SourceRoot{
		declarations: decls,
	}
}

func NewSourceRoot(rootToken tokenIndex, start nodeIndex, end nodeIndex) Node {
	return Node{
		tag:      NodeSource,
		tokenIdx: rootToken,
		lhs:      start,
		rhs:      end,
	}
}

func SourceRoot_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.SourceRoot(ast.nodes[i])
	return n.declarations
}

func SourceRoot_String(ast AST, i nodeIndex) string {
	return "Source"
}

type FunctionDecl struct {
	name      nodeIndex
	signature nodeIndex
	body      nodeIndex
}

func (ast AST) FunctionDecl(n Node) FunctionDecl {
	node := FunctionDecl{}
	// find identifier node by it's token index
	node.name = nodeIndexInvalid
	for i := range ast.nodes {
		id := ast.nodes[i]
		if n.tokenIdx == id.tokenIdx &&
			id.tag == NodeIdentifier {
			node.name = nodeIndex(i)
		}
	}
	if node.name == nodeIndexInvalid {
		panic("This shouldn't have happened!")
	}
	node.signature = n.lhs
	node.body = n.rhs
	return node
}

func NewFunctionDecl(tokenIdx tokenIndex, signature nodeIndex, body nodeIndex) Node {
	return Node{
		tag:      NodeFunctionDecl,
		tokenIdx: tokenIdx,
		lhs:      signature,
		rhs:      body,
	}
}

func FunctionDecl_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.FunctionDecl(ast.nodes[i])
	return []nodeIndex{n.name, n.signature, n.body}
}

func FunctionDecl_String(ast AST, i nodeIndex) string {
	return "FunctionDecl"
}

type Signature struct {
	parameters nodeIndex
}

func (ast AST) Signature(n Node) Signature {
	return Signature{
		parameters: n.lhs,
	}
}

func NewSignature(tokenIdx tokenIndex, parameters nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeSignature,
		tokenIdx: tokenIdx,
		lhs:      parameters,
		rhs:      rhs,
	}
}

func Signature_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Signature(ast.nodes[i])
	return []nodeIndex{n.parameters}
}

func Signature_String(ast AST, i nodeIndex) string {
	return "Signature"
}

type Block struct {
	statements []nodeIndex
}

func (ast AST) Block(n Node) Block {
	statements := make([]nodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := nodeIndex(ast.extra[i])
		statements = append(statements, c_i)
	}
	return Block{
		statements: statements,
	}
}

func NewBlock(tokenIdx tokenIndex, start nodeIndex, end nodeIndex) Node {
	return Node{
		tag:      NodeBlock,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}
}

func Block_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Block(ast.nodes[i])
	return n.statements
}

func Block_String(ast AST, i nodeIndex) string {
	return "Block"
}

type ConstDecl struct {
	identifierList nodeIndex
	expressionList nodeIndex
}

func (ast AST) ConstDecl(n Node) ConstDecl {
	return ConstDecl{
		identifierList: n.lhs,
		expressionList: n.rhs,
	}
}

func NewConstDecl(tokenIdx tokenIndex, identifierList nodeIndex, expressionList nodeIndex) Node {
	return Node{
		tag:      NodeConstDecl,
		tokenIdx: tokenIdx,
		lhs:      identifierList,
		rhs:      expressionList,
	}
}

func ConstDecl_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.ConstDecl(ast.nodes[i])
	return []nodeIndex{n.identifierList, n.expressionList}
}

func ConstDecl_String(ast AST, i nodeIndex) string {
	return "ConstDecl"
}

type Assignment struct {
	lhsList nodeIndex
	rhsList nodeIndex
}

func (ast AST) Assignment(n Node) Assignment {
	return Assignment{
		lhsList: n.lhs,
		rhsList: n.rhs,
	}
}

func NewAssignment(tokenIdx tokenIndex, exprList1 nodeIndex, exprList2 nodeIndex) Node {
	return Node{
		tag:      NodeAssignment,
		tokenIdx: tokenIdx,
		lhs:      exprList1,
		rhs:      exprList2,
	}
}

func Assignment_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Assignment(ast.nodes[i])
	return []nodeIndex{n.lhsList, n.rhsList}
}

func Assignment_String(ast AST, i nodeIndex) string {
	return "Assign"
}

type Selector struct {
	lhsExpr    nodeIndex
	identifier nodeIndex
}

func (ast AST) Selector(n Node) Selector {
	return Selector{
		lhsExpr:    n.lhs,
		identifier: n.rhs,
	}
}

func NewSelector(tokenIdx tokenIndex, expr nodeIndex, identifier nodeIndex) Node {
	return Node{
		tag:      NodeSelector,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      identifier,
	}
}

func Selector_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Selector(ast.nodes[i])
	return []nodeIndex{n.lhsExpr, n.identifier}
}

func Selector_String(ast AST, i nodeIndex) string {
	return "Get"
}

type Call struct {
	lhsExpr   nodeIndex
	arguments nodeIndex
}

func (ast AST) Call(n Node) Call {
	return Call{
		lhsExpr:   n.lhs,
		arguments: n.rhs,
	}
}

func NewCall(tokenIdx tokenIndex, expr nodeIndex, args nodeIndex) Node {
	return Node{
		tag:      NodeCall,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      args,
	}

}

func Call_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Call(ast.nodes[i])
	return []nodeIndex{n.lhsExpr, n.arguments}
}

func Call_String(ast AST, i nodeIndex) string {
	return "Call"
}

type Or struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) Or(n Node) Or {
	return Or{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewOr(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeOr,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Or_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Or(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func Or_String(ast AST, i nodeIndex) string {
	return "||"
}

type And struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) And(n Node) And {
	return And{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewAnd(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeAnd,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func And_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.And(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func And_String(ast AST, i nodeIndex) string {
	return "&&"
}

type Equals struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) Equals(n Node) Equals {
	return Equals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewEquals(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Equals_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Equals(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func Equals_String(ast AST, i nodeIndex) string {
	return "=="
}

type NotEquals struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) NotEquals(n Node) NotEquals {
	return NotEquals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewNotEquals(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeNotEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func NotEquals_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.NotEquals(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func NotEquals_String(ast AST, i nodeIndex) string {
	return "!="
}

type GreaterThan struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) GreaterThan(n Node) GreaterThan {
	return GreaterThan{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewGreaterThan(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeGreaterThan,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func GreaterThan_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.GreaterThan(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func GreaterThan_String(ast AST, i nodeIndex) string {
	return ">"
}

type LessThan struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) LessThan(n Node) LessThan {
	return LessThan{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewLessThan(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeLessThan,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func LessThan_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.LessThan(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func LessThan_String(ast AST, i nodeIndex) string {
	return "<"
}

type GreaterThanEquals struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) GreaterThanEquals(n Node) GreaterThanEquals {
	return GreaterThanEquals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewGreaterThanEquals(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeGreaterThanEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func GreaterThanEquals_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.GreaterThanEquals(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func GreaterThanEquals_String(ast AST, i nodeIndex) string {
	return ">="
}

type LessThanEquals struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) LessThanEquals(n Node) LessThanEquals {
	return LessThanEquals{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewLessThanEquals(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeLessThanEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func LessThanEquals_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.LessThanEquals(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func LessThanEquals_String(ast AST, i nodeIndex) string {
	return "<="
}

type BinaryPlus struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) BinaryPlus(n Node) BinaryPlus {
	return BinaryPlus{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewBinaryPlus(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeBinaryPlus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BinaryPlus_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.BinaryPlus(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func BinaryPlus_String(ast AST, i nodeIndex) string {
	return "+"
}

type BinaryMinus struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) BinaryMinus(n Node) BinaryMinus {
	return BinaryMinus{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewBinaryMinus(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeBinaryMinus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BinaryMinus_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.BinaryMinus(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func BinaryMinus_String(ast AST, i nodeIndex) string {
	return "-"
}

type Multiply struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) Multiply(n Node) Multiply {
	return Multiply{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewMultiply(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeMultiply,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Multiply_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Multiply(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func Multiply_String(ast AST, i nodeIndex) string {
	return "*"
}

type Divide struct {
	lhs nodeIndex
	rhs nodeIndex
}

func (ast AST) Divide(n Node) Divide {
	return Divide{
		lhs: n.lhs,
		rhs: n.rhs,
	}
}

func NewDivide(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeDivide,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Divide_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Divide(ast.nodes[i])
	return []nodeIndex{n.lhs, n.rhs}
}

func Divide_String(ast AST, i nodeIndex) string {
	return "/"
}

type UnaryPlus struct {
	unary nodeIndex
}

func (ast AST) UnaryPlus(n Node) UnaryPlus {
	return UnaryPlus{
		unary: n.lhs,
	}
}

func NewUnaryPlus(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeUnaryPlus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func UnaryPlus_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.UnaryPlus(ast.nodes[i])
	return []nodeIndex{n.unary}
}

func UnaryPlus_String(ast AST, i nodeIndex) string {
	return "+"
}

type UnaryMinus struct {
	unary nodeIndex
}

func (ast AST) UnaryMinus(n Node) UnaryMinus {
	return UnaryMinus{
		unary: n.lhs,
	}
}

func NewUnaryMinus(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeUnaryMinus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func UnaryMinus_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.UnaryMinus(ast.nodes[i])
	return []nodeIndex{n.unary}
}

func UnaryMinus_String(ast AST, i nodeIndex) string {
	return "-"
}

type Not struct {
	unary nodeIndex
}

func (ast AST) Not(n Node) Not {
	return Not{
		unary: n.lhs,
	}
}
func NewNot(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeNot,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Not_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.Not(ast.nodes[i])
	return []nodeIndex{n.unary}
}

func Not_String(ast AST, i nodeIndex) string {
	return "!"
}

type IdentifierList struct {
	identifiers []nodeIndex
}

func (ast AST) IdentifierList(n Node) IdentifierList {
	ids := make([]nodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		ids = append(ids, nodeIndex(c_i))
	}
	return IdentifierList{
		identifiers: ids,
	}
}
func NewIdentifierList(tokenIdx tokenIndex, start nodeIndex, end nodeIndex) Node {
	return Node{
		tag:      NodeIdentifierList,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}

}

func IdentifierList_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.IdentifierList(ast.nodes[i])
	return n.identifiers
}

func IdentifierList_String(ast AST, i nodeIndex) string {
	ids := IdentifierList_Children(ast, i)
	s := make([]string, 0, len(ids))
	for _, i := range ids {
		s = append(s, Identifier_String(ast, i))
	}
	return strings.Join(s, " ")
}

type ExpressionList struct {
	expressions []nodeIndex
}

func (ast AST) ExpressionList(n Node) ExpressionList {
	exprs := make([]nodeIndex, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		exprs = append(exprs, nodeIndex(c_i))
	}
	return ExpressionList{
		expressions: exprs,
	}
}
func NewExpressionList(tokenIdx tokenIndex, start nodeIndex, end nodeIndex) Node {
	return Node{
		tag:      NodeExpressionList,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}

}

func ExpressionList_Children(ast AST, i nodeIndex) []nodeIndex {
	n := ast.IdentifierList(ast.nodes[i])
	return n.identifiers
}

func ExpressionList_String(ast AST, i nodeIndex) string {
	return "ExpressionList"
}

type IntLiteral struct {
	token tokenIndex
}

func (ast AST) IntLiteral(n Node) IntLiteral {
	return IntLiteral{
		token: n.tokenIdx,
	}
}
func NewIntLiteral(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeIntLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func IntLiteral_Children(ast AST, i nodeIndex) []nodeIndex {
	return []nodeIndex{}
}

func IntLiteral_String(ast AST, i nodeIndex) string {
	n := ast.IntLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.token)
}

type FloatLiteral struct {
	token tokenIndex
}

func (ast AST) FloatLiteral(n Node) FloatLiteral {
	return FloatLiteral{
		token: n.tokenIdx,
	}
}
func NewFloatLiteral(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeFloatLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func FloatLiteral_Children(ast AST, i nodeIndex) []nodeIndex {
	return []nodeIndex{}
}

func FloatLiteral_String(ast AST, i nodeIndex) string {
	n := ast.FloatLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.token)
}

type StringLiteral struct {
	token tokenIndex
}

func (ast AST) StringLiteral(n Node) StringLiteral {
	return StringLiteral{
		token: n.tokenIdx,
	}
}
func NewStringLiteral(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeStringLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func StringLiteral_Children(ast AST, i nodeIndex) []nodeIndex {
	return []nodeIndex{}
}

func StringLiteral_String(ast AST, i nodeIndex) string {
	n := ast.StringLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.token)
}

type Identifier struct {
	token tokenIndex
}

func (ast AST) Identifier(n Node) Identifier {
	return Identifier{
		token: n.tokenIdx,
	}
}
func NewIdentifier(tokenIdx tokenIndex, lhs nodeIndex, rhs nodeIndex) Node {
	return Node{
		tag:      NodeIdentifier,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Identifier_Children(ast AST, i nodeIndex) []nodeIndex {
	return []nodeIndex{}
}

func Identifier_String(ast AST, i nodeIndex) string {
	n := ast.Identifier(ast.nodes[i])
	return ast.src.Lexeme(n.token)
}

type AST struct {
	src   *source
	nodes []Node

	// NOTE: That whole thing about "extra" like in zig compiler
	// is different here - my "extra" stores arbitrary number of indicies for any
	// particular node, in contrast, zig's "extra" stores compile time known
	// fixed number of indicies (for any node), and that number depends on the node type
	// latter is more plausable, bc it's more versatile and basically superset
	// of my implementation. Well, will stick to current implemenation...
	extra []anyIndex
}

func NewAST(src *source) AST {
	return AST{src: src}
}

func (ast AST) GetNodeString(i nodeIndex) string {
	n := ast.nodes[i]
	getString := NodeString[n.tag]
	return getString(ast, i)
}

func (ast *AST) Traverse(onEnter nodeAction, onExit nodeAction) {
	ast.traverseNode(onEnter, onExit, 0)
}

func (ast *AST) Dump(doFormat bool) string {
	str := strings.Builder{}
	onEnter := func(ast *AST, i nodeIndex) (stopTraversal bool) {
		str.WriteByte('(')
		str.WriteString(ast.GetNodeString(i))

		// filter nodes that are composite by themselves
		n := ast.nodes[i]
		stopTraversal = n.tag == NodeIdentifierList
		return
	}
	onExit := func(ast *AST, i nodeIndex) (stopTraversal bool) {
		str.WriteByte(')')
		return false
	}
	ast.Traverse(onEnter, onExit)

	if doFormat {
		return formatSExpr(str.String())
	}
	return str.String()
}

func (ast *AST) traverseNode(onEnter nodeAction, onExit nodeAction, i nodeIndex) {
	if i == nodeIndexUndefined {
		return
	}
	n := ast.nodes[i]
	stopTraversal := onEnter(ast, i)
	defer onExit(ast, i)
	if stopTraversal {
		return
	}

	getChildren := NodeChildren[n.tag]
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
