package main

import (
	"ctranspiler/parser"
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

const (
	TagSource = iota
	TagStatement
	TagDeclaration
	TagExpression
)

type Tag = int
type Data = int

type Node struct {
	tag      Tag
	token    Data
	lhs, rhs Data
}

type AST struct {
	Source string
	Nodes  []Node
	Extra  []Data
}

func Parse(config Config, data []byte) AST {
	is := antlr.NewInputStream(string(data))
	lexer := parser.NewGoLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	parse := parser.NewGoParser(stream)
	visitor := parserVisitor{}

	source := parse.Source().(*parser.SourceContext)
	ast := (visitor.VisitSource(source)).(AST)
	ast.Source = string(data)

	return ast
}

type parserVisitor struct {
	parser.BaseGoParserVisitor
}

func (v *parserVisitor) VisitSource(ctx *parser.SourceContext) any {
	s := ctx.StatementList().AllStatement()
	for i := range s {
		v.VisitStatement(s[i].(*parser.StatementContext))
	}
	return AST{}
}

func (v *parserVisitor) VisitStatement(ctx *parser.StatementContext) any {
	fmt.Println(ctx)
	return nil
}
