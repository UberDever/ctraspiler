package syntax

import (
	"fmt"
	"math"
	"strings"
)

type NodeTag int
type NodeID int
type NodeAction = func(*AST, NodeID) bool

const (
	NodeIDInvalid   NodeID = math.MinInt
	NodeIDUndefined        = -1
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
	NodeReturnStmt

	NodeExpression
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
var NodeConstructor = [...]func(TokenID, NodeID, NodeID) Node{
	NodeSource: NewSourceRoot,
	NodeBlock:  NewBlock,

	NodeFunctionDecl: NewFunctionDecl,
	NodeSignature:    NewSignature,
	NodeConstDecl:    NewConstDecl,
	NodeAssignment:   NewAssignment,
	NodeReturnStmt:   NewReturnStmt,

	NodeExpression: NewExpression,
	NodeSelector:   NewSelector,
	NodeCall:       NewCall,

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

var NodeString = [...]func(AST, NodeID) string{
	NodeSource: SourceRoot_String,
	NodeBlock:  Block_String,

	NodeFunctionDecl: FunctionDecl_String,
	NodeSignature:    Signature_String,
	NodeConstDecl:    ConstDecl_String,
	NodeAssignment:   Assignment_String,
	NodeReturnStmt:   ReturnStmt_String,

	NodeExpression: Expression_String,
	NodeSelector:   Selector_String,
	NodeCall:       Call_String,

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

var NodeChildren = [...]func(AST, NodeID) []NodeID{
	NodeSource: SourceRoot_Children,
	NodeBlock:  Block_Children,

	NodeFunctionDecl: FunctionDecl_Children,
	NodeSignature:    Signature_Children,
	NodeConstDecl:    ConstDecl_Children,
	NodeAssignment:   Assignment_Children,
	NodeReturnStmt:   ReturnStmt_Children,

	NodeExpression: Expression_Children,
	NodeSelector:   Selector_Children,
	NodeCall:       Call_Children,

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
	tag      NodeTag
	tokenIdx TokenID
	lhs, rhs NodeID
}

func (n Node) Tag() NodeTag   { return n.tag }
func (n Node) Token() TokenID { return n.tokenIdx }

// General pattern of typed nodes:
// Full and interpreted struct data (typed node)
// Untyped node constructor
// Function to convert from node to typed node
// Function to convert node to string
// Function to get node children

// TODO: Add to every typed node types of it's children
type SourceRoot struct {
	Declarations []NodeID
}

func (ast AST) SourceRoot(n Node) SourceRoot {
	decls := make([]NodeID, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := NodeID(ast.extra[i])
		decls = append(decls, c_i)
	}
	return SourceRoot{
		Declarations: decls,
	}
}

func NewSourceRoot(rootToken TokenID, start NodeID, end NodeID) Node {
	return Node{
		tag:      NodeSource,
		tokenIdx: rootToken,
		lhs:      start,
		rhs:      end,
	}
}

func SourceRoot_Children(ast AST, i NodeID) []NodeID {
	n := ast.SourceRoot(ast.nodes[i])
	return n.Declarations
}

func SourceRoot_String(ast AST, i NodeID) string {
	return "Source"
}

type FunctionDecl struct {
	Name      NodeID
	Signature NodeID
	Body      NodeID
}

func (ast AST) FunctionDecl(n Node) FunctionDecl {
	node := FunctionDecl{}
	node.Name = n.lhs
	extra := n.rhs
	node.Signature = NodeID(ast.extra[extra])
	node.Body = NodeID(ast.extra[extra+1])

	return node
}

func NewFunctionDecl(tokenIdx TokenID, signature NodeID, body NodeID) Node {
	return Node{
		tag:      NodeFunctionDecl,
		tokenIdx: tokenIdx,
		lhs:      signature,
		rhs:      body,
	}
}

func FunctionDecl_Children(ast AST, i NodeID) []NodeID {
	n := ast.FunctionDecl(ast.nodes[i])
	return []NodeID{n.Name, n.Signature, n.Body}
}

func FunctionDecl_String(ast AST, i NodeID) string {
	return "FunctionDecl"
}

type Signature struct {
	Parameters NodeID
}

func (ast AST) Signature(n Node) Signature {
	return Signature{
		Parameters: n.lhs,
	}
}

func NewSignature(tokenIdx TokenID, parameters NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeSignature,
		tokenIdx: tokenIdx,
		lhs:      parameters,
		rhs:      rhs,
	}
}

func Signature_Children(ast AST, i NodeID) []NodeID {
	n := ast.Signature(ast.nodes[i])
	return []NodeID{n.Parameters}
}

func Signature_String(ast AST, i NodeID) string {
	return "Signature"
}

type Block struct {
	Statements []NodeID
}

func (ast AST) Block(n Node) Block {
	statements := make([]NodeID, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := NodeID(ast.extra[i])
		statements = append(statements, c_i)
	}
	return Block{
		Statements: statements,
	}
}

func NewBlock(tokenIdx TokenID, start NodeID, end NodeID) Node {
	return Node{
		tag:      NodeBlock,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}
}

func Block_Children(ast AST, i NodeID) []NodeID {
	n := ast.Block(ast.nodes[i])
	return n.Statements
}

func Block_String(ast AST, i NodeID) string {
	return "Block"
}

type ConstDecl struct {
	IdentifierList NodeID
	ExpressionList NodeID
}

func (ast AST) ConstDecl(n Node) ConstDecl {
	return ConstDecl{
		IdentifierList: n.lhs,
		ExpressionList: n.rhs,
	}
}

func NewConstDecl(tokenIdx TokenID, identifierList NodeID, expressionList NodeID) Node {
	return Node{
		tag:      NodeConstDecl,
		tokenIdx: tokenIdx,
		lhs:      identifierList,
		rhs:      expressionList,
	}
}

func ConstDecl_Children(ast AST, i NodeID) []NodeID {
	n := ast.ConstDecl(ast.nodes[i])
	return []NodeID{n.IdentifierList, n.ExpressionList}
}

func ConstDecl_String(ast AST, i NodeID) string {
	return "ConstDecl"
}

type Assignment struct {
	LhsList NodeID
	RhsList NodeID
}

func (ast AST) Assignment(n Node) Assignment {
	return Assignment{
		LhsList: n.lhs,
		RhsList: n.rhs,
	}
}

func NewAssignment(tokenIdx TokenID, exprList1 NodeID, exprList2 NodeID) Node {
	return Node{
		tag:      NodeAssignment,
		tokenIdx: tokenIdx,
		lhs:      exprList1,
		rhs:      exprList2,
	}
}

func Assignment_Children(ast AST, i NodeID) []NodeID {
	n := ast.Assignment(ast.nodes[i])
	return []NodeID{n.LhsList, n.RhsList}
}

func Assignment_String(ast AST, i NodeID) string {
	return "Assign"
}

type ReturnStmt struct {
	ExpressionList NodeID
}

func (ast AST) ReturnStmt(n Node) ReturnStmt {
	return ReturnStmt{
		ExpressionList: n.lhs,
	}
}

func NewReturnStmt(tokenIdx TokenID, expressions NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeReturnStmt,
		tokenIdx: tokenIdx,
		lhs:      expressions,
		rhs:      rhs,
	}
}

func ReturnStmt_Children(ast AST, i NodeID) []NodeID {
	n := ast.ReturnStmt(ast.nodes[i])
	return []NodeID{n.ExpressionList}
}

func ReturnStmt_String(ast AST, i NodeID) string {
	return "Return"
}

type Expression struct {
	Expression NodeID
}

func (ast AST) Expression(n Node) Expression {
	return Expression{
		Expression: n.lhs,
	}
}

func NewExpression(tokenIdx TokenID, expr NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeExpression,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      rhs,
	}
}

