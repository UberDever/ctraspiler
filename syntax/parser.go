package syntax

import "some/util"

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
	return precLowest, nodeUndefined
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
	return nodeUndefined
}

// NOTE: this is like
// type anyIndex = tokenIndex | nodeIndex
// but since golang doesn't support sum types, I forced to do conversions
type anyIndex int

type parser struct {
	ast     *AST
	src     *Source
	handler *util.ErrorHandler

	current   tokenIndex
	line, col int
	scratch   []anyIndex

	saved tokenIndex
	atEOF bool
}

func NewParser(handler *util.ErrorHandler) parser {
	return parser{
		handler: handler,
		current: -1,
		line:    0,
		col:     0,
		scratch: make([]anyIndex, 0, 64),
		saved:   tokenIndexInvalid,
		atEOF:   false,
	}

}

func (p *parser) Parse(src *Source) AST {
	ast := NewAST(src)
	p.ast = &ast
	p.src = src
	p.parseSource()
	return ast
}

func (p *parser) next() {
	for {
		p.current++
		c := p.src.token(p.current)
		p.line = c.line
		p.col = c.col
		p.atEOF = c.tag == TokenEOF

		if c.tag == TokenLineComment {
			continue
		}
		return
	}
}

func (p *parser) save() {
	p.saved = p.current
}

func (p *parser) rollback() {
	p.current = p.saved
	p.saved = -1
	c := p.src.token(p.current)
	p.line = c.line
	p.col = c.col
}

func (p *parser) addNode(n Node) NodeIndex {
	p.ast.nodes = append(p.ast.nodes, n)
	return NodeIndex(len(p.ast.nodes) - 1)
}

func (p *parser) matchTag(tag tokenTag) bool {
	if p.atEOF {
		return false
	}

	c := p.src.token(p.current)
	return c.tag == tag
}

func (p *parser) matchToken(tag tokenTag, lexeme string) bool {
	if p.atEOF {
		return false
	}

	current := p.src.token(p.current)
	return current.tag == tag && p.src.Lexeme(p.current) == lexeme
}

// NOTE: zero value of string is "" and because this is valid in my case (I don't pass that value
// from variables only from literal strings) i can use it as Optional<string>
// but optionals really is missing...
func (p *parser) expect(tag tokenTag, lexeme string) {
	if p.atEOF {
		return
	}

	matched := false
	if lexeme == "" {
		matched = p.matchTag(tag)
	} else {
		matched = p.matchToken(tag, lexeme)
	}

	if !matched {
		c := p.src.token(p.current)
		expected := p.src.traceToken(tag, lexeme, int(TokenEOF), int(TokenEOF))
		got := p.src.traceToken(c.tag, p.src.Lexeme(p.current), c.line, c.col)
		p.handler.Add(util.NewError(
			util.Parser, util.EP_ExpectedToken, c.line, c.col, p.src.filename, expected, got,
		))

		// discard tokens
		for !p.atEOF {
			p.next()
		}
		return
	}

	p.next()
}

func (p *parser) restoreScratch(old_size int) {
	p.scratch = p.scratch[:old_size]
}

func (p *parser) addScratchToExtra(scratch_top int) (start NodeIndex, end NodeIndex) {
	slice := p.scratch[scratch_top:]
	p.ast.extra = append(p.ast.extra, slice...)
	start = NodeIndex(len(p.ast.extra) - len(slice))
	end = NodeIndex(len(p.ast.extra))
	return
}

func (p *parser) isLiteral() bool {
	return p.matchTag(TokenIntLit) ||
		p.matchTag(TokenFloatLit) ||
		p.matchTag(TokenStringLit)
}

func (p *parser) parseSource() {
	tag, tokenIdx, lhs, rhs := NodeSource, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current
	// make root the first node
	p.ast.nodes = append(p.ast.nodes, Node{})

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.next()

	for {
		if p.atEOF {
			break
		}

		index := p.parseFunctionDecl()
		p.scratch = append(p.scratch, anyIndex(index))
	}
	lhs, rhs = p.addScratchToExtra(scratch_top)

	p.ast.nodes[0] = NodeConstructor[tag](tokenIdx, lhs, rhs)
}

