package ast

import (
	s "some/syntax"
	u "some/util"
)

const (
	precLowest  = iota
	precHighest = 7
)

func binaryPrecedenceAndTag(src *s.Source, i s.TokenID) (int, NodeTag) {
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
	return precLowest, NodeIDUndefined
}

func unaryTag(src *s.Source, i s.TokenID) NodeTag {
	lexeme := src.Lexeme(i)
	switch lexeme {
	case "+":
		return NodeUnaryPlus
	case "-":
		return NodeUnaryMinus
	case "!":
		return NodeNot
	}
	return NodeIDUndefined
}

type parser struct {
	ast     *AST
	src     *s.Source
	handler *u.ErrorHandler

	current   s.TokenID
	line, col int
	scratch   []int

	saved s.TokenID
	atEOF bool
}

func NewParser(handler *u.ErrorHandler) parser {
	return parser{
		handler: handler,
		current: -1,
		line:    0,
		col:     0,
		scratch: make([]int, 0, 64),
		saved:   s.TokenIDInvalid,
		atEOF:   false,
	}

}

func (p *parser) Parse(src *s.Source) AST {
	ast := NewAST(src)
	p.ast = &ast
	p.src = src
	p.parseSource()
	return ast
}

func (p *parser) next() {
	for {
		p.current++
		c := p.src.Token(p.current)
		p.line = c.Line
		p.col = c.Col
		p.atEOF = c.Tag == s.TokenEOF

		if c.Tag == s.TokenLineComment {
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
	c := p.src.Token(p.current)
	p.line = c.Line
	p.col = c.Col
}

func (p *parser) matchTag(tag s.TokenTag) bool {
	if p.atEOF {
		return false
	}

	c := p.src.Token(p.current)
	return c.Tag == tag
}

func (p *parser) matchToken(tag s.TokenTag, lexeme string) bool {
	if p.atEOF {
		return false
	}

	current := p.src.Token(p.current)
	return current.Tag == tag && p.src.Lexeme(p.current) == lexeme
}

// NOTE: zero value of string is "" and because this is valid in my case (I don't pass that value
// from variables only from literal strings) i can use it as Optional<string>
// but optionals really is missing...
// NOTE: this should be much more complicated with respect to error reporting
func (p *parser) expect(tag s.TokenTag, lexeme string) (ok bool) {
	if p.atEOF {
		return
	}

	if lexeme == "" {
		ok = p.matchTag(tag)
	} else {
		ok = p.matchToken(tag, lexeme)
	}

	if !ok {
		c := p.src.Token(p.current)
		expected := p.src.TraceToken(tag, lexeme, int(s.TokenEOF), int(s.TokenEOF))
		got := p.src.TraceToken(c.Tag, p.src.Lexeme(p.current), c.Line, c.Col)
		p.handler.Add(u.NewError(
			u.Parser, u.EP_ExpectedToken, c.Line, c.Col, p.src.Filename(), expected, got,
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

func (p *parser) addScratchToExtra(scratch_top int) (NodeID, NodeID) {
	slice := p.scratch[scratch_top:]
	return p.ast.AddExtra(slice)
}

func (p *parser) isLiteral() bool {
	return p.matchTag(s.TokenIntLit) ||
		p.matchTag(s.TokenFloatLit) ||
		p.matchTag(s.TokenStringLit)
}

// NOTE: AST could be built with explicit parents in nodes, this could simplify
// some analysis phases (maybe?) but I won't bother right now
func (p *parser) parseSource() {
	tag, tokenIdx, lhs, rhs := NodeSource, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	// make root the first node
	p.ast.AddNode(Node{})

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.next()
	tokenIdx = p.current

	for {
		if p.atEOF {
			break
		}

		index := p.parseFunctionDecl()
		p.scratch = append(p.scratch, int(index))
	}
	lhs, rhs = p.addScratchToExtra(scratch_top)

	p.ast.SetNode(0, NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseFunctionDecl() NodeID {
	tag, tokenIdx, extra := NodeFunctionDecl, s.TokenIDInvalid, NodeIDInvalid
	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	tokenIdx = p.current
	ok := p.expect(s.TokenKeyword, "fn")
	if !ok {
		return NodeIDInvalid
	}
	name := p.parseIdentifier()

	signature := p.parseSignature()
	var block NodeID = NodeIDUndefined
	if p.matchToken(s.TokenPunctuation, "{") {
		block = p.parseBlock()
	}
	ok = p.expect(s.TokenTerminator, "")
	if !ok {
		return NodeIDInvalid
	}

	p.scratch = append(p.scratch, int(signature))
	p.scratch = append(p.scratch, int(block))
	extra, _ = p.addScratchToExtra(scratch_top)

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, name, extra))
}

func (p *parser) parseSignature() NodeID {
	tag, tokenIdx, lhs, rhs := NodeSignature, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current
	rhs = NodeIDUndefined

	ok := p.expect(s.TokenPunctuation, "(")
	if !ok {
		return NodeIDInvalid
	}
	if p.matchToken(s.TokenPunctuation, ")") {
		p.next()
		lhs = p.ast.AddNode(NodeConstructor[NodeIdentifierList](
			p.current, NodeIDUndefined, NodeIDUndefined))
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	lhs = p.parseIdentifierList()
	ok = p.expect(s.TokenPunctuation, ")")
	if !ok {
		return NodeIDInvalid
	}

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBlock() NodeID {
	tag, tokenIdx, lhs, rhs := NodeBlock, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	ok := p.expect(s.TokenPunctuation, "{")
	if !ok {
		return NodeIDInvalid
	}
	if p.matchToken(s.TokenPunctuation, "}") {
		lhs, rhs = NodeIDUndefined, NodeIDUndefined
		p.next()
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	for !p.matchToken(s.TokenPunctuation, "}") {
		i := p.parseStatement()
		if i != NodeIDInvalid {
			p.scratch = append(p.scratch, int(i))
		}
		ok = p.expect(s.TokenTerminator, "")
		if !ok {
			return NodeIDInvalid
		}
	}

	ok = p.expect(s.TokenPunctuation, "}")
	if !ok {
		return NodeIDInvalid
	}
	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseStatement() NodeID {
	if p.matchTag(s.TokenTerminator) {
		// skip empty statement
		return NodeIDInvalid
	} else if p.matchToken(s.TokenKeyword, "const") {
		return p.parseConstDecl()
	} else if p.matchToken(s.TokenKeyword, "return") {
		return p.parseReturnStmt()
	} else if p.matchToken(s.TokenPunctuation, "{") {
		return p.parseBlock()
	} else {
		// NOTE: need to rollback here, because I don't bother
		// to find all terminals that start an expression
		// if grammar lets you do that this is very convenient and the right thing(TM)

		p.save()
		i := p.parseExpression()
		if p.matchTag(s.TokenTerminator) {
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
	tag, tokenIdx, lhs, rhs := NodeReturnStmt, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	ok := p.expect(s.TokenKeyword, "return")
	if !ok {
		return NodeIDInvalid
	}
	lhs = p.parseExpressionList()
	rhs = NodeIDUndefined

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseConstDecl() NodeID {
	tag, tokenIdx, lhs, rhs := NodeConstDecl, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	ok := p.expect(s.TokenKeyword, "const")
	if !ok {
		return NodeIDInvalid
	}
	lhs = p.parseIdentifierList()
	ok = p.expect(s.TokenPunctuation, "=")
	if !ok {
		return NodeIDInvalid
	}
	rhs = p.parseExpressionList()

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseAssignment() NodeID {
	tag, tokenIdx, lhs, rhs := NodeAssignment, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	lhs = p.parseExpressionList()
	ok := p.expect(s.TokenPunctuation, "=")
	if !ok {
		return NodeIDInvalid
	}
	rhs = p.parseExpressionList()

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpressionList() NodeID {
	tag, tokenIdx, lhs, rhs := NodeExpressionList, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, int(p.parseExpression()))
	for {
		if p.matchToken(s.TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, int(p.parseExpression()))
		} else {
			break
		}
	}

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpression() NodeID {
	tag, tokenIdx, lhs, rhs := NodeExpression, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current
	lhs = p.parseBinaryExpr(precLowest + 1)
	rhs = NodeIDUndefined

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBinaryExpr(precedence int) NodeID {
	tokenIdx, lhs, rhs := s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid

	lhs = p.parseUnaryExpr()
	for {
		opPrec, tag := binaryPrecedenceAndTag(p.src, p.current)
		if opPrec < precedence {
			return lhs
		}
		tokenIdx = p.current
		p.next()

		rhs = p.parseBinaryExpr(opPrec + 1)
		lhs = p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
}

func (p *parser) parseUnaryExpr() NodeID {
	tokenIdx, lhs, rhs := s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	tag := unaryTag(p.src, p.current)
	if tag == NodeIDUndefined {
		return p.parsePrimaryExpr()
	}
	p.next()
	lhs = p.parseUnaryExpr()
	rhs = NodeIDUndefined
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parsePrimaryExpr() NodeID {
	tokenIdx, lhs, rhs := s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	lhs = p.parseOperand()
	if p.matchToken(s.TokenPunctuation, ".") {
		p.next()
		tag := NodeSelector
		rhs = p.parseIdentifier()
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	} else if p.matchToken(s.TokenPunctuation, "(") {
		p.next()
		tag := NodeCall
		rhs = NodeIDUndefined
		if !p.matchToken(s.TokenPunctuation, ")") {
			rhs = p.parseExpressionList()
			ok := p.expect(s.TokenPunctuation, ")")
			if !ok {
				return NodeIDInvalid
			}
			return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
		}
		p.next()
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
	return lhs
}

func (p *parser) parseOperand() NodeID {
	if p.matchTag(s.TokenIdentifier) {
		return p.parseIdentifier()
	}
	if p.isLiteral() {
		return p.parseLiteral()
	}
	ok := p.expect(s.TokenPunctuation, "(")
	if !ok {
		return NodeIDInvalid
	}
	i := p.parseExpression()
	ok = p.expect(s.TokenPunctuation, ")")
	if !ok {
		return NodeIDInvalid
	}
	return i
}

func (p *parser) parseIdentifierList() NodeID {
	tag, tokenIdx, lhs, rhs := NodeIdentifierList, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, int(p.parseIdentifier()))
	for {
		if p.matchToken(s.TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, int(p.parseIdentifier()))
		} else {
			break
		}
	}

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseIdentifier() NodeID {
	tag, tokenIdx, lhs, rhs := NodeIdentifier, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current

	lhs, rhs = NodeIDUndefined, NodeIDUndefined
	ok := p.expect(s.TokenIdentifier, "")
	if !ok {
		return NodeIDInvalid
	}
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseLiteral() NodeID {
	tag, tokenIdx, lhs, rhs := NodeIntLiteral, s.TokenIDInvalid, NodeIDInvalid, NodeIDInvalid
	tokenIdx = p.current
	lhs, rhs = NodeIDUndefined, NodeIDUndefined

	if p.matchTag(s.TokenIntLit) {
		tag = NodeIntLiteral
	} else if p.matchTag(s.TokenFloatLit) {
		tag = NodeFloatLiteral
	} else if p.matchTag(s.TokenStringLit) {
		tag = NodeStringLiteral
	} else {
		panic("parseLiteral unimplemented")
	}

	p.next()
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}
