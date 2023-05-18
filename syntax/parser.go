package syntax

import "some/util"

const (
	precLowest  = iota
	precHighest = 7
)

func binaryPrecedenceAndTag(src *Source, i TokenID) (int, NodeTag) {
	lexeme := src.Lexeme(i)
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

func unaryTag(src *Source, i TokenID) NodeTag {
	lexeme := src.Lexeme(i)
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

	current   TokenID
	line, col int
	scratch   []anyIndex

	saved TokenID
	atEOF bool
}

func NewParser(handler *util.ErrorHandler) parser {
	return parser{
		handler: handler,
		current: -1,
		line:    0,
		col:     0,
		scratch: make([]anyIndex, 0, 64),
		saved:   tokenIDInvalid,
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

func (p *parser) addNode(n Node) NodeID {
	p.ast.nodes = append(p.ast.nodes, n)
	return NodeID(len(p.ast.nodes) - 1)
}

func (p *parser) matchTag(tag TokenTag) bool {
	if p.atEOF {
		return false
	}

	c := p.src.token(p.current)
	return c.tag == tag
}

func (p *parser) matchToken(tag TokenTag, lexeme string) bool {
	if p.atEOF {
		return false
	}

	current := p.src.token(p.current)
	return current.tag == tag && p.src.Lexeme(p.current) == lexeme
}

// NOTE: zero value of string is "" and because this is valid in my case (I don't pass that value
// from variables only from literal strings) i can use it as Optional<string>
// but optionals really is missing...
// NOTE: this should be much more complicated with respect to error reporting
func (p *parser) expect(tag TokenTag, lexeme string) (ok bool) {
	if p.atEOF {
		return
	}

	if lexeme == "" {
		ok = p.matchTag(tag)
	} else {
		ok = p.matchToken(tag, lexeme)
	}

	if !ok {
		c := p.src.token(p.current)
		expected := p.src.traceToken(tag, lexeme, int(TokenEOF), int(TokenEOF))
		got := p.src.traceToken(c.tag, p.src.Lexeme(p.current), c.line, c.col)
		p.handler.Add(util.NewError(
			util.Parser, util.EP_ExpectedToken, c.line, c.col, p.src.filename, expected, got,
		))
		p.atEOF = true
		return
	}

	p.next()
	return
}

func (p *parser) restoreScratch(old_size int) {
	p.scratch = p.scratch[:old_size]
}

func (p *parser) addScratchToExtra(scratch_top int) (start NodeID, end NodeID) {
	slice := p.scratch[scratch_top:]
	p.ast.extra = append(p.ast.extra, slice...)
	start = NodeID(len(p.ast.extra) - len(slice))
	end = NodeID(len(p.ast.extra))
	return
}

func (p *parser) isLiteral() bool {
	return p.matchTag(TokenIntLit) ||
		p.matchTag(TokenFloatLit) ||
		p.matchTag(TokenStringLit)
}

// NOTE: AST could be built with explicit parents in nodes, this could simplify
// some analysis phases (maybe?) but I won't bother right now
func (p *parser) parseSource() {
	tag, tokenIdx, lhs, rhs := NodeSource, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	// make root the first node
	p.ast.nodes = append(p.ast.nodes, Node{})

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.next()
	tokenIdx = p.current

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

func (p *parser) parseFunctionDecl() NodeID {
	tag, tokenIdx, extra := NodeFunctionDecl, tokenIDInvalid, NodeIDInvalid
	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	tokenIdx = p.current
	ok := p.expect(TokenKeyword, "fn")
	if !ok {
		return NodeIDInvalid
	}
	name := p.parseIdentifier()

	signature := p.parseSignature()
	var block NodeID = NodeIDUndefined
	if p.matchToken(TokenPunctuation, "{") {
		block = p.parseBlock()
	}
	ok = p.expect(TokenTerminator, "")
	if !ok {
		return NodeIDInvalid
	}

	p.scratch = append(p.scratch, anyIndex(signature))
	p.scratch = append(p.scratch, anyIndex(block))
	extra, _ = p.addScratchToExtra(scratch_top)

	return p.addNode(NodeConstructor[tag](tokenIdx, name, extra))
}

func (p *parser) parseSignature() NodeID {
	tag, tokenIdx, lhs, rhs := NodeSignature, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current
	rhs = NodeIDUndefined

	ok := p.expect(TokenPunctuation, "(")
	if !ok {
		return NodeIDInvalid
	}
	if p.matchToken(TokenPunctuation, ")") {
		p.next()
		lhs = p.addNode(NodeConstructor[NodeIdentifierList](
			p.current, NodeIDUndefined, NodeIDUndefined))
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	lhs = p.parseIdentifierList()
	ok = p.expect(TokenPunctuation, ")")
	if !ok {
		return NodeIDInvalid
	}

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBlock() NodeID {
	tag, tokenIdx, lhs, rhs := NodeBlock, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	ok := p.expect(TokenPunctuation, "{")
	if !ok {
		return NodeIDInvalid
	}
	if p.matchToken(TokenPunctuation, "}") {
		lhs, rhs = NodeIDUndefined, NodeIDUndefined
		p.next()
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	for !p.matchToken(TokenPunctuation, "}") {
		i := p.parseStatement()
		if i != NodeIDInvalid {
			p.scratch = append(p.scratch, anyIndex(i))
		}
		ok = p.expect(TokenTerminator, "")
		if !ok {
			return NodeIDInvalid
		}
	}

	ok = p.expect(TokenPunctuation, "}")
	if !ok {
		return NodeIDInvalid
	}
	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseStatement() NodeID {
	if p.matchTag(TokenTerminator) {
		// skip empty statement
		return NodeIDInvalid
	} else if p.matchToken(TokenKeyword, "const") {
		return p.parseConstDecl()
	} else if p.matchToken(TokenKeyword, "return") {
		return p.parseReturnStmt()
	} else if p.matchToken(TokenPunctuation, "{") {
		return p.parseBlock()
	} else {
		// NOTE: need to rollback here, because I don't bother
		// to find all terminals that start an expression
		// if grammar lets you do that this is very convenient and the right thing(TM)

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

func (p *parser) parseReturnStmt() NodeID {
	tag, tokenIdx, lhs, rhs := NodeReturnStmt, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	ok := p.expect(TokenKeyword, "return")
	if !ok {
		return NodeIDInvalid
	}
	lhs = p.parseExpressionList()
	rhs = NodeIDUndefined

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseConstDecl() NodeID {
	tag, tokenIdx, lhs, rhs := NodeConstDecl, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	ok := p.expect(TokenKeyword, "const")
	if !ok {
		return NodeIDInvalid
	}
	lhs = p.parseIdentifierList()
	ok = p.expect(TokenPunctuation, "=")
	if !ok {
		return NodeIDInvalid
	}
	rhs = p.parseExpressionList()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseAssignment() NodeID {
	tag, tokenIdx, lhs, rhs := NodeAssignment, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	lhs = p.parseExpressionList()
	ok := p.expect(TokenPunctuation, "=")
	if !ok {
		return NodeIDInvalid
	}
	rhs = p.parseExpressionList()

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpressionList() NodeID {
	tag, tokenIdx, lhs, rhs := NodeExpressionList, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
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

func (p *parser) parseExpression() NodeID {
	tag, tokenIdx, lhs, rhs := NodeExpression, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current
	lhs = p.parseBinaryExpr(precLowest + 1)
	rhs = NodeIDUndefined

	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBinaryExpr(precedence int) NodeID {
	tokenIdx, lhs, rhs := tokenIDInvalid, NodeIDInvalid, NodeIDInvalid

	lhs = p.parseUnaryExpr()
	for {
		opPrec, tag := binaryPrecedenceAndTag(p.src, p.current)
		if opPrec < precedence {
			return lhs
		}
		tokenIdx = p.current
		p.next()

		rhs = p.parseBinaryExpr(opPrec + 1)
		lhs = p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
}

func (p *parser) parseUnaryExpr() NodeID {
	tokenIdx, lhs, rhs := tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	tag := unaryTag(p.src, p.current)
	if tag == nodeUndefined {
		return p.parsePrimaryExpr()
	}
	p.next()
	lhs = p.parseUnaryExpr()
	rhs = NodeIDUndefined
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parsePrimaryExpr() NodeID {
	tokenIdx, lhs, rhs := tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
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
		rhs = NodeIDUndefined
		if !p.matchToken(TokenPunctuation, ")") {
			rhs = p.parseExpressionList()
			ok := p.expect(TokenPunctuation, ")")
			if !ok {
				return NodeIDInvalid
			}
			return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
		}
		p.next()
		return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
	return lhs
}

func (p *parser) parseOperand() NodeID {
	if p.matchTag(TokenIdentifier) {
		return p.parseIdentifier()
	}
	if p.isLiteral() {
		return p.parseLiteral()
	}
	ok := p.expect(TokenPunctuation, "(")
	if !ok {
		return NodeIDInvalid
	}
	i := p.parseExpression()
	ok = p.expect(TokenPunctuation, ")")
	if !ok {
		return NodeIDInvalid
	}
	return i
}

func (p *parser) parseIdentifierList() NodeID {
	tag, tokenIdx, lhs, rhs := NodeIdentifierList, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
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

func (p *parser) parseIdentifier() NodeID {
	tag, tokenIdx, lhs, rhs := NodeIdentifier, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	lhs, rhs = NodeIDUndefined, NodeIDUndefined
	ok := p.expect(TokenIdentifier, "")
	if !ok {
		return NodeIDInvalid
	}
	return p.addNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseLiteral() NodeID {
	tag, tokenIdx, lhs, rhs := NodeIntLiteral, tokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current
	lhs, rhs = NodeIDUndefined, NodeIDUndefined

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