func Expression_Children(ast AST, i NodeID) []NodeID {
	n := ast.Expression(ast.nodes[i])
	return []NodeID{n.Expression}
}

func Expression_String(ast AST, i NodeID) string {
	return "Expr"
}

type Selector struct {
	LhsExpr    NodeID
	Identifier NodeID
}

func (ast AST) Selector(n Node) Selector {
	return Selector{
		LhsExpr:    n.lhs,
		Identifier: n.rhs,
	}
}

func NewSelector(tokenIdx TokenID, expr NodeID, identifier NodeID) Node {
	return Node{
		tag:      NodeSelector,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      identifier,
	}
}

func Selector_Children(ast AST, i NodeID) []NodeID {
	n := ast.Selector(ast.nodes[i])
	return []NodeID{n.LhsExpr, n.Identifier}
}

func Selector_String(ast AST, i NodeID) string {
	return "Get"
}

type Call struct {
	LhsExpr   NodeID
	Arguments NodeID
}

func (ast AST) Call(n Node) Call {
	return Call{
		LhsExpr:   n.lhs,
		Arguments: n.rhs,
	}
}

func NewCall(tokenIdx TokenID, expr NodeID, args NodeID) Node {
	return Node{
		tag:      NodeCall,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      args,
	}

}

func Call_Children(ast AST, i NodeID) []NodeID {
	n := ast.Call(ast.nodes[i])
	return []NodeID{n.LhsExpr, n.Arguments}
}

