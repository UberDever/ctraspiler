package syntax

import (
	"fmt"
	"some/util"
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
		expected := p.src.trace(tag, "", int(TokenEOF), int(TokenEOF))
		got := p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col)
		util.ErrorHandler.Add(util.NewError(
			util.Parser, util.EP_ExpectedTag, c.line, c.col, p.src.file, expected, got,
		))
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
		expected := p.src.trace(tag, lexeme, int(TokenEOF), int(TokenEOF))
		got := p.src.trace(c.tag, p.src.lexeme(c), c.line, c.col)
		util.ErrorHandler.Add(util.NewError(
			util.Parser, util.EP_ExpectedToken, c.line, c.col, p.src.file, expected, got,
		))
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
	tag, tokenIdx, lhs, rhs := NodeSource, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	// make root the first node
	p.ast.nodes = append(p.ast.nodes, Node{})

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
			e := util.NewError(util.Parser, 0, c.line, c.col, p.src.file, tokenTrace)
			util.ErrorHandler.Add(e)
			// TODO: Add error handling after this on the call of Parse
			// also, do we need restore here?
			return
		}

		index := p.parseFunctionDecl()
		p.scratch = append(p.scratch, AnyIndex(index))
	}
	lhs, rhs = p.addScratchToExtra(scratch_top)

	p.ast.nodes[0] = NodeConstructor[tag](tokenIdx, lhs, rhs)
}

func (p *Parser) parseFunctionDecl() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeFunctionDecl, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid

	p.expectToken(TokenKeyword, "fn")
	// this will store identifier to ast.nodes
	// we will find that node later by ast.nodes traversal
	tokenIdx = p.current
	_ = p.parseIdentifier()

	lhs = p.parseSignature()
	rhs = NodeIndexUndefined
	if p.matchToken(TokenPunctuation, "{") {
		rhs = p.parseBlock()
	}
	p.expectTerminator()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseSignature() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeSignature, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current
	rhs = NodeIndexUndefined

	p.expectToken(TokenPunctuation, "(")
	if p.matchToken(TokenPunctuation, ")") {
		p.next()
		lhs = p.addNode(NodeConstructor[NodeIdentifierList](
			p.current, NodeIndexUndefined, NodeIndexUndefined))
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	lhs = p.parseIdentifierList()
	p.expectToken(TokenPunctuation, ")")

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseBlock() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeBlock, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.expectToken(TokenPunctuation, "{")
	if p.matchToken(TokenPunctuation, "}") {
		lhs, rhs = NodeIndexUndefined, NodeIndexUndefined
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	for !p.matchToken(TokenPunctuation, "}") {
		i := p.parseStatement()
		if i != NodeIndexInvalid {
			p.scratch = append(p.scratch, AnyIndex(i))
		}
		p.expectTerminator()
	}

	p.expectToken(TokenPunctuation, "}")
	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
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
	tag, tokenIdx, lhs, rhs := NodeConstDecl, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	p.expectToken(TokenKeyword, "const")
	lhs = p.parseIdentifierList()
	p.expectToken(TokenPunctuation, "=")
	rhs = p.parseExpressionList()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseAssignment() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeAssignment, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	lhs = p.parseExpressionList()
	p.expectToken(TokenPunctuation, "=")
	rhs = p.parseExpressionList()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseExpressionList() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeExpressionList, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

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

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseExpression() NodeIndex {
	return p.parseBinaryExpr(precLowest + 1)
}

func (p *Parser) parseBinaryExpr(precedence int) NodeIndex {
	tokenIdx, lhs, rhs := TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid

	lhs = p.parseUnaryExpr()
	for {
		op := p.src.lexeme(p.src.token(p.current))
		opPrec, tag := binaryPrecedenceAndTag(op)
		if opPrec < precedence {
			return lhs
		}
		tokenIdx = p.current
		p.next()

		rhs = p.parseBinaryExpr(opPrec + 1)
		lhs = p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
}

func (p *Parser) parseUnaryExpr() NodeIndex {
	tokenIdx, lhs, rhs := TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	tag := unaryTag(p.src.lexeme(p.src.token(p.current)))
	if tag == NodeUndefined {
		return p.parsePrimaryExpr()
	}
	p.next()
	lhs = p.parseUnaryExpr()
	rhs = NodeIndexUndefined
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parsePrimaryExpr() NodeIndex {
	tokenIdx, lhs, rhs := TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	lhs = p.parseOperand()
	if p.matchToken(TokenPunctuation, ".") {
		p.next()
		tag := NodeSelector
		rhs = p.parseIdentifier()
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	} else if p.matchToken(TokenPunctuation, "(") {
		p.next()
		tag := NodeCall
		rhs = NodeIndexUndefined
		if !p.matchToken(TokenPunctuation, ")") {
			rhs = p.parseExpressionList()
			p.expectToken(TokenPunctuation, ")")
			return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
		}
		p.next()
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
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
	tag, tokenIdx, lhs, rhs := NodeIdentifierList, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

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

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseIdentifier() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeIdentifier, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	lhs, rhs = NodeIndexUndefined, NodeIndexUndefined
	p.expectTag(TokenIdentifier)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *Parser) parseLiteral() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeIntLiteral, TokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current
	lhs, rhs = NodeIndexUndefined, NodeIndexUndefined

	if p.matchTag(TokenIntLit) {
		tag = NodeIntLiteral
	} else if p.matchTag(TokenFloatLit) {
		tag = NodeFloatLiteral
	} else if p.matchTag(TokenStringLit) {
		tag = NodeStringLiteral
	} else {
		panic("parseLiteral unimplemented")
	}

	p.next()
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}
