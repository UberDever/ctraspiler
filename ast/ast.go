package ast

import (
	"fmt"
	ID "some/domain"
	s "some/syntax"
	T "some/typesystem"
	"strings"
)

type NodeTag int
type NodeAction = func(*AST, ID.Node) bool

// Note: These three tables (constructor, string and children) handlers
// can be very convenient and robust at compile time if language has compile
// time introspection / macro system. But the latter assumes AST transformation,
// what is not the point here - we just want to fill the maps with handlers, that
// named specifically, that's all
var NodeConstructor = [...]func(ID.Token, ID.Node, ID.Node) Node{
	ID.NodeSource: NewSourceRoot,
	ID.NodeBlock:  NewBlock,

	ID.NodeFunctionDecl: NewFunctionDecl,
	ID.NodeSignature:    NewSignature,
	ID.NodeConstDecl:    NewConstDecl,
	ID.NodeAssignment:   NewAssignment,
	ID.NodeReturnStmt:   NewReturnStmt,

	ID.NodeExpression: NewExpression,
	ID.NodeSelector:   NewSelector,
	ID.NodeCall:       NewCall,

	ID.NodeOr:                NewOr,
	ID.NodeAnd:               NewAnd,
	ID.NodeEquals:            NewEquals,
	ID.NodeNotEquals:         NewNotEquals,
	ID.NodeGreaterThan:       NewGreaterThan,
	ID.NodeLessThan:          NewLessThan,
	ID.NodeGreaterThanEquals: NewGreaterThanEquals,
	ID.NodeLessThanEquals:    NewLessThanEquals,
	ID.NodeBinaryPlus:        NewBinaryPlus,
	ID.NodeBinaryMinus:       NewBinaryMinus,
	ID.NodeMultiply:          NewMultiply,
	ID.NodeDivide:            NewDivide,

	ID.NodeUnaryPlus:  NewUnaryPlus,
	ID.NodeUnaryMinus: NewUnaryMinus,
	ID.NodeNot:        NewNot,

	ID.NodeIntLiteral:    NewIntLiteral,
	ID.NodeFloatLiteral:  NewFloatLiteral,
	ID.NodeStringLiteral: NewStringLiteral,
	ID.NodeBoolLiteral:   NewBoolLiteral,
	ID.NodeIdentifier:    NewIdentifier,

	ID.NodeIdentifierList: NewIdentifierList,
	ID.NodeExpressionList: NewExpressionList,
}

var NodeString = [...]func(AST, ID.Node) string{
	ID.NodeSource: SourceRoot_String,
	ID.NodeBlock:  Block_String,

	ID.NodeFunctionDecl: FunctionDecl_String,
	ID.NodeSignature:    Signature_String,
	ID.NodeConstDecl:    ConstDecl_String,
	ID.NodeAssignment:   Assignment_String,
	ID.NodeReturnStmt:   ReturnStmt_String,

	ID.NodeExpression: Expression_String,
	ID.NodeSelector:   Selector_String,
	ID.NodeCall:       Call_String,

	ID.NodeOr:                Or_String,
	ID.NodeAnd:               And_String,
	ID.NodeEquals:            Equals_String,
	ID.NodeNotEquals:         NotEquals_String,
	ID.NodeGreaterThan:       GreaterThan_String,
	ID.NodeLessThan:          LessThan_String,
	ID.NodeGreaterThanEquals: GreaterThanEquals_String,
	ID.NodeLessThanEquals:    LessThanEquals_String,
	ID.NodeBinaryPlus:        BinaryPlus_String,
	ID.NodeBinaryMinus:       BinaryMinus_String,
	ID.NodeMultiply:          Multiply_String,
	ID.NodeDivide:            Divide_String,

	ID.NodeUnaryPlus:  UnaryPlus_String,
	ID.NodeUnaryMinus: UnaryMinus_String,
	ID.NodeNot:        Not_String,

	ID.NodeIntLiteral:    IntLiteral_String,
	ID.NodeFloatLiteral:  FloatLiteral_String,
	ID.NodeStringLiteral: StringLiteral_String,
	ID.NodeBoolLiteral:   BoolLiteral_String,
	ID.NodeIdentifier:    Identifier_String,

	ID.NodeIdentifierList: IdentifierList_String,
	ID.NodeExpressionList: ExpressionList_String,
}