func Call_String(ast AST, i NodeID) string {
	return "Call"
}

type Or struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) Or(n Node) Or {
	return Or{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewOr(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeOr,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Or_Children(ast AST, i NodeID) []NodeID {
	n := ast.Or(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func Or_String(ast AST, i NodeID) string {
	return "||"
}

type And struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) And(n Node) And {
	return And{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewAnd(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeAnd,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func And_Children(ast AST, i NodeID) []NodeID {
	n := ast.And(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func And_String(ast AST, i NodeID) string {
	return "&&"
}

type Equals struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) Equals(n Node) Equals {
	return Equals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewEquals(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Equals_Children(ast AST, i NodeID) []NodeID {
	n := ast.Equals(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func Equals_String(ast AST, i NodeID) string {
	return "=="
}

type NotEquals struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) NotEquals(n Node) NotEquals {
	return NotEquals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewNotEquals(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeNotEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func NotEquals_Children(ast AST, i NodeID) []NodeID {
	n := ast.NotEquals(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func NotEquals_String(ast AST, i NodeID) string {
	return "!="
}

type GreaterThan struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) GreaterThan(n Node) GreaterThan {
	return GreaterThan{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewGreaterThan(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeGreaterThan,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func GreaterThan_Children(ast AST, i NodeID) []NodeID {
	n := ast.GreaterThan(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func GreaterThan_String(ast AST, i NodeID) string {
	return ">"
}

type LessThan struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) LessThan(n Node) LessThan {
	return LessThan{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewLessThan(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeLessThan,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func LessThan_Children(ast AST, i NodeID) []NodeID {
	n := ast.LessThan(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func LessThan_String(ast AST, i NodeID) string {
	return "<"
}

type GreaterThanEquals struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) GreaterThanEquals(n Node) GreaterThanEquals {
	return GreaterThanEquals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewGreaterThanEquals(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeGreaterThanEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func GreaterThanEquals_Children(ast AST, i NodeID) []NodeID {
	n := ast.GreaterThanEquals(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func GreaterThanEquals_String(ast AST, i NodeID) string {
	return ">="
}

type LessThanEquals struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) LessThanEquals(n Node) LessThanEquals {
	return LessThanEquals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewLessThanEquals(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeLessThanEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func LessThanEquals_Children(ast AST, i NodeID) []NodeID {
	n := ast.LessThanEquals(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func LessThanEquals_String(ast AST, i NodeID) string {
	return "<="
}

type BinaryPlus struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) BinaryPlus(n Node) BinaryPlus {
	return BinaryPlus{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewBinaryPlus(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeBinaryPlus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BinaryPlus_Children(ast AST, i NodeID) []NodeID {
	n := ast.BinaryPlus(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func BinaryPlus_String(ast AST, i NodeID) string {
	return "+"
}

type BinaryMinus struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) BinaryMinus(n Node) BinaryMinus {
	return BinaryMinus{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewBinaryMinus(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeBinaryMinus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BinaryMinus_Children(ast AST, i NodeID) []NodeID {
	n := ast.BinaryMinus(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func BinaryMinus_String(ast AST, i NodeID) string {
	return "-"
}

type Multiply struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) Multiply(n Node) Multiply {
	return Multiply{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewMultiply(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeMultiply,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Multiply_Children(ast AST, i NodeID) []NodeID {
	n := ast.Multiply(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func Multiply_String(ast AST, i NodeID) string {
	return "*"
}

type Divide struct {
	Lhs NodeID
	Rhs NodeID
}

func (ast AST) Divide(n Node) Divide {
	return Divide{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewDivide(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeDivide,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Divide_Children(ast AST, i NodeID) []NodeID {
	n := ast.Divide(ast.nodes[i])
	return []NodeID{n.Lhs, n.Rhs}
}

func Divide_String(ast AST, i NodeID) string {
	return "/"
}

type UnaryPlus struct {
	Unary NodeID
}

func (ast AST) UnaryPlus(n Node) UnaryPlus {
	return UnaryPlus{
		Unary: n.lhs,
	}
}

func NewUnaryPlus(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeUnaryPlus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func UnaryPlus_Children(ast AST, i NodeID) []NodeID {
	n := ast.UnaryPlus(ast.nodes[i])
	return []NodeID{n.Unary}
}

func UnaryPlus_String(ast AST, i NodeID) string {
	return "+"
}

type UnaryMinus struct {
	Unary NodeID
}

func (ast AST) UnaryMinus(n Node) UnaryMinus {
	return UnaryMinus{
		Unary: n.lhs,
	}
}

func NewUnaryMinus(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeUnaryMinus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func UnaryMinus_Children(ast AST, i NodeID) []NodeID {
	n := ast.UnaryMinus(ast.nodes[i])
	return []NodeID{n.Unary}
}

func UnaryMinus_String(ast AST, i NodeID) string {
	return "-"
}

type Not struct {
	Unary NodeID
}

func (ast AST) Not(n Node) Not {
	return Not{
		Unary: n.lhs,
	}
}
func NewNot(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeNot,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Not_Children(ast AST, i NodeID) []NodeID {
	n := ast.Not(ast.nodes[i])
	return []NodeID{n.Unary}
}

func Not_String(ast AST, i NodeID) string {
	return "!"
}

type IdentifierList struct {
	Identifiers []NodeID
}

func (ast AST) IdentifierList(n Node) IdentifierList {
	ids := make([]NodeID, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		ids = append(ids, NodeID(c_i))
	}
	return IdentifierList{
		Identifiers: ids,
	}
}
func NewIdentifierList(tokenIdx TokenID, start NodeID, end NodeID) Node {
	return Node{
		tag:      NodeIdentifierList,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}

}

func IdentifierList_Children(ast AST, i NodeID) []NodeID {
	n := ast.IdentifierList(ast.nodes[i])
	return n.Identifiers
}

func IdentifierList_String(ast AST, i NodeID) string {
	ids := IdentifierList_Children(ast, i)
	s := make([]string, 0, len(ids))
	for _, i := range ids {
		s = append(s, Identifier_String(ast, i))
	}
	return strings.Join(s, " ")
}

type ExpressionList struct {
	Expressions []NodeID
}

func (ast AST) ExpressionList(n Node) ExpressionList {
	exprs := make([]NodeID, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		exprs = append(exprs, NodeID(c_i))
	}
	return ExpressionList{
		Expressions: exprs,
	}
}

func NewExpressionList(tokenIdx TokenID, start NodeID, end NodeID) Node {
	return Node{
		tag:      NodeExpressionList,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}

}

func ExpressionList_Children(ast AST, i NodeID) []NodeID {
	n := ast.IdentifierList(ast.nodes[i])
	return n.Identifiers
}

func ExpressionList_String(ast AST, i NodeID) string {
	return "Expr[]"
}

type IntLiteral struct {
	Token TokenID
}

func (ast AST) IntLiteral(n Node) IntLiteral {
	return IntLiteral{
		Token: n.tokenIdx,
	}
}
func NewIntLiteral(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeIntLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func IntLiteral_Children(ast AST, i NodeID) []NodeID {
	return []NodeID{}
}

func IntLiteral_String(ast AST, i NodeID) string {
	n := ast.IntLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type FloatLiteral struct {
	Token TokenID
}

func (ast AST) FloatLiteral(n Node) FloatLiteral {
	return FloatLiteral{
		Token: n.tokenIdx,
	}
}
func NewFloatLiteral(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeFloatLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func FloatLiteral_Children(ast AST, i NodeID) []NodeID {
	return []NodeID{}
}

func FloatLiteral_String(ast AST, i NodeID) string {
	n := ast.FloatLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type StringLiteral struct {
	Token TokenID
}

func (ast AST) StringLiteral(n Node) StringLiteral {
	return StringLiteral{
		Token: n.tokenIdx,
	}
}
func NewStringLiteral(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeStringLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func StringLiteral_Children(ast AST, i NodeID) []NodeID {
	return []NodeID{}
}

func StringLiteral_String(ast AST, i NodeID) string {
	n := ast.StringLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type Identifier struct {
	Token TokenID
}

func (ast AST) Identifier(n Node) Identifier {
	return Identifier{
		Token: n.tokenIdx,
	}
}
func NewIdentifier(tokenIdx TokenID, lhs NodeID, rhs NodeID) Node {
	return Node{
		tag:      NodeIdentifier,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Identifier_Children(ast AST, i NodeID) []NodeID {
	return []NodeID{}
}

func Identifier_String(ast AST, i NodeID) string {
	n := ast.Identifier(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type AST struct {
	src   *Source
	nodes []Node

	// NOTE: That whole thing about "extra" like in zig compiler
	// is different here - my "extra" stores arbitrary number of indicies for any
	// particular node, in contrast, zig's "extra" stores compile time known
	// fixed number of indicies (for any node), and that number depends on the node type
	// latter is more plausable, bc it's more versatile and basically superset
	// of my implementation. Well, will stick to current implemenation...
	extra []anyIndex
}

func NewAST(src *Source) AST {
	return AST{src: src}
}

func (ast AST) GetNode(i NodeID) Node {
	return ast.nodes[i]
}

func (ast AST) GetNodeString(i NodeID) string {
	n := ast.nodes[i]
	getString := NodeString[n.tag]
	return getString(ast, i)
}

func (ast *AST) Traverse(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNode(onEnter, onExit, 0)
}

func (ast *AST) Dump(doFormat bool) string {
	str := strings.Builder{}
	onEnter := func(ast *AST, i NodeID) (stopTraversal bool) {
		str.WriteByte('(')
		str.WriteString(ast.GetNodeString(i))

		// filter nodes that are composite by themselves
		n := ast.nodes[i]
		stopTraversal = n.tag == NodeIdentifierList
		return
	}
	onExit := func(ast *AST, i NodeID) (stopTraversal bool) {
		str.WriteByte(')')
		return false
	}
	ast.Traverse(onEnter, onExit)

	if doFormat {
		return formatSExpr(str.String())
	}
	return str.String()
}

func (ast *AST) traverseNode(onEnter NodeAction, onExit NodeAction, i NodeID) {
	if i == NodeIDUndefined {
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
