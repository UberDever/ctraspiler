package main

import (
	"ctranspiler/parser"

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

type Token struct {
	tag   Tag
	start uint
	end   uint
}

// TODO: Write my own visitor that can return value (node index)
// TODO: Add types to grammar
// TODO: Add nodes:
// 1. lhs = extra.len()
// 2. rhs = extra.len() + node.len()
// 3. extra.reserve(node.len())
// 4. for i in node { extra[extra_prevlen + i] = visit(node[i]) }

func Parse(config Config, data []byte) ([]Token, AST) {
	is := antlr.NewInputStream(string(data))
	lexer := parser.NewSomeLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	antlrTokens := lexer.GetAllTokens()
	tokens := make([]Token, 0, len(antlrTokens))
	for i := range antlrTokens {
		t := antlrTokens[i]
		tokens = append(tokens, Token{
			tag:   t.GetTokenType(),
			start: uint(t.GetStart()),
			end:   uint(t.GetStop()),
		})
	}

	parse := parser.NewSomeParser(stream)
	parse.BuildParseTrees = true
	parse.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

	visitor := parserVisitor{}
	visitor.Visit(parse.Source())

	visitor.ast.Source = string(data)
	return tokens, visitor.ast
}

type parserVisitor struct {
	parser.BaseSomeVisitor
	ast AST
}

func (v *parserVisitor) Visit(tree antlr.ParseTree) interface{} {
	return tree.Accept(v)
}

func (v *parserVisitor) VisitChildren(node antlr.RuleNode) interface{} {
	for _, child := range node.GetChildren() {
		child.(antlr.ParseTree).Accept(v)
	}
	return nil
}

// func (v *parserVisitor) VisitSource(ctx *parser.parser.SourceContext) interface{} {
// 	// s := ctx.StatementList().AllStatement()
// 	// for i := range s {
// 	// 	v.VisitStatement(s[i].(*parser.StatementContext))
// 	// }
// 	fmt.Println("Visiting source")
// 	return AST{}
// }

// func (v *parserVisitor) VisitFunctionDecl(ctx *parser.parser.FunctionDeclContext) any {
// 	fmt.Println("Visiting function decl")
// 	fmt.Println(ctx.IDENTIFIER())
// 	return nil
// }

// func (v *parserVisitor) VisitStatement(ctx *parser.parser.StatementContext) any {
// 	return v.VisitSimpleStmt(ctx.SimpleStmt().(*parser.SimpleStmtContext))
// }

// func (v *parserVisitor) VisitSimpleStmt(ctx *parser.parser.SimpleStmtContext) any {
// 	fmt.Println(ctx.IsEmpty())
// 	return ctx.ExpressionStmt()
// }

func (v *parserVisitor) VisitSource(ctx *parser.SourceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitFunctionDecl(ctx *parser.FunctionDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitFunction(ctx *parser.FunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitSignature(ctx *parser.SignatureContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitParameters(ctx *parser.ParametersContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitIdentifierList(ctx *parser.IdentifierListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitBlock(ctx *parser.BlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitStatementList(ctx *parser.StatementListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitStatement(ctx *parser.StatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitSimpleStmt(ctx *parser.SimpleStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitEmptyStmt(ctx *parser.EmptyStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitExpressionStmt(ctx *parser.ExpressionStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitExpression(ctx *parser.ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitUnaryExpr(ctx *parser.UnaryExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitPrimaryExpr(ctx *parser.PrimaryExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitOperand(ctx *parser.OperandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitLiteral(ctx *parser.LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitBasicLit(ctx *parser.BasicLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitOperandName(ctx *parser.OperandNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *parserVisitor) VisitQualifiedIdent(ctx *parser.QualifiedIdentContext) interface{} {
	return v.VisitChildren(ctx)
}