var NodeChildren = [...]func(AST, ID.Node) []ID.Node{
	ID.NodeSource: SourceRoot_Children,
	ID.NodeBlock:  Block_Children,

	ID.NodeFunctionDecl: FunctionDecl_Children,
	ID.NodeSignature:    Signature_Children,
	ID.NodeConstDecl:    ConstDecl_Children,
	ID.NodeAssignment:   Assignment_Children,
	ID.NodeReturnStmt:   ReturnStmt_Children,

	ID.NodeExpression: Expression_Children,
	ID.NodeSelector:   Selector_Children,
	ID.NodeCall:       Call_Children,

	ID.NodeOr:                Or_Children,
	ID.NodeAnd:               And_Children,
	ID.NodeEquals:            Equals_Children,
	ID.NodeNotEquals:         NotEquals_Children,
	ID.NodeGreaterThan:       GreaterThan_Children,
	ID.NodeLessThan:          LessThan_Children,
	ID.NodeGreaterThanEquals: GreaterThanEquals_Children,
	ID.NodeLessThanEquals:    LessThanEquals_Children,
	ID.NodeBinaryPlus:        BinaryPlus_Children,
	ID.NodeBinaryMinus:       BinaryMinus_Children,
	ID.NodeMultiply:          Multiply_Children,
	ID.NodeDivide:            Divide_Children,

	ID.NodeUnaryPlus:  UnaryPlus_Children,
	ID.NodeUnaryMinus: UnaryMinus_Children,
	ID.NodeNot:        Not_Children,

	ID.NodeIntLiteral:    IntLiteral_Children,
	ID.NodeFloatLiteral:  FloatLiteral_Children,
	ID.NodeStringLiteral: StringLiteral_Children,
	ID.NodeBoolLiteral:   BoolLiteral_Children,
	ID.NodeIdentifier:    Identifier_Children,

	ID.NodeIdentifierList: IdentifierList_Children,
	ID.NodeExpressionList: ExpressionList_Children,
}

func init() {
	for i := 0; i < ID.NodeMax; i++ {
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
	tokenIdx ID.Token
	lhs, rhs ID.Node
}

func (n Node) Tag() NodeTag    { return n.tag }
func (n Node) Token() ID.Token { return n.tokenIdx }

// General pattern of typed nodes:
// Full and interpreted struct data (typed node)
// Untyped node constructor
// Function to convert from node to typed node
// Function to convert node to string
// Function to get node children

// TODO: Add to every typed node types of it's children
type SourceRoot struct {
	Declarations []ID.Node
}

func (ast AST) SourceRoot(n Node) SourceRoot {
	decls := make([]ID.Node, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ID.Node(ast.extra[i])
		decls = append(decls, c_i)
	}
	return SourceRoot{
		Declarations: decls,
	}
}

func NewSourceRoot(rootToken ID.Token, start ID.Node, end ID.Node) Node {
	return Node{
		tag:      ID.NodeSource,
		tokenIdx: rootToken,
		lhs:      start,
		rhs:      end,
	}
}

func SourceRoot_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.SourceRoot(ast.nodes[i])
	return n.Declarations
}

func SourceRoot_String(ast AST, i ID.Node) string {
	return "Source"
}

type FunctionDecl struct {
	Name      ID.Node
	Signature ID.Node
	Body      ID.Node
}

func (ast AST) FunctionDecl(n Node) FunctionDecl {
	node := FunctionDecl{}
	node.Name = n.lhs
	extra := n.rhs
	node.Signature = ID.Node(ast.extra[extra])
	node.Body = ID.Node(ast.extra[extra+1])

	return node
}

func NewFunctionDecl(tokenIdx ID.Token, signature ID.Node, body ID.Node) Node {
	return Node{
		tag:      ID.NodeFunctionDecl,
		tokenIdx: tokenIdx,
		lhs:      signature,
		rhs:      body,
	}
}

func FunctionDecl_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.FunctionDecl(ast.nodes[i])
	return []ID.Node{n.Name, n.Signature, n.Body}
}

func FunctionDecl_String(ast AST, i ID.Node) string {
	return "FunctionDecl"
}

type Signature struct {
	Parameters ID.Node
}

func (ast AST) Signature(n Node) Signature {
	return Signature{
		Parameters: n.lhs,
	}
}

func NewSignature(tokenIdx ID.Token, parameters ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeSignature,
		tokenIdx: tokenIdx,
		lhs:      parameters,
		rhs:      rhs,
	}
}

