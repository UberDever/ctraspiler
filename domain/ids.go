package domain

import (
	"math"
	antlr_parser "some/antlr"
)

type Token int

const (
	TokenInvalid Token = math.MinInt
)

// NOTE: This was a bad idea - full mapping of tokens is better solution
// i.e. increase granularity
const (
	TokenEOF          Token = -1
	TokenKeyword            = antlr_parser.SomeKEYWORD
	TokenIdentifier         = antlr_parser.SomeIDENTIFIER
	TokenPunctuation        = antlr_parser.SomeOTHER_OP
	TokenUnaryOp            = antlr_parser.SomeUNARY_OP
	TokenBinaryOp           = antlr_parser.SomeBINARY_OP
	TokenIntLit             = antlr_parser.SomeINT_LIT
	TokenFloatLit           = antlr_parser.SomeFLOAT_LIT
	TokenImaginaryLit       = antlr_parser.SomeIMAGINARY_LIT
	TokenRuneLit            = antlr_parser.SomeRUNE_LIT
	TokenLittleUValue       = antlr_parser.SomeLITTLE_U_VALUE
	TokenBigUValue          = antlr_parser.SomeBIG_U_VALUE
	TokenStringLit          = antlr_parser.SomeSTRING_LIT
	TokenBoolLit            = antlr_parser.SomeBOOL_LIT
	TokenWS                 = antlr_parser.SomeWS
	TokenTerminator         = antlr_parser.SomeTERMINATOR
	TokenLineComment        = antlr_parser.SomeLINE_COMMENT
)

type Node int

const (
	NodeInvalid   Node = math.MinInt
	NodeUndefined      = -1
)

const (
	NodeSource = iota
	NodeBlock

	NodeFunctionDecl
	NodeSignature
	NodeConstDecl
	NodeVarDecl
	NodeAssignment
	NodeReturnStmt
	NodeIfStmt

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
	NodeBoolLiteral
	NodeIdentifier

	NodeIdentifierList
	NodeExpressionList
	NodeMax
)

type Kind int

const (
	KindIdentity Kind = iota
	KindPtr
	KindFunction
)

type Type int

const TypeInvalid Type = math.MinInt
const TypeVar Type = math.MaxInt

const (
	TypeInt Type = -iota - 1
	TypeFloat
	TypeBool
	TypeString
)
