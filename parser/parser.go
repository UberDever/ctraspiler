package parser

import (
	"fmt"
)

type Tag = int
type Index = int

type Parser struct {
	ast *AST
	src *Source

	current   int
	line, col int
}

func (p *Parser) next() {
	for {
		p.current++
		current := p.src.token(p.current)
		p.line = current.line
		p.col = current.col
		tag := current.tag

		if tag == TokenLineComment {
			continue
		}
		return
	}
}

func (p *Parser) addNode(n Node) int {
	p.ast.nodes = append(p.ast.nodes, n)
	return len(p.ast.nodes) - 1
}

func (p *Parser) matchTag(tag Tag) bool {
	current := p.src.token(p.current)
	return current.tag == tag
}

func (p *Parser) expectTag(tag Tag) {
	current := p.src.token(p.current)
	if !p.matchTag(tag) {
		expected := fmt.Sprintf("Expected %d", tag)
		panic(expected + ", but got " + p.src.trace(current))
	}
	p.next()
}

func (p *Parser) matchToken(tag Tag, lexeme string) bool {
	current := p.src.token(p.current)
	return current.tag == tag && p.src.lexeme(current) == lexeme
}

func (p *Parser) expectToken(tag Tag, lexeme string) {
	current := p.src.token(p.current)
	if !p.matchToken(tag, lexeme) {
		expected := fmt.Sprintf("Expected %d/%s", tag, lexeme)
		panic(expected + ", but got " + p.src.trace(current))
	}
	p.next()
}

func (p *Parser) expectTerminator() {
	current := p.src.token(p.current)
	if current.tag != TokenTerminator {
		panic("Expected semicolon near " + p.src.trace(current))
	}
}

func Parse(src *Source) AST {
	ast := AST{
		src: src,
	}
	p := Parser{
		ast:     &ast,
		src:     src,
		current: -1,
		line:    0,
		col:     0,
	}

	p.parseSource()

	return ast
}

func (p *Parser) parseSource() {
	p.ast.nodes = append(p.ast.nodes, Node{
		tag: NodeSource,
	})

	for {
		p.next()

		t := p.src.token(p.current)
		if t.tag == TokenEOF {
			break
		}

		if !p.matchToken(TokenKeyword, "fn") {
			tokenTrace := p.src.trace(p.src.token(p.current))
			panic("At " + tokenTrace + " expected function declaration")
		}

		index := p.parseFunctionDecl()
		p.ast.extra = append(p.ast.extra, index)
		p.ast.nodes[0].rhs++
	}

}

func (p *Parser) parseFunctionDecl() int {
	n := NullNode
	n.tag, n.tokenIdx = NodeFunctionDecl, p.current

	p.expectToken(TokenKeyword, "fn")
	p.expectTag(TokenIdentifier)
	n.lhs = p.parseSignature()
	n.rhs = p.parseBlock()
	p.expectTerminator()

	return p.addNode(n)
}

func (p *Parser) parseSignature() int {
	n := NullNode
	n.tag, n.tokenIdx = NodeParameters, p.current

	p.expectToken(TokenPunctuation, "(")
	if p.matchToken(TokenPunctuation, ")") {
		p.next()
		n.lhs, n.rhs = 0, 0
		return p.addNode(n)
	}

	p.parseIdentifierList()
	p.expectToken(TokenPunctuation, ")")

	return p.addNode(n)
}

func (p *Parser) parseBlock() int {
	n := NullNode
	n.tag, n.tokenIdx = NodeBlock, p.current

	return p.addNode(n)
}

func (p *Parser) parseIdentifierList() int {
	n := NullNode

	return p.addNode(n)
}

// func (p *Parser) parseLiteral() int {
// 	t := p.ast.tokens[p.token_i]
// 	switch t.tag {
// 	case TokenIntLit:
// 		p.ast.nodes = append(p.ast.nodes, Node{
// 			tag:      NodeIntLiteral,
// 			tokenIdx: p.token_i,
// 		})
// 		p.token_i++
// 	case TokenFloatLit:
// 		p.ast.nodes = append(p.ast.nodes, Node{
// 			tag:      NodeFloatLiteral,
// 			tokenIdx: p.token_i,
// 		})
// 		p.token_i++
// 	case TokenStringLit:
// 		p.ast.nodes = append(p.ast.nodes, Node{
// 			tag:      NodeStringLiteral,
// 			tokenIdx: p.token_i,
// 		})
// 		p.token_i++

// 	default:
// 		return -1
// 	}
// 	index := len(p.ast.nodes) - 1
// 	return index
// }
