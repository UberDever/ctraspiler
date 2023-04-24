package parser

import (
	antlr_parser "some/antlr"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"golang.org/x/exp/utf8string"
)

const (
	TokenEOF          = -1
	TokenUndefined    = 0
	TokenKeyword      = antlr_parser.SomeKEYWORD
	TokenIdentifier   = antlr_parser.SomeIDENTIFIER
	TokenBinaryOp     = antlr_parser.SomeBINARY_OP
	TokenUnaryOp      = antlr_parser.SomeUNARY_OP
	TokenOtherOp      = antlr_parser.SomeOTHER_OP
	TokenIntLit       = antlr_parser.SomeINT_LIT
	TokenFloatLit     = antlr_parser.SomeFLOAT_LIT
	TokenImaginaryLit = antlr_parser.SomeIMAGINARY_LIT
	TokenRuneLit      = antlr_parser.SomeRUNE_LIT
	TokenLittleUValue = antlr_parser.SomeLITTLE_U_VALUE
	TokenBigUValue    = antlr_parser.SomeBIG_U_VALUE
	TokenStringLit    = antlr_parser.SomeSTRING_LIT
	TokenWS           = antlr_parser.SomeWS
	TokenTerminator   = antlr_parser.SomeTERMINATOR
	TokenLineComment  = antlr_parser.SomeLINE_COMMENT
)

const (
	NodeSource = iota
	NodeIntLiteral
	NodeFloatLiteral
	NodeStringLiteral
)

type Tag = int
type Index = int

type Token struct {
	tag   Tag
	start uint
	end   uint
	line  int
	col   int
}

func tokenize(source []byte) []Token {
	is := antlr.NewInputStream(string(source))
	lexer := antlr_parser.NewSome(is)

	antlrTokens := lexer.GetAllTokens()
	tokens := make([]Token, 0, len(antlrTokens))
	for i := range antlrTokens {
		t := antlrTokens[i]
		if t.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}
		tokens = append(tokens, Token{
			tag:   t.GetTokenType(),
			start: uint(t.GetStart()),
			end:   uint(t.GetStop()),
			line:  t.GetLine(),
			col:   t.GetColumn(),
		})
	}
	tokens = append(tokens, Token{tag: TokenEOF})

	return tokens
}

type Parser struct {
	ast     *AST
	current int
	wasNL   bool
}


func Parse(source []byte, tokens []Token) AST {
	ast := AST{
		source: string(source),
		tokens: tokens,
	}

	p := Parser{
		ast: &ast,
	}

	p.parseSource()

	return ast
}

func (p *Parser) parseSource() {
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

func (p *Parser) parseLiteral() int {
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

func (p *Parser) advance() {
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

type Node struct {
	tag      Tag
	tokenIdx Index
	lhs, rhs Index
}

type Source struct {
	start Index
	end Index
}


type AST struct {
	source string
	tokens []Token
	nodes  []Node
	extra  []Index
}

func (ast *AST) GetNodeString(n *Node) string {
	t := ast.tokens[n.tokenIdx]
	switch n.tag {
	case NodeSource:
		return "Source"
	case NodeIntLiteral:
		fallthrough
	case NodeFloatLiteral:
		fallthrough
	case NodeStringLiteral:
		return utf8string.NewString(ast.source).Slice(int(t.start), int(t.end)+1)
	}
	return "Undefined"

}

func (ast *AST) Traverse(f func(*Node)) {
	root := -1
	for i := range ast.nodes {
		n := ast.nodes[i]
		if n.tag == NodeSource {
			root = i
		}
	}
	if root == -1 {
		panic("Root was not found")
	}

	ast.traverseNodes(f, root)
}

func (ast *AST) traverseNodes(f func(*Node), current int) {
	n := ast.nodes[current]
	f(&n)
	switch n.tag {
	case NodeSource:
		{
			for i := n.lhs; i < n.rhs; i++ {
				c_i := ast.extra[i]
				ast.traverseNodes(f, c_i)
			}
		}
	}
}