func Signature_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Signature(ast.nodes[i])
	return []ID.Node{n.Parameters}
}

func Signature_String(ast AST, i ID.Node) string {
	return "Signature"
}

type Block struct {
	Statements []ID.Node
}

func (ast AST) Block(n Node) Block {
	statements := make([]ID.Node, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ID.Node(ast.extra[i])
		statements = append(statements, c_i)
	}
	return Block{
		Statements: statements,
	}
}

func NewBlock(tokenIdx ID.Token, start ID.Node, end ID.Node) Node {
	return Node{
		tag:      ID.NodeBlock,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}
}

func Block_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Block(ast.nodes[i])
	return n.Statements
}

func Block_String(ast AST, i ID.Node) string {
	return "Block"
}

type ConstDecl struct {
	IdentifierList ID.Node
	ExpressionList ID.Node
}

func (ast AST) ConstDecl(n Node) ConstDecl {
	return ConstDecl{
		IdentifierList: n.lhs,
		ExpressionList: n.rhs,
	}
}

func NewConstDecl(tokenIdx ID.Token, identifierList ID.Node, expressionList ID.Node) Node {
	return Node{
		tag:      ID.NodeConstDecl,
		tokenIdx: tokenIdx,
		lhs:      identifierList,
		rhs:      expressionList,
	}
}

func ConstDecl_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.ConstDecl(ast.nodes[i])
	return []ID.Node{n.IdentifierList, n.ExpressionList}
}

func ConstDecl_String(ast AST, i ID.Node) string {
	return "ConstDecl"
}

type Assignment struct {
	LhsList ID.Node
	RhsList ID.Node
}

func (ast AST) Assignment(n Node) Assignment {
	return Assignment{
		LhsList: n.lhs,
		RhsList: n.rhs,
	}
}

func NewAssignment(tokenIdx ID.Token, exprList1 ID.Node, exprList2 ID.Node) Node {
	return Node{
		tag:      ID.NodeAssignment,
		tokenIdx: tokenIdx,
		lhs:      exprList1,
		rhs:      exprList2,
	}
}

func Assignment_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Assignment(ast.nodes[i])
	return []ID.Node{n.LhsList, n.RhsList}
}

func Assignment_String(ast AST, i ID.Node) string {
	return "Assign"
}

type ReturnStmt struct {
	ExpressionList ID.Node
}

func (ast AST) ReturnStmt(n Node) ReturnStmt {
	return ReturnStmt{
		ExpressionList: n.lhs,
	}
}

func NewReturnStmt(tokenIdx ID.Token, expressions ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeReturnStmt,
		tokenIdx: tokenIdx,
		lhs:      expressions,
		rhs:      rhs,
	}
}

func ReturnStmt_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.ReturnStmt(ast.nodes[i])
	return []ID.Node{n.ExpressionList}
}

func ReturnStmt_String(ast AST, i ID.Node) string {
	return "Return"
}

type Expression struct {
	Expression ID.Node
}

func (ast AST) Expression(n Node) Expression {
	return Expression{
		Expression: n.lhs,
	}
}

func NewExpression(tokenIdx ID.Token, expr ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeExpression,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      rhs,
	}
}

func Expression_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Expression(ast.nodes[i])
	return []ID.Node{n.Expression}
}

func Expression_String(ast AST, i ID.Node) string {
	return "Expr"
}

type Selector struct {
	LhsExpr    ID.Node
	Identifier ID.Node
}

func (ast AST) Selector(n Node) Selector {
	return Selector{
		LhsExpr:    n.lhs,
		Identifier: n.rhs,
	}
}

func NewSelector(tokenIdx ID.Token, expr ID.Node, identifier ID.Node) Node {
	return Node{
		tag:      ID.NodeSelector,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      identifier,
	}
}

func Selector_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Selector(ast.nodes[i])
	return []ID.Node{n.LhsExpr, n.Identifier}
}

func Selector_String(ast AST, i ID.Node) string {
	return "Get"
}

type Call struct {
	LhsExpr   ID.Node
	Arguments ID.Node
}

func (ast AST) Call(n Node) Call {
	return Call{
		LhsExpr:   n.lhs,
		Arguments: n.rhs,
	}
}

func NewCall(tokenIdx ID.Token, expr ID.Node, args ID.Node) Node {
	return Node{
		tag:      ID.NodeCall,
		tokenIdx: tokenIdx,
		lhs:      expr,
		rhs:      args,
	}

}

