package parser

import "golang.org/x/exp/utf8string"

type Tag = int
type Index = int

type parser struct {
	ast     *AST
	current int
	wasNL   bool
}

func (p *parser) advance() {
	p.wasNL = false
	for {
		tag := p.ast.tokens[p.current].tag
		switch tag {
		case TokenEOF:
			p.wasNL = true
			return
		case TokenTerminator:
			fallthrough
		case TokenLineComment:
			p.wasNL = true
		default:
			return
		}
		p.current++
	}
}

func Parse(source []byte, tokens []Token) AST {
	ast := AST{
		source: *utf8string.NewString(string(source)),
		tokens: tokens,
	}

	p := parser{
		ast: &ast,
	}

	p.parseSource()

	return ast
}

func (p *parser) parseSource() {
	sourceNode := Node{
		tag: NodeSource,
		lhs: len(p.ast.extra),
		rhs: 0,
	}

	for {
		t := p.ast.tokens[p.current]
		if t.tag == TokenEOF {
			break
		}

		p.advance()
		index := p.parseLiteral()

		p.ast.extra = append(p.ast.extra, index)
		sourceNode.rhs++
	}

	p.ast.nodes = append(p.ast.nodes, sourceNode)
}

func (p *parser) parseLiteral() int {
	t := p.ast.tokens[p.current]
	switch t.tag {
	case TokenIntLit:
		p.ast.nodes = append(p.ast.nodes, Node{
			tag:      NodeIntLiteral,
			tokenIdx: p.current,
		})
		p.current++
	case TokenFloatLit:
		p.ast.nodes = append(p.ast.nodes, Node{
			tag:      NodeFloatLiteral,
			tokenIdx: p.current,
		})
		p.current++
	case TokenStringLit:
		p.ast.nodes = append(p.ast.nodes, Node{
			tag:      NodeStringLiteral,
			tokenIdx: p.current,
		})
		p.current++

	default:
		return -1
	}
	index := len(p.ast.nodes) - 1
	return index
}
