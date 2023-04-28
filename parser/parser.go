package parser

type Tag = int
type Index = int

type Parser struct {
	ast *AST
	src *Source

	current   int
	line, col int
	scratch   []Index
}

func (p *Parser) next() {
	for {
		p.current++
		c := p.src.token(p.current)
		p.line = c.line
		p.col = c.col
		tag := c.tag

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
	c := p.src.token(p.current)
	return c.tag == tag
}

func (p *Parser) expectTag(tag Tag) {
	c := p.src.token(p.current)
	if !p.matchTag(tag) {
		panic("\nExpected\n" + p.src.trace(tag, "", -1, -1) +
			"Got\n" + p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col))
	}
	p.next()
}

func (p *Parser) matchToken(tag Tag, lexeme string) bool {
	current := p.src.token(p.current)
	return current.tag == tag && p.src.lexeme(current) == lexeme
}

func (p *Parser) expectToken(tag Tag, lexeme string) {
	c := p.src.token(p.current)
	if !p.matchToken(tag, lexeme) {
		panic("\nExpected\n" + p.src.trace(tag, lexeme, -1, -1) +
			"Got\n" + p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col))
	}
	p.next()
}

func (p *Parser) expectTerminator() {
	c := p.src.token(p.current)
	if c.tag != TokenTerminator {
		panic("Expected semicolon near " + p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col))
	}
}

func (p *Parser) restoreScratch(old_size int) {
	p.scratch = p.scratch[:old_size]
}

func (p *Parser) addScratchToExtra(scratch_top int) (start int, end int) {
	slice := p.scratch[scratch_top:]
	p.ast.extra = append(p.ast.extra, slice...)
	start = len(p.ast.extra) - len(slice)
	end = len(p.ast.extra)
	return
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
		scratch: make([]Index, 0, 64),
	}

	p.parseSource()

	return ast
}

func (p *Parser) parseSource() {
	p.ast.nodes = append(p.ast.nodes, Node{
		tag: NodeSource,
	})
	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	for {
		p.next()

		t := p.src.token(p.current)
		if t.tag == TokenEOF {
			break
		}

		if !p.matchToken(TokenKeyword, "fn") {
			c := p.src.token(p.current)
			tokenTrace := p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col)
			panic("At " + tokenTrace + " expected function declaration")
		}

		index := p.parseFunctionDecl()
		p.scratch = append(p.scratch, index)
	}
	p.ast.nodes[0].lhs, p.ast.nodes[0].rhs = p.addScratchToExtra(scratch_top)
}

func (p *Parser) parseFunctionDecl() int {
	n := NullNode
	n.tag = NodeFunctionDecl

	p.expectToken(TokenKeyword, "fn")
	n.tokenIdx = p.current
	p.expectTag(TokenIdentifier)

	n.lhs = p.parseSignature()
	n.rhs = 0
	if p.matchToken(TokenPunctuation, "{") {
		n.rhs = p.parseBlock()
	}
	p.expectTerminator()

	return p.addNode(n)
}

func (p *Parser) parseSignature() int {
	n := NullNode
	n.tag, n.tokenIdx = NodeSignature, p.current
	n.rhs = 0

	n.lhs = 0
	p.expectToken(TokenPunctuation, "(")
	if p.matchToken(TokenPunctuation, ")") {
		p.next()
		return p.addNode(n)
	}

	n.lhs = p.parseIdentifierList()
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
	n.tag, n.tokenIdx = NodeIdentifierList, p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, p.current)
	p.expectTag(TokenIdentifier)
	for {
		if p.matchToken(TokenPunctuation, ",") {
			p.next()

			p.scratch = append(p.scratch, p.current)
			p.expectTag(TokenIdentifier)
		} else {
			break
		}
	}

	n.lhs, n.rhs = p.addScratchToExtra(scratch_top)
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