func (p *parser) parseFunctionDecl() NodeIndex {
	tag, tokenIdx, extra := NodeFunctionDecl, tokenIndexInvalid, NodeIndexInvalid
	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	tokenIdx = p.current
	p.expect(TokenKeyword, "fn")
	name := p.parseIdentifier()

	signature := p.parseSignature()
	var block NodeIndex = NodeIndexUndefined
	if p.matchToken(TokenPunctuation, "{") {
		block = p.parseBlock()
	}
	p.expect(TokenTerminator, "")

	p.scratch = append(p.scratch, anyIndex(signature))
	p.scratch = append(p.scratch, anyIndex(block))
	extra, _ = p.addScratchToExtra(scratch_top)

	return p.addNode(NodeConstructor[tag](tokenIdx, name, extra))
}

func (p *parser) parseSignature() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeSignature, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current
	rhs = NodeIndexUndefined

	p.expect(TokenPunctuation, "(")
	if p.matchToken(TokenPunctuation, ")") {
		p.next()
		lhs = p.addNode(NodeConstructor[NodeIdentifierList](
			p.current, NodeIndexUndefined, NodeIndexUndefined))
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	lhs = p.parseIdentifierList()
	p.expect(TokenPunctuation, ")")

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBlock() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeBlock, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.expect(TokenPunctuation, "{")
	if p.matchToken(TokenPunctuation, "}") {
		lhs, rhs = NodeIndexUndefined, NodeIndexUndefined
		p.next()
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	for !p.matchToken(TokenPunctuation, "}") {
		i := p.parseStatement()
		if i != NodeIndexInvalid {
			p.scratch = append(p.scratch, anyIndex(i))
		}
		p.expect(TokenTerminator, "")
	}

	p.expect(TokenPunctuation, "}")
	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseStatement() NodeIndex {
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
		s := p.src.Lexeme(p.current)
		_ = s
		p.rollback()
		return p.parseAssignment()
	}
}

func (p *parser) parseConstDecl() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeConstDecl, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	p.expect(TokenKeyword, "const")
	lhs = p.parseIdentifierList()
	p.expect(TokenPunctuation, "=")
	rhs = p.parseExpressionList()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseAssignment() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeAssignment, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	lhs = p.parseExpressionList()
	p.expect(TokenPunctuation, "=")
	rhs = p.parseExpressionList()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpressionList() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeExpressionList, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, anyIndex(p.parseExpression()))
	for {
		if p.matchToken(TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, anyIndex(p.parseExpression()))
		} else {
			break
		}
	}

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpression() NodeIndex {
	return p.parseBinaryExpr(precLowest + 1)
}

func (p *parser) parseBinaryExpr(precedence int) NodeIndex {
	tokenIdx, lhs, rhs := tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid

	lhs = p.parseUnaryExpr()
	for {
		op := p.src.Lexeme(p.current)
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

func (p *parser) parseUnaryExpr() NodeIndex {
	tokenIdx, lhs, rhs := tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	tag := unaryTag(p.src.Lexeme(p.current))
	if tag == nodeUndefined {
		return p.parsePrimaryExpr()
	}
	p.next()
	lhs = p.parseUnaryExpr()
	rhs = NodeIndexUndefined
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parsePrimaryExpr() NodeIndex {
	tokenIdx, lhs, rhs := tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
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
			p.expect(TokenPunctuation, ")")
			return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
		}
		p.next()
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
	return lhs
}

func (p *parser) parseOperand() NodeIndex {
	if p.matchTag(TokenIdentifier) {
		return p.parseIdentifier()
	}
	if p.isLiteral() {
		return p.parseLiteral()
	}
	p.expect(TokenPunctuation, "(")
	i := p.parseExpression()
	p.expect(TokenPunctuation, ")")
	return i
}

func (p *parser) parseIdentifierList() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeIdentifierList, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, anyIndex(p.parseIdentifier()))
	for {
		if p.matchToken(TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, anyIndex(p.parseIdentifier()))
		} else {
			break
		}
	}

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseIdentifier() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeIdentifier, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
	tokenIdx = p.current

	lhs, rhs = NodeIndexUndefined, NodeIndexUndefined
	p.expect(TokenIdentifier, "")
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseLiteral() NodeIndex {
	tag, tokenIdx, lhs, rhs := NodeIntLiteral, tokenIndexInvalid, NodeIndexInvalid, NodeIndexInvalid
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
