package main

import (
	"ctranspiler/parser"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

const (
	TokenUndefined    = 0
	TokenIdentifier   = 1
	TokenKeyword      = 2
	TokenBinaryOp     = 3
	TokenUnaryOp      = 4
	TokenIntLit       = 5
	TokenFloatLit     = 6
	TokenImaginaryLit = 7
	TokenRuneLit      = 8
	TokenLittleUValue = 9
	TokenBigUValue    = 10
	TokenStringLit    = 11
	TokenWS           = 12
	TokenTerminator   = 13
	TokenComment      = 14
	TokenLineComment  = 15
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

func tokenize(data []byte) []Token {
	is := antlr.NewInputStream(string(data))
	lexer := parser.NewSome(is)

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

// func Parse(data []byte) ([]Token, AST) {

// 	parse := parser.NewSomeParser(stream)
// 	parse.BuildParseTrees = true
// 	parse.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

// 	visitor := parserVisitor{}
// 	visitor.VisitSource(parse.Source().(*parser.SourceContext))

// 	visitor.ast.Source = string(data)
// 	return tokens, visitor.ast
// }
