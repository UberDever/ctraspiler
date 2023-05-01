package parser

import (
	"fmt"
)

const (
	precLowest  = iota
	precHighest = 7
)

func binaryPrecedenceAndTag(lexeme string) (int, NodeTag) {
	switch lexeme {
	case "||":
		return 1, NodeOr
	case "&&":
		return 2, NodeAnd
	case "==":
		return 3, NodeEquals
	case "!=":
		return 3, NodeNotEquals
	case ">":
		return 3, NodeGreaterThan
	case "<":
		return 3, NodeLessThan
	case ">=":
		return 3, NodeGreaterThanEquals
	case "<=":
		return 3, NodeLessThanEquals
	case "+":
		return 4, NodeBinaryPlus
	case "-":
		return 4, NodeBinaryMinus
	case "*":
		return 5, NodeMultiply
	case "/":
		return 5, NodeDivide
	}
	return precLowest, NodeUndefined
}

func unaryTag(lexeme string) NodeTag {
	switch lexeme {
	case "+":
		return NodeUnaryPlus
	case "-":
		return NodeUnaryMinus
	case "!":
		return NodeNot
	}
	return NodeUndefined
}

// NOTE: this is like
// type AnyIndex = TokenIndex | NodeIndex
// but since golang doesn't support sum types, i forced to do conversions
type AnyIndex int

type Parser struct {
	ast *AST
	src *Source

	current   TokenIndex
	line, col int
	scratch   []AnyIndex

	saved TokenIndex
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

func (p *Parser) save() {
	p.saved = p.current
}

func (p *Parser) rollback() {
	p.current = p.saved
	p.saved = -1
	c := p.src.token(p.current)
	p.line = c.line
	p.col = c.col
}

func (p *Parser) addNode(n Node) NodeIndex {
	p.ast.nodes = append(p.ast.nodes, n)
	return NodeIndex(len(p.ast.nodes) - 1)
}

func (p *Parser) matchTag(tag TokenTag) bool {
	c := p.src.token(p.current)
	return c.tag == tag
}

func (p *Parser) expectTag(tag TokenTag) {
	c := p.src.token(p.current)
	if !p.matchTag(tag) {
		panic("\nExpected\n" + p.src.trace(tag, "", int(TokenEOF), int(TokenEOF)) +
			"Got\n" + p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col) +
			fmt.Sprintf("Near\n%#v", p.src.near(p.current)))
	}
	p.next()
}

func (p *Parser) matchToken(tag TokenTag, lexeme string) bool {
	current := p.src.token(p.current)
	return current.tag == tag && p.src.lexeme(current) == lexeme
}

func (p *Parser) expectToken(tag TokenTag, lexeme string) {
	c := p.src.token(p.current)
	if !p.matchToken(tag, lexeme) {
		panic("\nExpected\n" + p.src.trace(tag, lexeme, int(TokenEOF), int(TokenEOF)) +
			"Got\n" + p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col) +
			fmt.Sprintf("Line: %#v", p.src.near(p.current)))
	}
	p.next()
}

func (p *Parser) expectTerminator() {
	c := p.src.token(p.current)
	if c.tag != TokenTerminator {
		panic("Expected semicolon\n" + p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col) +
			fmt.Sprintf("Near\n%#v", p.src.near(p.current)))
	}
	p.next()
}

func (p *Parser) restoreScratch(old_size int) {
	p.scratch = p.scratch[:old_size]
}

func (p *Parser) addScratchToExtra(scratch_top int) (start NodeIndex, end NodeIndex) {
	slice := p.scratch[scratch_top:]
	p.ast.extra = append(p.ast.extra, slice...)
	start = NodeIndex(len(p.ast.extra) - len(slice))
	end = NodeIndex(len(p.ast.extra))
	return
}

// this thing should be the norm - need to increase token granularity
func (p *Parser) isLiteral() bool {
	return p.matchTag(TokenIntLit) ||
		p.matchTag(TokenFloatLit) ||
		p.matchTag(TokenStringLit)
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
		scratch: make([]AnyIndex, 0, 64),
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

	p.next()

	for {
		t := p.src.token(p.current)
		if t.tag == TokenEOF {
			break
		}

		if !p.matchToken(TokenKeyword, "fn") {
			c := p.src.token(p.current)
			tokenTrace := p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col)
			panic("At\n" + tokenTrace + "expected function declaration")
		}

		index := p.parseFunctionDecl()
		p.scratch = append(p.scratch, AnyIndex(index))
	}
	p.ast.nodes[0].lhs, p.ast.nodes[0].rhs = p.addScratchToExtra(scratch_top)
}

func (p *Parser) parseFunctionDecl() NodeIndex {
	n := InvalidNode
	n.tag = NodeFunctionDecl

	p.expectToken(TokenKeyword, "fn")

	// this will store identifier to ast.nodes
	// we will find that node later by ast.nodes traversal
	n.tokenIdx = p.current
	_ = p.parseIdentifier()

	n.lhs = p.parseSignature()
	n.rhs = NodeIndexUndefined
	if p.matchToken(TokenPunctuation, "{") {
		n.rhs = p.parseBlock()
	}
	p.expectTerminator()

	return p.addNode(n)
}

func (p *Parser) parseSignature() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeSignature, p.current
	n.lhs = NodeIndexUndefined
	n.rhs = NodeIndexUndefined

	p.expectToken(TokenPunctuation, "(")
	if p.matchToken(TokenPunctuation, ")") {
		p.next()
		n.lhs = p.addNode(Node{tag: NodeIdentifierList})
		return p.addNode(n)
	}

	n.lhs = p.parseIdentifierList()
	p.expectToken(TokenPunctuation, ")")

	return p.addNode(n)
}