func Call_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Call(ast.nodes[i])
	return []ID.Node{n.LhsExpr, n.Arguments}
}

func Call_String(ast AST, i ID.Node) string {
	return "Call"
}

type Or struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) Or(n Node) Or {
	return Or{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewOr(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeOr,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Or_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Or(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func Or_String(ast AST, i ID.Node) string {
	return "||"
}

type And struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) And(n Node) And {
	return And{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewAnd(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeAnd,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func And_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.And(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func And_String(ast AST, i ID.Node) string {
	return "&&"
}

type Equals struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) Equals(n Node) Equals {
	return Equals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewEquals(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Equals_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Equals(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func Equals_String(ast AST, i ID.Node) string {
	return "=="
}

type NotEquals struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) NotEquals(n Node) NotEquals {
	return NotEquals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewNotEquals(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeNotEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func NotEquals_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.NotEquals(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func NotEquals_String(ast AST, i ID.Node) string {
	return "!="
}

type GreaterThan struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) GreaterThan(n Node) GreaterThan {
	return GreaterThan{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewGreaterThan(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeGreaterThan,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func GreaterThan_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.GreaterThan(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func GreaterThan_String(ast AST, i ID.Node) string {
	return ">"
}

type LessThan struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) LessThan(n Node) LessThan {
	return LessThan{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewLessThan(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeLessThan,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func LessThan_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.LessThan(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func LessThan_String(ast AST, i ID.Node) string {
	return "<"
}

type GreaterThanEquals struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) GreaterThanEquals(n Node) GreaterThanEquals {
	return GreaterThanEquals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewGreaterThanEquals(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeGreaterThanEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func GreaterThanEquals_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.GreaterThanEquals(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func GreaterThanEquals_String(ast AST, i ID.Node) string {
	return ">="
}

type LessThanEquals struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) LessThanEquals(n Node) LessThanEquals {
	return LessThanEquals{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewLessThanEquals(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeLessThanEquals,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func LessThanEquals_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.LessThanEquals(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func LessThanEquals_String(ast AST, i ID.Node) string {
	return "<="
}

type BinaryPlus struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) BinaryPlus(n Node) BinaryPlus {
	return BinaryPlus{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewBinaryPlus(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeBinaryPlus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BinaryPlus_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.BinaryPlus(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func BinaryPlus_String(ast AST, i ID.Node) string {
	return "+"
}

type BinaryMinus struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) BinaryMinus(n Node) BinaryMinus {
	return BinaryMinus{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewBinaryMinus(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeBinaryMinus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BinaryMinus_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.BinaryMinus(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func BinaryMinus_String(ast AST, i ID.Node) string {
	return "-"
}

type Multiply struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) Multiply(n Node) Multiply {
	return Multiply{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewMultiply(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeMultiply,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Multiply_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Multiply(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func Multiply_String(ast AST, i ID.Node) string {
	return "*"
}

type Divide struct {
	Lhs ID.Node
	Rhs ID.Node
}

func (ast AST) Divide(n Node) Divide {
	return Divide{
		Lhs: n.lhs,
		Rhs: n.rhs,
	}
}

func NewDivide(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeDivide,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Divide_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Divide(ast.nodes[i])
	return []ID.Node{n.Lhs, n.Rhs}
}

func Divide_String(ast AST, i ID.Node) string {
	return "/"
}

type UnaryPlus struct {
	Unary ID.Node
}

func (ast AST) UnaryPlus(n Node) UnaryPlus {
	return UnaryPlus{
		Unary: n.lhs,
	}
}

func NewUnaryPlus(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeUnaryPlus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func UnaryPlus_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.UnaryPlus(ast.nodes[i])
	return []ID.Node{n.Unary}
}

func UnaryPlus_String(ast AST, i ID.Node) string {
	return "+"
}

type UnaryMinus struct {
	Unary ID.Node
}

func (ast AST) UnaryMinus(n Node) UnaryMinus {
	return UnaryMinus{
		Unary: n.lhs,
	}
}

func NewUnaryMinus(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeUnaryMinus,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func UnaryMinus_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.UnaryMinus(ast.nodes[i])
	return []ID.Node{n.Unary}
}

func UnaryMinus_String(ast AST, i ID.Node) string {
	return "-"
}

type Not struct {
	Unary ID.Node
}

func (ast AST) Not(n Node) Not {
	return Not{
		Unary: n.lhs,
	}
}
func NewNot(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeNot,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Not_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.Not(ast.nodes[i])
	return []ID.Node{n.Unary}
}

func Not_String(ast AST, i ID.Node) string {
	return "!"
}

type IdentifierList struct {
	Identifiers []ID.Node
}

func (ast AST) IdentifierList(n Node) IdentifierList {
	ids := make([]ID.Node, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		ids = append(ids, ID.Node(c_i))
	}
	return IdentifierList{
		Identifiers: ids,
	}
}
func NewIdentifierList(tokenIdx ID.Token, start ID.Node, end ID.Node) Node {
	return Node{
		tag:      ID.NodeIdentifierList,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}

}

func IdentifierList_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.IdentifierList(ast.nodes[i])
	return n.Identifiers
}

func IdentifierList_String(ast AST, i ID.Node) string {
	return "ID[]"
}

type ExpressionList struct {
	Expressions []ID.Node
}

func (ast AST) ExpressionList(n Node) ExpressionList {
	exprs := make([]ID.Node, 0, 8)
	for i := n.lhs; i < n.rhs; i++ {
		c_i := ast.extra[i]
		exprs = append(exprs, ID.Node(c_i))
	}
	return ExpressionList{
		Expressions: exprs,
	}
}

func NewExpressionList(tokenIdx ID.Token, start ID.Node, end ID.Node) Node {
	return Node{
		tag:      ID.NodeExpressionList,
		tokenIdx: tokenIdx,
		lhs:      start,
		rhs:      end,
	}

}

func ExpressionList_Children(ast AST, i ID.Node) []ID.Node {
	n := ast.IdentifierList(ast.nodes[i])
	return n.Identifiers
}

func ExpressionList_String(ast AST, i ID.Node) string {
	return "Expr[]"
}

type IntLiteral struct {
	Token ID.Token
}

func (ast AST) IntLiteral(n Node) IntLiteral {
	return IntLiteral{
		Token: n.tokenIdx,
	}
}
func NewIntLiteral(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeIntLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func IntLiteral_Children(ast AST, i ID.Node) []ID.Node {
	return []ID.Node{}
}

func IntLiteral_String(ast AST, i ID.Node) string {
	n := ast.IntLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type FloatLiteral struct {
	Token ID.Token
}

func (ast AST) FloatLiteral(n Node) FloatLiteral {
	return FloatLiteral{
		Token: n.tokenIdx,
	}
}
func NewFloatLiteral(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeFloatLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func FloatLiteral_Children(ast AST, i ID.Node) []ID.Node {
	return []ID.Node{}
}

func FloatLiteral_String(ast AST, i ID.Node) string {
	n := ast.FloatLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type StringLiteral struct {
	Token ID.Token
}

func (ast AST) StringLiteral(n Node) StringLiteral {
	return StringLiteral{
		Token: n.tokenIdx,
	}
}

func NewStringLiteral(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeStringLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func StringLiteral_Children(ast AST, i ID.Node) []ID.Node {
	return []ID.Node{}
}

func StringLiteral_String(ast AST, i ID.Node) string {
	n := ast.StringLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type BoolLiteral struct {
	Token ID.Token
}

func (ast AST) BoolLiteral(n Node) BoolLiteral {
	return BoolLiteral{
		Token: n.tokenIdx,
	}
}

func NewBoolLiteral(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeBoolLiteral,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func BoolLiteral_Children(ast AST, i ID.Node) []ID.Node {
	return []ID.Node{}
}

func BoolLiteral_String(ast AST, i ID.Node) string {
	n := ast.BoolLiteral(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type Identifier struct {
	Token ID.Token
}

func (ast AST) Identifier(n Node) Identifier {
	return Identifier{
		Token: n.tokenIdx,
	}
}
func NewIdentifier(tokenIdx ID.Token, lhs ID.Node, rhs ID.Node) Node {
	return Node{
		tag:      ID.NodeIdentifier,
		tokenIdx: tokenIdx,
		lhs:      lhs,
		rhs:      rhs,
	}

}

func Identifier_Children(ast AST, i ID.Node) []ID.Node {
	return []ID.Node{}
}

func Identifier_String(ast AST, i ID.Node) string {
	n := ast.Identifier(ast.nodes[i])
	return ast.src.Lexeme(n.Token)
}

type AST struct {
	src   *s.Source
	nodes []Node

	// NOTE: That whole thing about "extra" like in zig compiler
	// is different here - my "extra" stores arbitrary number of indicies for any
	// particular node, in contrast, zig's "extra" stores compile time known
	// fixed number of indicies (for any node), and that number depends on the node type
	// latter is more plausable, bc it's more versatile and basically superset
	// of my implementation. Well, will stick to current implemenation...
	extra []int
}

func NewAST(src *s.Source) AST {
	return AST{src: src}
}

func (ast *AST) AddNode(n Node) ID.Node {
	ast.nodes = append(ast.nodes, n)
	return ID.Node(len(ast.nodes) - 1)
}

func (ast *AST) AddExtra(extra []int) (start ID.Node, end ID.Node) {
	ast.extra = append(ast.extra, extra...)
	start = ID.Node(len(ast.extra) - len(extra))
	end = ID.Node(len(ast.extra))
	return
}

func (ast AST) SetNode(i ID.Node, n Node) {
	ast.nodes[i] = n
}

func (ast AST) GetNode(i ID.Node) Node {
	return ast.nodes[i]
}

func (ast AST) GetNodeString(i ID.Node) string {
	n := ast.nodes[i]
	getString := NodeString[n.tag]
	return getString(ast, i)
}

func (ast *AST) TraversePreorder(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNodePreorder(onEnter, onExit, 0)
}

func (ast *AST) traverseNodePreorder(onEnter NodeAction, onExit NodeAction, i ID.Node) {
	if i == ID.NodeUndefined {
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
		ast.traverseNodePreorder(onEnter, onExit, c)
	}
}

func (ast *AST) TraversePostorder(onEnter NodeAction, onExit NodeAction) {
	ast.traverseNodePostorder(onEnter, onExit, 0)
}

func (ast *AST) traverseNodePostorder(onEnter NodeAction, onExit NodeAction, i ID.Node) {
	if i == ID.NodeUndefined {
		return
	}
	n := ast.nodes[i]

	getChildren := NodeChildren[n.tag]
	children := getChildren(*ast, i)
	for _, c := range children {
		ast.traverseNodePostorder(onEnter, onExit, c)
	}

	onEnter(ast, i)
	onExit(ast, i)
}

// NOTE: For dumping and testing purposes it is plausable to have two
// different representations of an AST - full one and dump one
// I didn't account for that, so will stick with the flags to augment output
const (
	DumpPlain = 1 << iota
	DumpShowNodeID
)

func (ast *AST) Dump(flags int) string {
	str := strings.Builder{}
	onEnter := func(ast *AST, id ID.Node) (stopTraversal bool) {
		str.WriteByte('(')
		str.WriteString(ast.GetNodeString(id))

		if (flags & DumpShowNodeID) > 0 {
			str.WriteString(fmt.Sprintf(":%d", id))
		}
		return
	}
	onExit := func(ast *AST, i ID.Node) (stopTraversal bool) {
		str.WriteByte(')')
		return false
	}
	ast.TraversePreorder(onEnter, onExit)

	return str.String()
}

type TypedAST struct {
	AST
	repo T.TypeRepo
}

func NewTypedAST(ast *AST, repo T.TypeRepo) TypedAST {
	tAst := TypedAST{
		AST:  *ast,
		repo: repo,
	}
	return tAst
}

func (ast TypedAST) GetNodeType(i ID.Node) ID.Type {
	return ast.repo.NodeType(i)
}

// NOTE: This operation is overloaded in a sense that it adds no additional behaviour, just
// more information. Therefore, it is reasonable to make this not plain procedure,
// but extention point (handler) for original plain AST
// that way we can build up augmented AST more and more without much copypaste
// how to do this succinctly is another story
func (ast *TypedAST) Dump() string {
	str := strings.Builder{}
	onEnter := func(_ *AST, id ID.Node) (stopTraversal bool) {
		str.WriteByte('(')

		str.WriteString(ast.GetNodeString(id))
		str.WriteString(fmt.Sprintf(":%d", id))
		if found := ast.repo.NodeType(id); found != ID.TypeInvalid {
			str.WriteByte(' ')
			str.WriteByte('`')
			str.WriteString(ast.repo.GetString(found))
			str.WriteByte('`')
		}

		return
	}
	onExit := func(_ *AST, i ID.Node) (stopTraversal bool) {
		str.WriteByte(')')
		return false
	}
	ast.TraversePreorder(onEnter, onExit)

	return str.String()
}
