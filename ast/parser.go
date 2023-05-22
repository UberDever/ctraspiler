package ast

import (
	ID "some/domain"
	s "some/syntax"
	u "some/util"
)

const (
	precLowest  = iota
	precHighest = 7
)

func binaryPrecedenceAndTag(src *s.Source, i ID.Token) (int, NodeTag) {
	lexeme := src.Lexeme(i)
	switch lexeme {
	case "||":
		return 1, ID.NodeOr
	case "&&":
		return 2, ID.NodeAnd
	case "==":
		return 3, ID.NodeEquals
	case "!=":
		return 3, ID.NodeNotEquals
	case ">":
		return 3, ID.NodeGreaterThan
	case "<":
		return 3, ID.NodeLessThan
	case ">=":
		return 3, ID.NodeGreaterThanEquals
	case "<=":
		return 3, ID.NodeLessThanEquals
	case "+":
		return 4, ID.NodeBinaryPlus
	case "-":
		return 4, ID.NodeBinaryMinus
	case "*":
		return 5, ID.NodeMultiply
	case "/":
		return 5, ID.NodeDivide
	}
	return precLowest, ID.NodeUndefined
}

func unaryTag(src *s.Source, i ID.Token) NodeTag {
	lexeme := src.Lexeme(i)
	switch lexeme {
	case "+":
		return ID.NodeUnaryPlus
	case "-":
		return ID.NodeUnaryMinus
	case "!":
		return ID.NodeNot
	}
	return ID.NodeUndefined
}

type parser struct {
	ast     *AST
	src     *s.Source
	handler *u.ErrorHandler

	current   ID.Token
	line, col int
	scratch   []int

	saved ID.Token
	atEOF bool
}

