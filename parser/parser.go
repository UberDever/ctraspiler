package parser

import (
	antlr_parser "some/antlr"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

const (
	TokenUndefined    = 0
	TokenKeyword      = antlr_parser.SomeKEYWORD
	TokenIdentifier   = antlr_parser.SomeIDENTIFIER
	TokenBinaryOp     = antlr_parser.SomeBINARY_OP
	TokenUnaryOp      = antlr_parser.SomeUNARY_OP
	TokenIntLit       = antlr_parser.SomeINT_LIT
	TokenFloatLit     = antlr_parser.SomeFLOAT_LIT
	TokenImaginaryLit = antlr_parser.SomeIMAGINARY_LIT
	TokenRuneLit      = antlr_parser.SomeRUNE_LIT
	TokenLittleUValue = antlr_parser.SomeLITTLE_U_VALUE
	TokenBigUValue    = antlr_parser.SomeBIG_U_VALUE
	TokenStringLit    = antlr_parser.SomeSTRING_LIT
	TokenWS           = antlr_parser.SomeWS
	TokenTerminator   = antlr_parser.SomeTERMINATOR
	TokenComment      = antlr_parser.SomeCOMMENT
	TokenLineComment  = antlr_parser.SomeLINE_COMMENT
)

const (
	NodeSource = iota
	NodeStatement
	NodeDeclaration
	NodeExpression
)

type Tag = int
type Data = int

type Token struct {
	tag   Tag
	start uint
	end   uint
	line  int
	col   int
}

type Node struct {
	tag      Tag
	tokenIdx Data
	lhs, rhs Data
}

type AST struct {
	Source string
	Nodes  []Node
	Extra  []Data
}

func tokenize(source []byte) []Token {
	is := antlr.NewInputStream(string(source))
	lexer := antlr_parser.NewSome(is)

	antlrTokens := lexer.GetAllTokens()
	tokens := make([]Token, 0, len(antlrTokens))
	for i := range antlrTokens {
		t := antlrTokens[i]
		if t.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}
		tokens = append(tokens, Token{
			tag:   t.GetTokenType(),
			start: uint(t.GetStart()),
			end:   uint(t.GetStop()),
			line:  t.GetLine(),
			col:   t.GetColumn(),
		})
	}

	return tokens
}

// TODO: Write my own visitor that can return value (node index)
// TODO: Add types to grammar
// TODO: Add nodes:
// 1. lhs = extra.len()
// 2. rhs = extra.len() + node.len()
// 3. extra.reserve(node.len())
// 4. for i in node { extra[extra_prevlen + i] = visit(node[i]) }

func Parse(source []byte, tokens []Token) AST {
	ast := AST{}
	ast.Source = string(source)

	return ast
}