func (p *Parser) parseBlock() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeBlock, p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.expectToken(TokenPunctuation, "{")
	if p.matchToken(TokenPunctuation, "}") {
		n.lhs, n.rhs = NodeIndexUndefined, NodeIndexUndefined
		return p.addNode(n)
	}

	for !p.matchToken(TokenPunctuation, "}") {
		i := p.parseStatement()
		if i != NodeIndexInvalid {
			p.scratch = append(p.scratch, AnyIndex(i))
		}
		p.expectTerminator()
	}

	p.expectToken(TokenPunctuation, "}")
	n.lhs, n.rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(n)
}

func (p *Parser) parseStatement() NodeIndex {
	if p.matchTag(TokenTerminator) {
		// skip empty statement
		return NodeIndexInvalid
	} else if p.matchToken(TokenKeyword, "const") {
		return p.parseConstDecl()
	} else {
		// NOTE: need to rollback here, because I don't bother
		// to find all terminals that start an expression
		// if grammar lets you do that this is very convenient and the right thing

		p.save()
		i := p.parseExpression()
		if p.matchTag(TokenTerminator) {
			// expression statement
			return i
		}
		s := p.src.lexeme(p.src.token(p.current))
		_ = s
		p.rollback()
		return p.parseAssignment()
	}
}

func (p *Parser) parseConstDecl() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeConstDecl, p.current

	p.expectToken(TokenKeyword, "const")
	n.lhs = p.parseIdentifierList()
	p.expectToken(TokenPunctuation, "=")
	n.rhs = p.parseExpressionList()

	return p.addNode(n)
}

func (p *Parser) parseAssignment() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeAssignment, p.current

	n.lhs = p.parseExpressionList()
	p.expectToken(TokenPunctuation, "=")
	n.rhs = p.parseExpressionList()

	return p.addNode(n)
}

func (p *Parser) parseExpressionList() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeExpressionList, p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, AnyIndex(p.parseExpression()))
	for {
		if p.matchToken(TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, AnyIndex(p.parseExpression()))
		} else {
			break
		}
	}

	n.lhs, n.rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(n)
}

func (p *Parser) parseExpression() NodeIndex {
	return p.parseBinaryExpr(precLowest + 1)
}

func (p *Parser) parseBinaryExpr(precedence int) NodeIndex {
	n := InvalidNode

	lhs := p.parseUnaryExpr()
	for {
		op := p.src.lexeme(p.src.token(p.current))
		opPrec, tag := binaryPrecedenceAndTag(op)
		if opPrec < precedence {
			return lhs
		}
		n.tokenIdx = p.current
		p.next()

		n.tag = tag
		n.rhs = p.parseBinaryExpr(opPrec + 1)
		n.lhs = lhs
		lhs = p.addNode(n)
	}
}

func (p *Parser) parseUnaryExpr() NodeIndex {
	n := InvalidNode
	n.tokenIdx = p.current
	n.tag = unaryTag(p.src.lexeme(p.src.token(p.current)))
	if n.tag == NodeUndefined {
		return p.parsePrimaryExpr()
	}
	p.next()
	n.lhs = p.parseUnaryExpr()
	n.rhs = NodeIndexUndefined
	return p.addNode(n)
}

func (p *Parser) parsePrimaryExpr() NodeIndex {
	n := InvalidNode
	n.tokenIdx = p.current
	lhs := p.parseOperand()
	if p.matchToken(TokenPunctuation, ".") {
		p.next()
		n.tag = NodeSelector
		n.lhs = lhs
		n.rhs = p.parseIdentifier()
		return p.addNode(n)
	} else if p.matchToken(TokenPunctuation, "(") {
		p.next()
		n.tag = NodeCall
		n.lhs = lhs
		n.rhs = NodeIndexUndefined
		if !p.matchToken(TokenPunctuation, ")") {
			n.rhs = p.parseExpressionList()
			p.expectToken(TokenPunctuation, ")")
			return p.addNode(n)
		}
		p.next()
		return p.addNode(n)
	}
	return lhs
}

func (p *Parser) parseOperand() NodeIndex {
	if p.matchTag(TokenIdentifier) {
		return p.parseIdentifier()
	}
	if p.isLiteral() {
		return p.parseLiteral()
	}
	p.expectToken(TokenPunctuation, "(")
	i := p.parseExpression()
	p.expectToken(TokenPunctuation, ")")
	return i
}

func (p *Parser) parseIdentifierList() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeIdentifierList, p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, AnyIndex(p.parseIdentifier()))
	for {
		if p.matchToken(TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, AnyIndex(p.parseIdentifier()))
		} else {
			break
		}
	}

	n.lhs, n.rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(n)
}

func (p *Parser) parseIdentifier() NodeIndex {
	n := InvalidNode
	n.tag, n.tokenIdx = NodeIdentifier, p.current
	n.lhs, n.rhs = NodeIndexUndefined, NodeIndexUndefined
	p.expectTag(TokenIdentifier)
	return p.addNode(n)
}

func (p *Parser) parseLiteral() NodeIndex {
	n := InvalidNode
	n.tokenIdx = p.current
	n.lhs, n.rhs = NodeIndexUndefined, NodeIndexUndefined

	if p.matchTag(TokenIntLit) {
		n.tag = NodeIntLiteral
	} else if p.matchTag(TokenFloatLit) {
		n.tag = NodeFloatLiteral
	} else if p.matchTag(TokenStringLit) {
		n.tag = NodeStringLiteral
	} else {
		panic("parseLiteral unimplemented")
	}

	p.next()
	return p.addNode(n)
}