func NewParser(handler *u.ErrorHandler) parser {
	return parser{
		handler: handler,
		current: -1,
		line:    0,
		col:     0,
		scratch: make([]int, 0, 64),
		saved:   ID.TokenInvalid,
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
		p.atEOF = c.Tag == ID.TokenEOF

		if c.Tag == ID.TokenLineComment {
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

func (p *parser) matchTag(tag ID.Token) bool {
	if p.atEOF {
		return false
	}

	c := p.src.Token(p.current)
	return c.Tag == tag
}

func (p *parser) matchToken(tag ID.Token, lexeme string) bool {
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
func (p *parser) expect(tag ID.Token, lexeme string) (ok bool) {
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
		expected := p.src.TraceToken(tag, lexeme, int(ID.TokenEOF), int(ID.TokenEOF))
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

func (p *parser) addScratchToExtra(scratch_top int) (ID.Node, ID.Node) {
	slice := p.scratch[scratch_top:]
	return p.ast.AddExtra(slice)
}

func (p *parser) isLiteral() bool {
	return p.matchTag(ID.TokenIntLit) ||
		p.matchTag(ID.TokenFloatLit) ||
		p.matchTag(ID.TokenStringLit) ||
		p.matchTag(ID.TokenBoolLit)
}

// NOTE: AST could be built with explicit parents in nodes, this could simplify
// some analysis phases (maybe?) but I won't bother right now
func (p *parser) parseSource() {
	tag, tokenIdx, lhs, rhs := ID.NodeSource, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
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

func (p *parser) parseFunctionDecl() ID.Node {
	tag, tokenIdx, extra := ID.NodeFunctionDecl, ID.TokenInvalid, ID.NodeInvalid
	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	tokenIdx = p.current
	ok := p.expect(ID.TokenKeyword, "fn")
	if !ok {
		return ID.NodeInvalid
	}
	name := p.parseIdentifier()

	signature := p.parseSignature()
	var block ID.Node = ID.NodeUndefined
	if p.matchToken(ID.TokenPunctuation, "{") {
		block = p.parseBlock()
	}
	ok = p.expect(ID.TokenTerminator, "")
	if !ok {
		return ID.NodeInvalid
	}

	p.scratch = append(p.scratch, int(signature))
	p.scratch = append(p.scratch, int(block))
	extra, _ = p.addScratchToExtra(scratch_top)

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, name, extra))
}

func (p *parser) parseSignature() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeSignature, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current
	rhs = ID.NodeUndefined

	ok := p.expect(ID.TokenPunctuation, "(")
	if !ok {
		return ID.NodeInvalid
	}
	if p.matchToken(ID.TokenPunctuation, ")") {
		p.next()
		lhs = p.ast.AddNode(NodeConstructor[ID.NodeIdentifierList](
			p.current, ID.NodeUndefined, ID.NodeUndefined))
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	lhs = p.parseIdentifierList()
	ok = p.expect(ID.TokenPunctuation, ")")
	if !ok {
		return ID.NodeInvalid
	}

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBlock() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeBlock, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	ok := p.expect(ID.TokenPunctuation, "{")
	if !ok {
		return ID.NodeInvalid
	}
	if p.matchToken(ID.TokenPunctuation, "}") {
		lhs, rhs = ID.NodeUndefined, ID.NodeUndefined
		p.next()
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}

	for !p.matchToken(ID.TokenPunctuation, "}") {
		i := p.parseStatement()
		if i != ID.NodeInvalid {
			p.scratch = append(p.scratch, int(i))
		}
		ok = p.expect(ID.TokenTerminator, "")
		if !ok {
			return ID.NodeInvalid
		}
	}

	ok = p.expect(ID.TokenPunctuation, "}")
	if !ok {
		return ID.NodeInvalid
	}
	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseStatement() ID.Node {
	if p.matchTag(ID.TokenTerminator) {
		// skip empty statement
		return ID.NodeInvalid
	} else if p.matchToken(ID.TokenKeyword, "const") {
		return p.parseConstDecl()
	} else if p.matchToken(ID.TokenKeyword, "var") {
		return p.parseVarDecl()
	} else if p.matchToken(ID.TokenKeyword, "return") {
		return p.parseReturnStmt()
	} else if p.matchToken(ID.TokenPunctuation, "{") {
		return p.parseBlock()
	} else {
		// NOTE: need to rollback here, because I don't bother
		// to find all terminals that start an expression
		// if grammar lets you do that this is very convenient and the right thing(TM)

		p.save()
		i := p.parseExpression()
		if p.matchTag(ID.TokenTerminator) {
			// expression statement
			return i
		}
		s := p.src.Lexeme(p.current)
		_ = s
		p.rollback()
		return p.parseAssignment()
	}
}

func (p *parser) parseReturnStmt() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeReturnStmt, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	ok := p.expect(ID.TokenKeyword, "return")
	if !ok {
		return ID.NodeInvalid
	}
	lhs = p.parseExpressionList()
	rhs = ID.NodeUndefined

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseConstDecl() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeConstDecl, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	ok := p.expect(ID.TokenKeyword, "const")
	if !ok {
		return ID.NodeInvalid
	}
	lhs = p.parseIdentifierList()
	ok = p.expect(ID.TokenPunctuation, "=")
	if !ok {
		return ID.NodeInvalid
	}
	rhs = p.parseExpressionList()

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseVarDecl() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeVarDecl, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	ok := p.expect(ID.TokenKeyword, "var")
	if !ok {
		return ID.NodeInvalid
	}
	lhs = p.parseIdentifierList()
	ok = p.expect(ID.TokenPunctuation, "=")
	if !ok {
		return ID.NodeInvalid
	}
	rhs = p.parseExpressionList()

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseAssignment() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeAssignment, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	lhs = p.parseExpressionList()
	ok := p.expect(ID.TokenPunctuation, "=")
	if !ok {
		return ID.NodeInvalid
	}
	rhs = p.parseExpressionList()

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpressionList() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeExpressionList, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, int(p.parseExpression()))
	for {
		if p.matchToken(ID.TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, int(p.parseExpression()))
		} else {
			break
		}
	}

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseExpression() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeExpression, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current
	lhs = p.parseBinaryExpr(precLowest + 1)
	rhs = ID.NodeUndefined

	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseBinaryExpr(precedence int) ID.Node {
	tokenIdx, lhs, rhs := ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid

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

func (p *parser) parseUnaryExpr() ID.Node {
	tokenIdx, lhs, rhs := ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	tag := unaryTag(p.src, p.current)
	if tag == ID.NodeUndefined {
		return p.parsePrimaryExpr()
	}
	p.next()
	lhs = p.parseUnaryExpr()
	rhs = ID.NodeUndefined
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parsePrimaryExpr() ID.Node {
	tokenIdx, lhs, rhs := ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	lhs = p.parseOperand()
	if p.matchToken(ID.TokenPunctuation, ".") {
		p.next()
		tag := ID.NodeSelector
		rhs = p.parseIdentifier()
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	} else if p.matchToken(ID.TokenPunctuation, "(") {
		p.next()
		tag := ID.NodeCall
		rhs = ID.NodeUndefined
		if !p.matchToken(ID.TokenPunctuation, ")") {
			rhs = p.parseExpressionList()
			ok := p.expect(ID.TokenPunctuation, ")")
			if !ok {
				return ID.NodeInvalid
			}
			return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
		}
		p.next()
		return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
	}
	return lhs
}

func (p *parser) parseOperand() ID.Node {
	if p.matchTag(ID.TokenIdentifier) {
		return p.parseIdentifier()
	}
	if p.isLiteral() {
		return p.parseLiteral()
	}
	ok := p.expect(ID.TokenPunctuation, "(")
	if !ok {
		return ID.NodeInvalid
	}
	i := p.parseExpression()
	ok = p.expect(ID.TokenPunctuation, ")")
	if !ok {
		return ID.NodeInvalid
	}
	return i
}

func (p *parser) parseIdentifierList() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeIdentifierList, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	scratch_top := len(p.scratch)
	defer p.restoreScratch(scratch_top)

	p.scratch = append(p.scratch, int(p.parseIdentifier()))
	for {
		if p.matchToken(ID.TokenPunctuation, ",") {
			p.next()
			p.scratch = append(p.scratch, int(p.parseIdentifier()))
		} else {
			break
		}
	}

	lhs, rhs = p.addScratchToExtra(scratch_top)
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseIdentifier() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeIdentifier, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current

	lhs, rhs = ID.NodeUndefined, ID.NodeUndefined
	ok := p.expect(ID.TokenIdentifier, "")
	if !ok {
		return ID.NodeInvalid
	}
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}

func (p *parser) parseLiteral() ID.Node {
	tag, tokenIdx, lhs, rhs := ID.NodeIntLiteral, ID.TokenInvalid, ID.NodeInvalid, ID.NodeInvalid
	tokenIdx = p.current
	lhs, rhs = ID.NodeUndefined, ID.NodeUndefined

	if p.matchTag(ID.TokenIntLit) {
		tag = ID.NodeIntLiteral
	} else if p.matchTag(ID.TokenFloatLit) {
		tag = ID.NodeFloatLiteral
	} else if p.matchTag(ID.TokenStringLit) {
		tag = ID.NodeStringLiteral
	} else if p.matchTag(ID.TokenBoolLit) {
		tag = ID.NodeBoolLiteral
	} else {
		panic("parseLiteral unimplemented")
	}

	p.next()
	return p.ast.AddNode(NodeConstructor[tag](tokenIdx, lhs, rhs))
}
